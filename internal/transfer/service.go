package transfer

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/local/telecom/internal/clients"
	"github.com/local/telecom/internal/devices"
	"github.com/local/telecom/internal/diagrams"
	"github.com/local/telecom/internal/documents"
	telecomexport "github.com/local/telecom/internal/export"
)

type DeviceBundle struct {
	Device    devices.Device    `json:"device"`
	Addresses []devices.Address `json:"addresses"`
}
type DiagramBundle struct {
	Diagram diagrams.Diagram `json:"diagram"`
	Nodes   []diagrams.Node  `json:"nodes"`
	Edges   []diagrams.Edge  `json:"edges"`
}

func ExportProject(ctx context.Context, db *sql.DB, projectID string) (telecomexport.Document, error) {
	clientRepository := clients.NewRepository(db)
	deviceRepository := devices.NewRepository(db)
	diagramRepository := diagrams.New(db)
	documentRepository := documents.New(db)
	project, err := clientRepository.GetProject(ctx, projectID)
	if err != nil {
		return telecomexport.Document{}, err
	}
	client, err := clientRepository.GetClient(ctx, project.ClientID)
	if err != nil {
		return telecomexport.Document{}, err
	}
	result := telecomexport.NewDocument()
	result.Client = client
	result.Project = project
	inventory, err := deviceRepository.List(ctx, projectID, "")
	if err != nil {
		return result, err
	}
	for _, device := range inventory {
		addresses, loadErr := deviceRepository.Addresses(ctx, device.ID)
		if loadErr != nil {
			return result, loadErr
		}
		result.Devices = append(result.Devices, DeviceBundle{Device: device, Addresses: addresses})
	}
	diagramList, err := diagramRepository.List(ctx, projectID)
	if err != nil {
		return result, err
	}
	for _, diagram := range diagramList {
		saved, nodes, edges, loadErr := diagramRepository.Graph(ctx, diagram.ID)
		if loadErr != nil {
			return result, loadErr
		}
		result.Diagrams = append(result.Diagrams, DiagramBundle{Diagram: saved, Nodes: nodes, Edges: edges})
	}
	document, err := documentRepository.Get(ctx, projectID)
	if err == nil {
		result.Documents = append(result.Documents, document)
	} else if err != sql.ErrNoRows {
		return result, err
	}
	return result, nil
}

func ImportProject(ctx context.Context, db *sql.DB, document telecomexport.Document) error {
	if err := telecomexport.Validate(document); err != nil {
		return err
	}
	client, err := decode[clients.Client](document.Client)
	if err != nil {
		return fmt.Errorf("cliente inválido: %w", err)
	}
	project, err := decode[clients.Project](document.Project)
	if err != nil {
		return fmt.Errorf("projeto inválido: %w", err)
	}
	if err = client.Validate(); err != nil {
		return err
	}
	if err = project.Validate(); err != nil {
		return err
	}
	if client.ID == "" || project.ID == "" || project.ClientID != client.ID {
		return fmt.Errorf("IDs ou relacionamento cliente/projeto inválidos")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `INSERT INTO clients(id,name,legal_name,document,phone,email,contact_name,address,city,state,postal_code,description,notes)VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`, client.ID, client.Name, client.LegalName, client.Document, client.Phone, client.Email, client.ContactName, client.Address, client.City, client.State, client.PostalCode, client.Description, client.Notes); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO projects(id,client_id,name,description,location,address,local_contact,notes)VALUES(?,?,?,?,?,?,?,?)`, project.ID, project.ClientID, project.Name, project.Description, project.Location, project.Address, project.LocalContact, project.Notes); err != nil {
		return err
	}
	for index, raw := range document.Devices {
		bundle, decodeErr := decode[DeviceBundle](raw)
		if decodeErr != nil {
			return fmt.Errorf("equipamento %d inválido", index+1)
		}
		device := bundle.Device
		if device.ID == "" || device.ProjectID != project.ID {
			return fmt.Errorf("relacionamento de equipamento inválido")
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO devices(id,project_id,name,description,category_id,manufacturer,model,serial_number,hostname,vlan,location,room,rack,rack_position,operating_system,firmware,admin_url,status,notes)VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, device.ID, device.ProjectID, device.Name, device.Description, device.CategoryID, device.Manufacturer, device.Model, device.SerialNumber, device.Hostname, device.VLAN, device.Location, device.Room, device.Rack, device.RackPosition, device.OperatingSystem, device.Firmware, device.AdminURL, device.Status, device.Notes); err != nil {
			return err
		}
		for _, address := range bundle.Addresses {
			if address.DeviceID != device.ID {
				return fmt.Errorf("relacionamento de endereço inválido")
			}
			if err = address.Validate(); err != nil {
				return err
			}
			if _, err = tx.ExecContext(ctx, "INSERT INTO device_addresses(id,device_id,type,address,interface,vlan,is_primary)VALUES(?,?,?,?,?,?,?)", address.ID, address.DeviceID, address.Type, address.Address, address.Interface, address.VLAN, address.Primary); err != nil {
				return err
			}
		}
	}
	for index, raw := range document.Diagrams {
		bundle, decodeErr := decode[DiagramBundle](raw)
		if decodeErr != nil {
			return fmt.Errorf("diagrama %d inválido", index+1)
		}
		if bundle.Diagram.ProjectID != project.ID {
			return fmt.Errorf("relacionamento de diagrama inválido")
		}
		if _, err = tx.ExecContext(ctx, "INSERT INTO diagrams(id,project_id,name,description)VALUES(?,?,?,?)", bundle.Diagram.ID, bundle.Diagram.ProjectID, bundle.Diagram.Name, bundle.Diagram.Description); err != nil {
			return err
		}
		for _, node := range bundle.Nodes {
			if node.DiagramID != bundle.Diagram.ID {
				return fmt.Errorf("relacionamento de node inválido")
			}
			if _, err = tx.ExecContext(ctx, "INSERT INTO diagram_nodes(id,diagram_id,device_id,label,x,y,width,height,style_json)VALUES(?,?,?,?,?,?,?,?,?)", node.ID, node.DiagramID, nullIfEmpty(node.DeviceID), node.Label, node.X, node.Y, node.Width, node.Height, node.StyleJSON); err != nil {
				return err
			}
		}
		for _, edge := range bundle.Edges {
			if edge.DiagramID != bundle.Diagram.ID {
				return fmt.Errorf("relacionamento de ligação inválido")
			}
			if _, err = tx.ExecContext(ctx, `INSERT INTO diagram_edges(id,diagram_id,source_node_id,target_node_id,name,description,type,source_interface,target_interface,speed,vlan,technology,color,line_style,notes)VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, edge.ID, edge.DiagramID, edge.SourceNodeID, edge.TargetNodeID, edge.Name, edge.Description, edge.Type, edge.SourceInterface, edge.TargetInterface, edge.Speed, edge.VLAN, edge.Technology, edge.Color, edge.LineStyle, edge.Notes); err != nil {
				return err
			}
		}
	}
	for _, raw := range document.Documents {
		saved, decodeErr := decode[documents.Document](raw)
		if decodeErr != nil {
			return decodeErr
		}
		if saved.ProjectID != project.ID {
			return fmt.Errorf("relacionamento de documentação inválido")
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO network_documents(id,project_id,title,responsible,general_description,internet_wan,lan,vlans,wifi,cctv,telephony,servers,racks,cabling,fiber,links,power,procedures,notes,free_text)VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, saved.ID, saved.ProjectID, saved.Title, saved.Responsible, saved.GeneralDescription, saved.InternetWAN, saved.LAN, saved.VLANs, saved.WiFi, saved.CCTV, saved.Telephony, saved.Servers, saved.Racks, saved.Cabling, saved.Fiber, saved.Links, saved.Power, saved.Procedures, saved.Notes, saved.FreeText); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func decode[T any](value any) (T, error) {
	var result T
	data, err := json.Marshal(value)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(data, &result)
	return result, err
}
func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}

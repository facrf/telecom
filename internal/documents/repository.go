package documents

import (
	"context"
	"database/sql"
)

type Document struct {
	ID                 string `json:"id"`
	ProjectID          string `json:"projectId"`
	Title              string `json:"title"`
	Responsible        string `json:"responsible"`
	GeneralDescription string `json:"generalDescription"`
	InternetWAN        string `json:"internetWan"`
	LAN                string `json:"lan"`
	VLANs              string `json:"vlans"`
	WiFi               string `json:"wifi"`
	CCTV               string `json:"cctv"`
	Telephony          string `json:"telephony"`
	Servers            string `json:"servers"`
	Racks              string `json:"racks"`
	Cabling            string `json:"cabling"`
	Fiber              string `json:"fiber"`
	Links              string `json:"links"`
	Power              string `json:"power"`
	Procedures         string `json:"procedures"`
	Notes              string `json:"notes"`
	FreeText           string `json:"freeText"`
}
type Repository struct{ db *sql.DB }

func New(db *sql.DB) *Repository { return &Repository{db} }
func (r *Repository) Get(ctx context.Context, projectID string) (Document, error) {
	var d Document
	e := r.db.QueryRowContext(ctx, `SELECT id,project_id,title,responsible,general_description,internet_wan,lan,vlans,wifi,cctv,telephony,servers,racks,cabling,fiber,links,power,procedures,notes,free_text FROM network_documents WHERE project_id=?`, projectID).Scan(&d.ID, &d.ProjectID, &d.Title, &d.Responsible, &d.GeneralDescription, &d.InternetWAN, &d.LAN, &d.VLANs, &d.WiFi, &d.CCTV, &d.Telephony, &d.Servers, &d.Racks, &d.Cabling, &d.Fiber, &d.Links, &d.Power, &d.Procedures, &d.Notes, &d.FreeText)
	return d, e
}
func (r *Repository) Save(ctx context.Context, d Document) (Document, error) {
	_, e := r.db.ExecContext(ctx, `INSERT INTO network_documents(id,project_id,title,responsible,general_description,internet_wan,lan,vlans,wifi,cctv,telephony,servers,racks,cabling,fiber,links,power,procedures,notes,free_text)VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(project_id) DO UPDATE SET title=excluded.title,responsible=excluded.responsible,general_description=excluded.general_description,internet_wan=excluded.internet_wan,lan=excluded.lan,vlans=excluded.vlans,wifi=excluded.wifi,cctv=excluded.cctv,telephony=excluded.telephony,servers=excluded.servers,racks=excluded.racks,cabling=excluded.cabling,fiber=excluded.fiber,links=excluded.links,power=excluded.power,procedures=excluded.procedures,notes=excluded.notes,free_text=excluded.free_text,updated_at=CURRENT_TIMESTAMP`, d.ID, d.ProjectID, d.Title, d.Responsible, d.GeneralDescription, d.InternetWAN, d.LAN, d.VLANs, d.WiFi, d.CCTV, d.Telephony, d.Servers, d.Racks, d.Cabling, d.Fiber, d.Links, d.Power, d.Procedures, d.Notes, d.FreeText)
	if e != nil {
		return d, e
	}
	return r.Get(ctx, d.ProjectID)
}

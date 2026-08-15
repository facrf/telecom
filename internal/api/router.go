package api

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/local/telecom/internal/attachments"
	"github.com/local/telecom/internal/audit"
	"github.com/local/telecom/internal/backup"
	"github.com/local/telecom/internal/clients"
	"github.com/local/telecom/internal/dashboard"
	"github.com/local/telecom/internal/devices"
	"github.com/local/telecom/internal/diagrams"
	"github.com/local/telecom/internal/documents"
	"github.com/local/telecom/internal/fingerprint"
	"github.com/local/telecom/internal/importer"
	"github.com/local/telecom/internal/reports"
	"github.com/local/telecom/internal/scanner"
	"github.com/local/telecom/internal/search"
	"github.com/local/telecom/internal/settings"
	"github.com/local/telecom/internal/tags"
	"github.com/local/telecom/internal/technicalvisits"
	"github.com/local/telecom/internal/transfer"
	"github.com/local/telecom/internal/web"
)

func NewRouter(db *sql.DB, logger *slog.Logger, workers int, dataDir string, onRestore func()) http.Handler {
	r := chi.NewRouter()
	r.Use(recovery(logger), requestLog(logger), securityHeaders)
	r.Get("/health", health(db))
	r.Mount("/api/v1", apiRouter(clients.NewRepository(db), workers, dataDir, onRestore))
	r.Mount("/", web.Handler())
	return r
}

func apiRouter(repository *clients.Repository, workers int, dataDir string, onRestore func()) http.Handler {
	deviceRepository := devices.NewRepository(repositoryDB(repository))
	documentRepository := documents.New(repositoryDB(repository))
	diagramRepository := diagrams.New(repositoryDB(repository))
	dashboardService := dashboard.New(repositoryDB(repository))
	auditRepository := audit.New(repositoryDB(repository))
	searchService := search.New(repositoryDB(repository))
	settingsRepository := settings.New(repositoryDB(repository))
	tagRepository := tags.New(repositoryDB(repository))
	attachmentStore := attachments.NewStore(dataDir, 20<<20)
	technicalVisitService := technicalvisits.NewService(technicalvisits.NewRepository(repositoryDB(repository)))
	backupService := backup.New(repositoryDB(repository))
	scanManager := scanner.NewManager(repositoryDB(repository), workers, 750*time.Millisecond)
	portManager := scanner.NewPortManager(repositoryDB(repository), workers, 750*time.Millisecond)
	r := chi.NewRouter()
	r.Use(auditMutations(auditRepository))
	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"version": "v1"})
	})
	clientAPI := clientHandlers{repository: repository}
	visitAPI := technicalVisitHandlers{service: technicalVisitService}
	r.Route("/clients", func(r chi.Router) {
		r.Get("/", clientAPI.list)
		r.Post("/", clientAPI.create)
		r.Route("/{clientID}", func(r chi.Router) {
			r.Get("/", clientAPI.get)
			r.Put("/", clientAPI.update)
			r.Delete("/", clientAPI.delete)
			r.Get("/projects", clientAPI.listProjects)
			r.Post("/projects", clientAPI.createProject)
			r.Get("/technical-visits", visitAPI.list)
		})
	})
	r.Route("/projects", func(r chi.Router) {
		r.Get("/", clientAPI.allProjects)
		r.Route("/{projectID}", func(r chi.Router) {
			r.Get("/", clientAPI.getProject)
			r.Put("/", clientAPI.updateProject)
			r.Delete("/", clientAPI.deleteProject)
			r.Get("/technical-visits", visitAPI.list)
		})
	})
	r.Route("/technical-visits", func(r chi.Router) {
		r.Get("/", visitAPI.list)
		r.Post("/", visitAPI.create)
		r.Route("/{visitID}", func(r chi.Router) {
			r.Get("/", visitAPI.get)
			r.Put("/", visitAPI.update)
			r.Delete("/", visitAPI.delete)
			visitAPI.detailRoutes(r)
		})
	})
	r.Get("/device-categories", func(w http.ResponseWriter, r *http.Request) {
		items, err := deviceRepository.Categories(r.Context())
		if err != nil {
			serverError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	})
	r.Route("/devices", func(r chi.Router) {
		h := deviceHandlers{deviceRepository}
		r.Get("/", h.list)
		r.Post("/", h.create)
		r.Route("/{deviceID}", func(r chi.Router) {
			r.Get("/", h.get)
			r.Put("/", h.update)
			r.Delete("/", h.delete)
			r.Get("/addresses", h.addresses)
			r.Post("/addresses", h.addAddress)
			r.Delete("/addresses/{addressID}", h.deleteAddress)
		})
	})
	r.Route("/scans", func(r chi.Router) {
		h := scanHandlers{scanManager}
		r.Get("/", h.list)
		r.Post("/", h.start)
		r.Route("/{scanID}", func(r chi.Router) {
			r.Get("/", h.get)
			r.Get("/hosts", h.hosts)
			r.Get("/changes", h.changes)
			r.Post("/cancel", h.cancel)
			r.Get("/events", h.events)
		})
	})
	r.Post("/devices/{deviceID}/port-scans", func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Mode  string `json:"mode"`
			Ports string `json:"ports"`
		}
		if !decodeJSON(w, r, &payload) {
			return
		}
		id := newID()
		if e := portManager.Start(r.Context(), id, chi.URLParam(r, "deviceID"), payload.Mode, payload.Ports); e != nil {
			writeError(w, 422, "INVALID_PORT_SCAN", e.Error())
			return
		}
		writeJSON(w, 201, map[string]string{"id": id, "status": "running"})
	})
	r.Get("/port-scans/{scanID}/ports", func(w http.ResponseWriter, r *http.Request) {
		items, e := portManager.Results(r.Context(), chi.URLParam(r, "scanID"))
		if e != nil {
			serverError(w, e)
			return
		}
		writeJSON(w, 200, map[string]any{"items": items})
	})
	r.Get("/port-scans", func(w http.ResponseWriter, r *http.Request) {
		items, e := portManager.List(r.Context(), r.URL.Query().Get("device_id"))
		if e != nil {
			serverError(w, e)
			return
		}
		writeJSON(w, 200, map[string]any{"items": items})
	})
	r.Route("/projects/{projectID}/document", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			document, err := documentRepository.Get(r.Context(), chi.URLParam(r, "projectID"))
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, 404, "NOT_FOUND", "Documentação não encontrada")
				return
			}
			if err != nil {
				serverError(w, err)
				return
			}
			writeJSON(w, 200, document)
		})
		r.Put("/", func(w http.ResponseWriter, r *http.Request) {
			var document documents.Document
			if !decodeJSON(w, r, &document) {
				return
			}
			document.ProjectID = chi.URLParam(r, "projectID")
			if document.ID == "" {
				document.ID = newID()
			}
			saved, err := documentRepository.Save(r.Context(), document)
			if err != nil {
				serverError(w, err)
				return
			}
			writeJSON(w, 200, saved)
		})
	})
	r.Route("/projects/{projectID}/diagrams", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			items, e := diagramRepository.List(r.Context(), chi.URLParam(r, "projectID"))
			if e != nil {
				serverError(w, e)
				return
			}
			writeJSON(w, 200, map[string]any{"items": items})
		})
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			var d diagrams.Diagram
			if !decodeJSON(w, r, &d) {
				return
			}
			d.ID = newID()
			d.ProjectID = chi.URLParam(r, "projectID")
			if d.Name == "" {
				writeError(w, 422, "VALIDATION_ERROR", "Nome é obrigatório")
				return
			}
			saved, e := diagramRepository.Save(r.Context(), d)
			if e != nil {
				serverError(w, e)
				return
			}
			writeJSON(w, 201, saved)
		})
	})
	r.Route("/diagrams/{diagramID}", func(r chi.Router) {
		r.Put("/", func(w http.ResponseWriter, r *http.Request) {
			var value diagrams.Diagram
			if !decodeJSON(w, r, &value) {
				return
			}
			value.ID = chi.URLParam(r, "diagramID")
			saved, e := diagramRepository.UpdateDiagram(r.Context(), value)
			if errors.Is(e, diagrams.ErrNotFound) {
				writeError(w, 404, "NOT_FOUND", "Diagrama não encontrado")
				return
			}
			if e != nil {
				serverError(w, e)
				return
			}
			writeJSON(w, 200, saved)
		})
		r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
			e := diagramRepository.DeleteDiagram(r.Context(), chi.URLParam(r, "diagramID"))
			if errors.Is(e, diagrams.ErrNotFound) {
				writeError(w, 404, "NOT_FOUND", "Diagrama não encontrado")
				return
			}
			if e != nil {
				serverError(w, e)
				return
			}
			w.WriteHeader(204)
		})
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			diagram, nodes, edges, e := diagramRepository.Graph(r.Context(), chi.URLParam(r, "diagramID"))
			if errors.Is(e, diagrams.ErrNotFound) {
				writeError(w, 404, "NOT_FOUND", "Diagrama não encontrado")
				return
			}
			if e != nil {
				serverError(w, e)
				return
			}
			writeJSON(w, 200, map[string]any{"diagram": diagram, "nodes": nodes, "edges": edges})
		})
		r.Post("/nodes", func(w http.ResponseWriter, r *http.Request) {
			var node diagrams.Node
			if !decodeJSON(w, r, &node) {
				return
			}
			node.ID = newID()
			node.DiagramID = chi.URLParam(r, "diagramID")
			saved, e := diagramRepository.SaveNode(r.Context(), node)
			if e != nil {
				serverError(w, e)
				return
			}
			writeJSON(w, 201, saved)
		})
		r.Put("/nodes/{nodeID}", func(w http.ResponseWriter, r *http.Request) {
			var value diagrams.Node
			if !decodeJSON(w, r, &value) {
				return
			}
			value.ID = chi.URLParam(r, "nodeID")
			value.DiagramID = chi.URLParam(r, "diagramID")
			saved, e := diagramRepository.UpdateNode(r.Context(), value)
			if errors.Is(e, diagrams.ErrNotFound) {
				writeError(w, 404, "NOT_FOUND", "Node não encontrado")
				return
			}
			if e != nil {
				serverError(w, e)
				return
			}
			writeJSON(w, 200, saved)
		})
		r.Delete("/nodes/{nodeID}", func(w http.ResponseWriter, r *http.Request) {
			e := diagramRepository.DeleteNode(r.Context(), chi.URLParam(r, "nodeID"))
			if errors.Is(e, diagrams.ErrNotFound) {
				writeError(w, 404, "NOT_FOUND", "Node não encontrado")
				return
			}
			if e != nil {
				serverError(w, e)
				return
			}
			w.WriteHeader(204)
		})
		r.Post("/edges", func(w http.ResponseWriter, r *http.Request) {
			var edge diagrams.Edge
			if !decodeJSON(w, r, &edge) {
				return
			}
			edge.ID = newID()
			edge.DiagramID = chi.URLParam(r, "diagramID")
			saved, e := diagramRepository.SaveEdge(r.Context(), edge)
			if e != nil {
				serverError(w, e)
				return
			}
			writeJSON(w, 201, saved)
		})
		r.Put("/edges/{edgeID}", func(w http.ResponseWriter, r *http.Request) {
			var value diagrams.Edge
			if !decodeJSON(w, r, &value) {
				return
			}
			value.ID = chi.URLParam(r, "edgeID")
			value.DiagramID = chi.URLParam(r, "diagramID")
			saved, e := diagramRepository.UpdateEdge(r.Context(), value)
			if errors.Is(e, diagrams.ErrNotFound) {
				writeError(w, 404, "NOT_FOUND", "Ligação não encontrada")
				return
			}
			if e != nil {
				serverError(w, e)
				return
			}
			writeJSON(w, 200, saved)
		})
		r.Delete("/edges/{edgeID}", func(w http.ResponseWriter, r *http.Request) {
			e := diagramRepository.DeleteEdge(r.Context(), chi.URLParam(r, "edgeID"))
			if errors.Is(e, diagrams.ErrNotFound) {
				writeError(w, 404, "NOT_FOUND", "Ligação não encontrada")
				return
			}
			if e != nil {
				serverError(w, e)
				return
			}
			w.WriteHeader(204)
		})
		r.Get("/export.svg", func(w http.ResponseWriter, r *http.Request) {
			_, nodes, edges, e := diagramRepository.Graph(r.Context(), chi.URLParam(r, "diagramID"))
			if errors.Is(e, diagrams.ErrNotFound) {
				writeError(w, 404, "NOT_FOUND", "Diagrama não encontrado")
				return
			}
			if e != nil {
				serverError(w, e)
				return
			}
			svgNodes := make([]reports.Node, 0, len(nodes))
			for _, n := range nodes {
				svgNodes = append(svgNodes, reports.Node{ID: n.ID, Label: n.Label, X: n.X, Y: n.Y, Width: n.Width, Height: n.Height})
			}
			svgEdges := make([]reports.Edge, 0, len(edges))
			for _, edge := range edges {
				svgEdges = append(svgEdges, reports.Edge{Source: edge.SourceNodeID, Target: edge.TargetNodeID, Label: edge.Name, Color: edge.Color})
			}
			w.Header().Set("Content-Type", "image/svg+xml")
			w.Header().Set("Content-Disposition", "attachment; filename=topologia.svg")
			_, _ = w.Write([]byte(reports.DiagramSVG(svgNodes, svgEdges)))
		})
	})
	r.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		summary, e := dashboardService.Summary(r.Context(), r.URL.Query().Get("project_id"))
		if e != nil {
			serverError(w, e)
			return
		}
		writeJSON(w, 200, summary)
	})
	r.Get("/audit-logs", func(w http.ResponseWriter, r *http.Request) {
		entries, e := auditRepository.Recent(r.Context(), 20)
		if e != nil {
			serverError(w, e)
			return
		}
		writeJSON(w, 200, map[string]any{"items": entries})
	})
	r.Get("/search", func(w http.ResponseWriter, r *http.Request) {
		items, e := searchService.Find(r.Context(), r.URL.Query().Get("q"))
		if e != nil {
			serverError(w, e)
			return
		}
		writeJSON(w, 200, map[string]any{"items": items})
	})
	r.Get("/settings", func(w http.ResponseWriter, r *http.Request) {
		values, e := settingsRepository.All(r.Context())
		if e != nil {
			serverError(w, e)
			return
		}
		writeJSON(w, 200, values)
	})
	r.Get("/system/status", func(w http.ResponseWriter, r *http.Request) {
		_, e := exec.LookPath("nmap")
		writeJSON(w, 200, map[string]any{"scannerNative": "available", "nmap": map[bool]string{true: "available", false: "not_installed"}[e == nil]})
	})
	r.Post("/oui/import", func(w http.ResponseWriter, r *http.Request) {
		content, e := io.ReadAll(io.LimitReader(r.Body, 5<<20))
		if e != nil {
			serverError(w, e)
			return
		}
		count, e := fingerprint.ImportOUI(r.Context(), repositoryDB(repository), strings.NewReader(string(content)))
		if e != nil {
			writeError(w, 422, "INVALID_OUI", e.Error())
			return
		}
		writeJSON(w, 200, map[string]int{"imported": count})
	})
	r.Route("/tags", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			items, e := tagRepository.List(r.Context())
			if e != nil {
				serverError(w, e)
				return
			}
			writeJSON(w, 200, map[string]any{"items": items})
		})
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			var tag tags.Tag
			if !decodeJSON(w, r, &tag) {
				return
			}
			tag.ID = newID()
			if tag.Color == "" {
				tag.Color = "#64748b"
			}
			saved, e := tagRepository.Create(r.Context(), tag)
			if e != nil {
				writeError(w, 422, "INVALID_TAG", e.Error())
				return
			}
			writeJSON(w, 201, saved)
		})
		r.Post("/{tagID}/assign/{entityType}/{entityID}", func(w http.ResponseWriter, r *http.Request) {
			if e := tagRepository.Assign(r.Context(), chi.URLParam(r, "entityType"), chi.URLParam(r, "entityID"), chi.URLParam(r, "tagID")); e != nil {
				writeError(w, 422, "INVALID_TAG_ASSIGNMENT", e.Error())
				return
			}
			w.WriteHeader(204)
		})
		r.Get("/{entityType}/{entityID}", func(w http.ResponseWriter, r *http.Request) {
			items, e := tagRepository.Assigned(r.Context(), chi.URLParam(r, "entityType"), chi.URLParam(r, "entityID"))
			if e != nil {
				serverError(w, e)
				return
			}
			writeJSON(w, 200, map[string]any{"items": items})
		})
		r.Delete("/{tagID}/assign/{entityType}/{entityID}", func(w http.ResponseWriter, r *http.Request) {
			if e := tagRepository.Unassign(r.Context(), chi.URLParam(r, "entityType"), chi.URLParam(r, "entityID"), chi.URLParam(r, "tagID")); e != nil {
				serverError(w, e)
				return
			}
			w.WriteHeader(204)
		})
		r.Delete("/{tagID}", func(w http.ResponseWriter, r *http.Request) {
			if e := tagRepository.Delete(r.Context(), chi.URLParam(r, "tagID")); e != nil {
				serverError(w, e)
				return
			}
			w.WriteHeader(204)
		})
	})
	r.Put("/settings/{key}", func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Value string `json:"value"`
		}
		if !decodeJSON(w, r, &payload) {
			return
		}
		if e := settingsRepository.Set(r.Context(), chi.URLParam(r, "key"), payload.Value); e != nil {
			writeError(w, 422, "INVALID_SETTING", e.Error())
			return
		}
		w.WriteHeader(204)
	})
	r.Post("/attachments/{entityType}/{entityID}", func(w http.ResponseWriter, r *http.Request) {
		if e := r.ParseMultipartForm(20 << 20); e != nil {
			writeError(w, 400, "INVALID_UPLOAD", "Upload inválido")
			return
		}
		file, header, e := r.FormFile("file")
		if e != nil {
			writeError(w, 400, "INVALID_UPLOAD", "Campo file é obrigatório")
			return
		}
		defer file.Close()
		content, e := io.ReadAll(io.LimitReader(file, 20<<20+1))
		if e != nil {
			serverError(w, e)
			return
		}
		metadata, hash, e := attachmentStore.Save(chi.URLParam(r, "entityType"), header.Filename, content)
		if e != nil {
			writeError(w, 422, "INVALID_UPLOAD", e.Error())
			return
		}
		id := newID()
		_, e = repositoryDB(repository).ExecContext(r.Context(), "INSERT INTO attachments(id,entity_type,entity_id,original_filename,stored_filename,mime_type,size,sha256,description)VALUES(?,?,?,?,?,?,?,?,?)", id, chi.URLParam(r, "entityType"), chi.URLParam(r, "entityID"), metadata.OriginalFilename, metadata.StoredFilename, metadata.MIMEType, metadata.Size, hash, r.FormValue("description"))
		if e != nil {
			serverError(w, e)
			return
		}
		writeJSON(w, 201, map[string]any{"id": id, "filename": metadata.OriginalFilename, "mimeType": metadata.MIMEType, "size": metadata.Size})
	})
	r.Get("/attachments/{entityType}/{entityID}", func(w http.ResponseWriter, r *http.Request) {
		rows, e := repositoryDB(repository).QueryContext(r.Context(), "SELECT id,original_filename,mime_type,size,description,created_at FROM attachments WHERE entity_type=? AND entity_id=? ORDER BY created_at DESC", chi.URLParam(r, "entityType"), chi.URLParam(r, "entityID"))
		if e != nil {
			serverError(w, e)
			return
		}
		defer rows.Close()
		items := []map[string]any{}
		for rows.Next() {
			var id, name, mime, description, created string
			var size int64
			if e = rows.Scan(&id, &name, &mime, &size, &description, &created); e != nil {
				serverError(w, e)
				return
			}
			items = append(items, map[string]any{"id": id, "filename": name, "mimeType": mime, "size": size, "description": description, "createdAt": created})
		}
		writeJSON(w, 200, map[string]any{"items": items})
	})
	r.Get("/attachments/{attachmentID}/download", func(w http.ResponseWriter, r *http.Request) {
		var entityType, stored, name, mime string
		e := repositoryDB(repository).QueryRowContext(r.Context(), "SELECT entity_type,stored_filename,original_filename,mime_type FROM attachments WHERE id=?", chi.URLParam(r, "attachmentID")).Scan(&entityType, &stored, &name, &mime)
		if errors.Is(e, sql.ErrNoRows) {
			writeError(w, 404, "NOT_FOUND", "Anexo não encontrado")
			return
		}
		if e != nil {
			serverError(w, e)
			return
		}
		if filepath.Base(stored) != stored {
			writeError(w, 500, "INVALID_ATTACHMENT", "Arquivo inválido")
			return
		}
		path := filepath.Join(dataDir, "attachments", entityType, stored)
		w.Header().Set("Content-Type", mime)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", name))
		http.ServeFile(w, r, path)
	})
	r.Delete("/attachments/{attachmentID}", func(w http.ResponseWriter, r *http.Request) {
		var entityType, stored string
		e := repositoryDB(repository).QueryRowContext(r.Context(), "SELECT entity_type,stored_filename FROM attachments WHERE id=?", chi.URLParam(r, "attachmentID")).Scan(&entityType, &stored)
		if errors.Is(e, sql.ErrNoRows) {
			writeError(w, 404, "NOT_FOUND", "Anexo não encontrado")
			return
		}
		if e != nil {
			serverError(w, e)
			return
		}
		if filepath.Base(stored) != stored {
			writeError(w, 500, "INVALID_ATTACHMENT", "Arquivo inválido")
			return
		}
		tx, e := repositoryDB(repository).BeginTx(r.Context(), nil)
		if e != nil {
			serverError(w, e)
			return
		}
		if _, e = tx.ExecContext(r.Context(), "DELETE FROM attachments WHERE id=?", chi.URLParam(r, "attachmentID")); e != nil {
			tx.Rollback()
			serverError(w, e)
			return
		}
		if e = os.Remove(filepath.Join(dataDir, "attachments", entityType, stored)); e != nil && !os.IsNotExist(e) {
			tx.Rollback()
			serverError(w, e)
			return
		}
		if e = tx.Commit(); e != nil {
			serverError(w, e)
			return
		}
		w.WriteHeader(204)
	})
	r.Post("/imports/preview", func(w http.ResponseWriter, r *http.Request) {
		content, e := io.ReadAll(io.LimitReader(r.Body, 10<<20))
		if e != nil {
			serverError(w, e)
			return
		}
		document, e := importer.Validate(content)
		if e != nil {
			writeError(w, 422, "INVALID_IMPORT", e.Error())
			return
		}
		writeJSON(w, 200, map[string]any{"format": document.Format, "schemaVersion": document.SchemaVersion, "devices": len(document.Devices), "diagrams": len(document.Diagrams), "documents": len(document.Documents)})
	})
	r.Post("/imports/apply", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("confirm") != "true" {
			writeError(w, 409, "CONFIRMATION_REQUIRED", "Confirmação explícita é obrigatória")
			return
		}
		content, e := io.ReadAll(io.LimitReader(r.Body, 10<<20))
		if e != nil {
			serverError(w, e)
			return
		}
		document, e := importer.Validate(content)
		if e != nil {
			writeError(w, 422, "INVALID_IMPORT", e.Error())
			return
		}
		if e = transfer.ImportProject(r.Context(), repositoryDB(repository), document); e != nil {
			writeError(w, 422, "IMPORT_FAILED", e.Error())
			return
		}
		writeJSON(w, 201, map[string]string{"status": "imported"})
	})
	r.Get("/exports/projects/{projectID}", func(w http.ResponseWriter, r *http.Request) {
		document, e := transfer.ExportProject(r.Context(), repositoryDB(repository), chi.URLParam(r, "projectID"))
		if errors.Is(e, clients.ErrNotFound) {
			writeError(w, 404, "NOT_FOUND", "Projeto não encontrado")
			return
		}
		if e != nil {
			serverError(w, e)
			return
		}
		w.Header().Set("Content-Disposition", "attachment; filename=telecom-project.json")
		writeJSON(w, 200, document)
	})
	r.Route("/backups", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			directory := filepath.Join(dataDir, "backups")
			entries, e := os.ReadDir(directory)
			if e != nil && !os.IsNotExist(e) {
				serverError(w, e)
				return
			}
			items := []map[string]any{}
			for _, entry := range entries {
				if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".zip") {
					continue
				}
				info, infoErr := entry.Info()
				if infoErr == nil {
					items = append(items, map[string]any{"name": entry.Name(), "size": info.Size(), "createdAt": info.ModTime().UTC()})
				}
			}
			writeJSON(w, 200, map[string]any{"items": items})
		})
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			directory := filepath.Join(dataDir, "backups")
			if e := os.MkdirAll(directory, 0o750); e != nil {
				serverError(w, e)
				return
			}
			name := "telecom-backup-" + time.Now().UTC().Format("20060102T150405Z") + ".zip"
			if e := backupService.Create(r.Context(), filepath.Join(directory, name), filepath.Join(dataDir, "attachments")); e != nil {
				serverError(w, e)
				return
			}
			writeJSON(w, 201, map[string]string{"name": name, "download": "/api/v1/backups/" + name + "/download"})
		})
		r.Get("/{name}/download", func(w http.ResponseWriter, r *http.Request) {
			name := chi.URLParam(r, "name")
			if filepath.Base(name) != name || !strings.HasSuffix(name, ".zip") {
				writeError(w, 400, "INVALID_BACKUP", "Nome inválido")
				return
			}
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", name))
			http.ServeFile(w, r, filepath.Join(dataDir, "backups", name))
		})
		r.Post("/restore", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("confirm") != "true" {
				writeError(w, 409, "CONFIRMATION_REQUIRED", "Confirmação explícita é obrigatória")
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, 2<<30)
			if e := r.ParseMultipartForm(32 << 20); e != nil {
				writeError(w, 400, "INVALID_BACKUP", e.Error())
				return
			}
			file, _, e := r.FormFile("file")
			if e != nil {
				writeError(w, 400, "INVALID_BACKUP", "Campo file é obrigatório")
				return
			}
			defer file.Close()
			inspection, e := backup.QueueRestore(dataDir, file)
			if e != nil {
				writeError(w, 422, "INVALID_BACKUP", e.Error())
				return
			}
			writeJSON(w, 202, map[string]any{"status": "restore_scheduled", "attachments": inspection.AttachmentCount})
			if onRestore != nil {
				go func() { time.Sleep(250 * time.Millisecond); onRestore() }()
			}
		})
	})
	return r
}

type scanHandlers struct{ manager *scanner.Manager }

func (h scanHandlers) list(w http.ResponseWriter, r *http.Request) {
	items, e := h.manager.List(r.Context(), r.URL.Query().Get("project_id"))
	if e != nil {
		serverError(w, e)
		return
	}
	writeJSON(w, 200, map[string]any{"items": items})
}

func (h scanHandlers) start(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		ProjectID string `json:"projectId"`
		Network   string `json:"network"`
	}
	if !decodeJSON(w, r, &payload) {
		return
	}
	if payload.ProjectID == "" || payload.Network == "" {
		writeError(w, 422, "VALIDATION_ERROR", "Projeto e rede são obrigatórios")
		return
	}
	scan, e := h.manager.Start(r.Context(), newID(), payload.ProjectID, payload.Network)
	if e != nil {
		writeError(w, 422, "INVALID_NETWORK", e.Error())
		return
	}
	writeJSON(w, 201, scan)
}
func (h scanHandlers) get(w http.ResponseWriter, r *http.Request) {
	scan, e := h.manager.Get(r.Context(), chi.URLParam(r, "scanID"))
	if errors.Is(e, sql.ErrNoRows) {
		writeError(w, 404, "NOT_FOUND", "Scan não encontrado")
		return
	}
	if e != nil {
		serverError(w, e)
		return
	}
	writeJSON(w, 200, scan)
}
func (h scanHandlers) hosts(w http.ResponseWriter, r *http.Request) {
	hosts, e := h.manager.Hosts(r.Context(), chi.URLParam(r, "scanID"))
	if e != nil {
		serverError(w, e)
		return
	}
	writeJSON(w, 200, map[string]any{"items": hosts})
}
func (h scanHandlers) changes(w http.ResponseWriter, r *http.Request) {
	items, err := h.manager.Events(r.Context(), chi.URLParam(r, "scanID"))
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}
func (h scanHandlers) cancel(w http.ResponseWriter, r *http.Request) {
	if e := h.manager.Cancel(r.Context(), chi.URLParam(r, "scanID")); e != nil {
		writeError(w, 409, "SCAN_NOT_RUNNING", e.Error())
		return
	}
	w.WriteHeader(202)
}
func (h scanHandlers) events(w http.ResponseWriter, r *http.Request) {
	events, unsubscribe, e := h.manager.Subscribe(chi.URLParam(r, "scanID"))
	if e != nil {
		writeError(w, 404, "SCAN_NOT_RUNNING", e.Error())
		return
	}
	defer unsubscribe()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, 500, "SSE_UNAVAILABLE", "Streaming indisponível")
		return
	}
	for {
		select {
		case event, open := <-events:
			if !open {
				return
			}
			encoded, _ := json.Marshal(event)
			_, _ = fmt.Fprintf(w, "event: progress\ndata: %s\n\n", encoded)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

// repositoryDB is temporary composition glue until repositories are grouped by an application service.
func repositoryDB(repository *clients.Repository) *sql.DB { return repository.Database() }

type deviceHandlers struct{ repository *devices.Repository }

func (h deviceHandlers) list(w http.ResponseWriter, r *http.Request) {
	items, e := h.repository.List(r.Context(), r.URL.Query().Get("project_id"), strings.TrimSpace(r.URL.Query().Get("q")))
	if e != nil {
		serverError(w, e)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}
func (h deviceHandlers) get(w http.ResponseWriter, r *http.Request) {
	v, e := h.repository.Get(r.Context(), chi.URLParam(r, "deviceID"))
	if errors.Is(e, devices.ErrNotFound) {
		writeError(w, 404, "NOT_FOUND", "Registro não encontrado")
		return
	}
	if e != nil {
		serverError(w, e)
		return
	}
	writeJSON(w, 200, v)
}
func (h deviceHandlers) create(w http.ResponseWriter, r *http.Request) {
	var v devices.Device
	if !decodeJSON(w, r, &v) {
		return
	}
	v.ID = newID()
	if e := v.Validate(); e != nil {
		writeError(w, 422, "VALIDATION_ERROR", e.Error())
		return
	}
	created, e := h.repository.Create(r.Context(), v)
	if e != nil {
		serverError(w, e)
		return
	}
	writeJSON(w, 201, created)
}
func (h deviceHandlers) update(w http.ResponseWriter, r *http.Request) {
	var v devices.Device
	if !decodeJSON(w, r, &v) {
		return
	}
	v.ID = chi.URLParam(r, "deviceID")
	if e := v.Validate(); e != nil {
		writeError(w, 422, "VALIDATION_ERROR", e.Error())
		return
	}
	updated, e := h.repository.Update(r.Context(), v)
	if errors.Is(e, devices.ErrNotFound) {
		writeError(w, 404, "NOT_FOUND", "Registro não encontrado")
		return
	}
	if e != nil {
		serverError(w, e)
		return
	}
	writeJSON(w, 200, updated)
}
func (h deviceHandlers) delete(w http.ResponseWriter, r *http.Request) {
	e := h.repository.Delete(r.Context(), chi.URLParam(r, "deviceID"))
	if errors.Is(e, devices.ErrNotFound) {
		writeError(w, 404, "NOT_FOUND", "Registro não encontrado")
		return
	}
	if e != nil {
		serverError(w, e)
		return
	}
	w.WriteHeader(204)
}
func (h deviceHandlers) addresses(w http.ResponseWriter, r *http.Request) {
	items, e := h.repository.Addresses(r.Context(), chi.URLParam(r, "deviceID"))
	if e != nil {
		serverError(w, e)
		return
	}
	writeJSON(w, 200, map[string]any{"items": items})
}
func (h deviceHandlers) addAddress(w http.ResponseWriter, r *http.Request) {
	var v devices.Address
	if !decodeJSON(w, r, &v) {
		return
	}
	v.ID = newID()
	v.DeviceID = chi.URLParam(r, "deviceID")
	if e := v.Validate(); e != nil {
		writeError(w, 422, "VALIDATION_ERROR", e.Error())
		return
	}
	if e := h.repository.AddAddress(r.Context(), v); e != nil {
		serverError(w, e)
		return
	}
	writeJSON(w, 201, v)
}
func (h deviceHandlers) deleteAddress(w http.ResponseWriter, r *http.Request) {
	e := h.repository.DeleteAddress(r.Context(), chi.URLParam(r, "addressID"))
	if errors.Is(e, devices.ErrNotFound) {
		writeError(w, 404, "NOT_FOUND", "Registro não encontrado")
		return
	}
	if e != nil {
		serverError(w, e)
		return
	}
	w.WriteHeader(204)
}

type clientHandlers struct {
	repository *clients.Repository
}

func (h clientHandlers) list(w http.ResponseWriter, r *http.Request) {
	items, err := h.repository.ListClients(r.Context(), strings.TrimSpace(r.URL.Query().Get("q")))
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}
func (h clientHandlers) get(w http.ResponseWriter, r *http.Request) {
	item, err := h.repository.GetClient(r.Context(), chi.URLParam(r, "clientID"))
	h.respondClient(w, item, err)
}
func (h clientHandlers) create(w http.ResponseWriter, r *http.Request) {
	var item clients.Client
	if !decodeJSON(w, r, &item) {
		return
	}
	item.ID = newID()
	if err := item.Validate(); err != nil {
		validationError(w, err)
		return
	}
	created, err := h.repository.CreateClient(r.Context(), item)
	h.respondClient(w, created, err)
}
func (h clientHandlers) update(w http.ResponseWriter, r *http.Request) {
	var item clients.Client
	if !decodeJSON(w, r, &item) {
		return
	}
	item.ID = chi.URLParam(r, "clientID")
	if err := item.Validate(); err != nil {
		validationError(w, err)
		return
	}
	updated, err := h.repository.UpdateClient(r.Context(), item)
	h.respondClient(w, updated, err)
}
func (h clientHandlers) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "clientID")
	err := h.repository.DeleteClient(r.Context(), id)
	h.respondDelete(w, err)
}
func (h clientHandlers) listProjects(w http.ResponseWriter, r *http.Request) {
	items, err := h.repository.ListProjects(r.Context(), chi.URLParam(r, "clientID"), strings.TrimSpace(r.URL.Query().Get("q")))
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}
func (h clientHandlers) allProjects(w http.ResponseWriter, r *http.Request) {
	items, err := h.repository.ListProjects(r.Context(), "", strings.TrimSpace(r.URL.Query().Get("q")))
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}
func (h clientHandlers) createProject(w http.ResponseWriter, r *http.Request) {
	var item clients.Project
	if !decodeJSON(w, r, &item) {
		return
	}
	item.ID = newID()
	item.ClientID = chi.URLParam(r, "clientID")
	if err := item.Validate(); err != nil {
		validationError(w, err)
		return
	}
	created, err := h.repository.CreateProject(r.Context(), item)
	h.respondProject(w, created, err)
}
func (h clientHandlers) getProject(w http.ResponseWriter, r *http.Request) {
	item, err := h.repository.GetProject(r.Context(), chi.URLParam(r, "projectID"))
	h.respondProject(w, item, err)
}
func (h clientHandlers) updateProject(w http.ResponseWriter, r *http.Request) {
	var item clients.Project
	if !decodeJSON(w, r, &item) {
		return
	}
	item.ID = chi.URLParam(r, "projectID")
	if err := item.Validate(); err != nil {
		validationError(w, err)
		return
	}
	updated, err := h.repository.UpdateProject(r.Context(), item)
	h.respondProject(w, updated, err)
}
func (h clientHandlers) deleteProject(w http.ResponseWriter, r *http.Request) {
	h.respondDelete(w, h.repository.DeleteProject(r.Context(), chi.URLParam(r, "projectID")))
}
func (h clientHandlers) respondClient(w http.ResponseWriter, item clients.Client, err error) {
	if err != nil {
		respondRepositoryError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}
func (h clientHandlers) respondProject(w http.ResponseWriter, item clients.Project, err error) {
	if err != nil {
		respondRepositoryError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}
func (h clientHandlers) respondDelete(w http.ResponseWriter, err error) {
	if err != nil {
		respondRepositoryError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "JSON inválido")
		return false
	}
	return true
}
func validationError(w http.ResponseWriter, err error) {
	var validation clients.ValidationError
	if errors.As(err, &validation) {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"error": map[string]string{"code": "VALIDATION_ERROR", "message": validation.Message, "field": validation.Field}})
		return
	}
	serverError(w, err)
}
func respondRepositoryError(w http.ResponseWriter, err error) {
	if errors.Is(err, clients.ErrNotFound) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Registro não encontrado")
		return
	}
	serverError(w, err)
}
func serverError(w http.ResponseWriter, err error) {
	slog.Error("request failed", "error", err)
	writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno")
}
func newID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err == nil {
		return hex.EncodeToString(bytes)
	}
	return time.Now().UTC().Format("20060102150405.000000000")
}

func health(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := db.PingContext(r.Context()); err != nil {
			writeError(w, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "Banco de dados indisponível")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}
func recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					logger.Error("panic recovered", "event", "http_panic", "error", recovered)
					writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
func requestLog(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			logger.Info("http request", "event", "http_request", "method", r.Method, "path", r.URL.Path, "duration_ms", time.Since(start).Milliseconds())
		})
	}
}

type statusCapture struct {
	http.ResponseWriter
	status int
}

func (w *statusCapture) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
func (w *statusCapture) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(data)
}

func auditMutations(repository *audit.Repository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}
			capture := &statusCapture{ResponseWriter: w}
			next.ServeHTTP(capture, r)
			status := capture.status
			if status == 0 {
				status = http.StatusOK
			}
			if status >= 400 {
				return
			}
			parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
			entityType := "system"
			entityID := ""
			if len(parts) >= 3 {
				entityType = parts[2]
			}
			if len(parts) >= 4 {
				entityID = parts[3]
			}
			_ = repository.Log(r.Context(), audit.Entry{ID: newID(), Action: strings.ToLower(r.Method) + "_" + entityType, EntityType: entityType, EntityID: entityID, Details: map[string]any{"path": r.URL.Path, "status": status}})
		})
	}
}

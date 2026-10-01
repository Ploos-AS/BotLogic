package main

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/Ploos-AS/BotLogic/internal/logic"
)

type server struct{ e *logic.Engine }
type consultRequest struct{ Source string `json:"source"` }
type queryRequest struct{ Query string `json:"query"` }

func writeJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(v)
}

func (s server) health(w http.ResponseWriter, _ *http.Request) {
    writeJSON(w, http.StatusOK, map[string]any{"ok": true, "service": "botlogic", "version": "0.1.0-m0"})
}

func (s server) consult(w http.ResponseWriter, r *http.Request) {
    var req consultRequest
    d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
    d.DisallowUnknownFields()
    if err := d.Decode(&req); err != nil || req.Source == "" {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
        return
    }
    if err := s.e.Consult(req.Source); err != nil {
        writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
        return
    }
    writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s server) query(w http.ResponseWriter, r *http.Request) {
    var req queryRequest
    d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10))
    d.DisallowUnknownFields()
    if err := d.Decode(&req); err != nil || req.Query == "" {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
        return
    }
    ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
    defer cancel()
    rows, err := s.e.Query(ctx, req.Query)
    if err != nil {
        writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
        return
    }
    writeJSON(w, http.StatusOK, map[string]any{"solutions": rows})
}

func main() {
    addr := os.Getenv("BOTLOGIC_LISTEN")
    if addr == "" { addr = "127.0.0.1:8091" }
    s := server{e: logic.New()}
    mux := http.NewServeMux()
    mux.HandleFunc("GET /healthz", s.health)
    mux.HandleFunc("GET /readyz", s.health)
    mux.HandleFunc("POST /v1/consult", s.consult)
    mux.HandleFunc("POST /v1/query", s.query)
    log.Printf("BotLogic listening on %s", addr)
    log.Fatal(http.ListenAndServe(addr, mux))
}

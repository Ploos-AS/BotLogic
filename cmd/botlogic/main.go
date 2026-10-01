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

type server struct{ store *logic.Store }
type consultRequest struct { Ruleset string `json:"ruleset"`; Source string `json:"source"` }
type queryRequest struct { Ruleset string `json:"ruleset"`; Query string `json:"query"` }
type factRequest struct { Ruleset string `json:"ruleset"`; Fact logic.Fact `json:"fact"`; ExpectedRevision *uint64 `json:"expected_revision,omitempty"` }
type factBatchRequest struct { Ruleset string `json:"ruleset"`; Operations []logic.FactOperation `json:"operations"`; ExpectedRevision *uint64 `json:"expected_revision,omitempty"` }

func writeJSON(w http.ResponseWriter,status int,v any){ w.Header().Set("Content-Type","application/json"); w.WriteHeader(status); _=json.NewEncoder(w).Encode(v) }
func (s server) health(w http.ResponseWriter,_ *http.Request){ writeJSON(w,http.StatusOK,map[string]any{"ok":true,"service":"botlogic","version":"0.1.5-m0.5"}) }
func (s server) rulesets(w http.ResponseWriter,_ *http.Request){ writeJSON(w,http.StatusOK,map[string]any{"rulesets":s.store.Names()}) }

func (s server) consult(w http.ResponseWriter,r *http.Request){
	var req consultRequest; d:=json.NewDecoder(http.MaxBytesReader(w,r.Body,64<<10)); d.DisallowUnknownFields()
	if err:=d.Decode(&req); err!=nil || req.Ruleset=="" || req.Source=="" { writeJSON(w,http.StatusBadRequest,map[string]string{"error":"invalid request"}); return }
	if err:=s.store.Put(req.Ruleset,req.Source); err!=nil { writeJSON(w,http.StatusUnprocessableEntity,map[string]string{"error":err.Error()}); return }
	revision,_:=s.store.Revision(req.Ruleset); writeJSON(w,http.StatusOK,map[string]any{"ok":true,"ruleset":req.Ruleset,"revision":revision})
}
func (s server) query(w http.ResponseWriter,r *http.Request){
	var req queryRequest; d:=json.NewDecoder(http.MaxBytesReader(w,r.Body,16<<10)); d.DisallowUnknownFields()
	if err:=d.Decode(&req); err!=nil || req.Ruleset=="" || req.Query=="" { writeJSON(w,http.StatusBadRequest,map[string]string{"error":"invalid request"}); return }
	ctx,cancel:=context.WithTimeout(r.Context(),2*time.Second); defer cancel()
	rows,revision,err:=s.store.QueryAt(ctx,req.Ruleset,req.Query); if err!=nil { writeJSON(w,http.StatusUnprocessableEntity,map[string]string{"error":err.Error()}); return }
	writeJSON(w,http.StatusOK,map[string]any{"ruleset":req.Ruleset,"revision":revision,"solutions":rows})
}
func (s server) explain(w http.ResponseWriter,r *http.Request){
	var req queryRequest; d:=json.NewDecoder(http.MaxBytesReader(w,r.Body,16<<10)); d.DisallowUnknownFields()
	if err:=d.Decode(&req); err!=nil || req.Ruleset=="" || req.Query=="" { writeJSON(w,http.StatusBadRequest,map[string]string{"error":"invalid request"}); return }
	ctx,cancel:=context.WithTimeout(r.Context(),2*time.Second); defer cancel()
	explanation,revision,err:=s.store.ExplainAt(ctx,req.Ruleset,req.Query); if err!=nil { writeJSON(w,http.StatusUnprocessableEntity,map[string]string{"error":err.Error()}); return }
	writeJSON(w,http.StatusOK,map[string]any{"ruleset":req.Ruleset,"revision":revision,"explanation":explanation})
}

func (s server) facts(w http.ResponseWriter,r *http.Request){
	name:=r.URL.Query().Get("ruleset"); if name=="" { writeJSON(w,http.StatusBadRequest,map[string]string{"error":"ruleset required"}); return }
	facts,err:=s.store.Facts(name); if err!=nil { writeJSON(w,http.StatusUnprocessableEntity,map[string]string{"error":err.Error()}); return }; revision,_:=s.store.Revision(name); writeJSON(w,http.StatusOK,map[string]any{"ruleset":name,"revision":revision,"facts":facts})
}
func (s server) factMutation(w http.ResponseWriter,r *http.Request){
	var req factRequest; d:=json.NewDecoder(http.MaxBytesReader(w,r.Body,16<<10)); d.DisallowUnknownFields(); if err:=d.Decode(&req);err!=nil||req.Ruleset=="" { writeJSON(w,http.StatusBadRequest,map[string]string{"error":"invalid request"});return }
	op:="assert";if r.Method!="POST"{op="retract"}; revision,err:=s.store.ApplyFactsExpected(req.Ruleset,[]logic.FactOperation{{Op:op,Fact:req.Fact}},req.ExpectedRevision); if err!=nil { writeJSON(w,http.StatusUnprocessableEntity,map[string]string{"error":err.Error()});return }; writeJSON(w,http.StatusOK,map[string]any{"ok":true,"ruleset":req.Ruleset,"revision":revision})
}

func (s server) factBatch(w http.ResponseWriter,r *http.Request){
	var req factBatchRequest; d:=json.NewDecoder(http.MaxBytesReader(w,r.Body,256<<10)); d.DisallowUnknownFields(); if err:=d.Decode(&req);err!=nil||req.Ruleset=="" { writeJSON(w,http.StatusBadRequest,map[string]string{"error":"invalid request"});return }
	revision,err:=s.store.ApplyFactsExpected(req.Ruleset,req.Operations,req.ExpectedRevision);if err!=nil { writeJSON(w,http.StatusUnprocessableEntity,map[string]string{"error":err.Error()});return }; writeJSON(w,http.StatusOK,map[string]any{"ok":true,"ruleset":req.Ruleset,"revision":revision,"operations":len(req.Operations)})
}

func main(){
	addr:=os.Getenv("BOTLOGIC_LISTEN"); if addr=="" { addr="127.0.0.1:8091" }
	dir:=os.Getenv("BOTLOGIC_DATA_DIR"); if dir=="" { dir="./data" }
	store,err:=logic.NewStore(dir); if err!=nil { log.Fatal(err) }
	s:=server{store:store}; mux:=http.NewServeMux()
	mux.HandleFunc("GET /healthz",s.health); mux.HandleFunc("GET /readyz",s.health)
	mux.HandleFunc("GET /v1/rulesets",s.rulesets); mux.HandleFunc("GET /v1/facts",s.facts); mux.HandleFunc("POST /v1/facts/batch",s.factBatch); mux.HandleFunc("POST /v1/facts",s.factMutation); mux.HandleFunc("DELETE /v1/facts",s.factMutation); mux.HandleFunc("POST /v1/consult",s.consult); mux.HandleFunc("POST /v1/query",s.query); mux.HandleFunc("POST /v1/explain",s.explain)
	log.Printf("BotLogic listening on %s with %d rulesets",addr,len(store.Names())); log.Fatal(http.ListenAndServe(addr,mux))
}

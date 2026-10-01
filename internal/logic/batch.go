package logic

import (
	"errors"
	"strings"
)

type FactOperation struct {
	Op   string `json:"op"`
	Fact Fact   `json:"fact"`
}

func cloneFacts(src map[string]Fact) map[string]Fact {
	dst:=make(map[string]Fact,len(src)); for k,v:=range src { dst[k]=v }; return dst
}

func emptyPredicateStubs(before,after map[string]Fact) []string {
	type sig struct{ p string; n int }; old:=map[sig]bool{}; now:=map[sig]bool{}
	for _,f:=range before { old[sig{f.Predicate,len(f.Args)}]=true }
	for _,f:=range after { now[sig{f.Predicate,len(f.Args)}]=true }
	var out []string
	for x:=range old { if now[x] { continue }; vars:=make([]string,x.n); for i:=range vars { vars[i]="_" }; p:=x.p; if x.n>0 { p+="("+strings.Join(vars,",")+")" }; out=append(out,p+" :- fail.") }
	return out
}

func (s *Store) ApplyFacts(name string, ops []FactOperation) error { _,err:=s.ApplyFactsExpected(name,ops,nil);return 0,err }

func (s *Store) ApplyFactsExpected(name string, ops []FactOperation, expected *uint64) (uint64,error) {
	if len(ops)==0 { return 0,errors.New("empty fact batch") }
	if len(ops)>256 { return 0,errors.New("too many fact operations") }
	for _,op:=range ops { if op.Op!="assert"&&op.Op!="retract" { return 0,errors.New("invalid fact operation") }; if err:=op.Fact.validate();err!=nil{return err} }

	s.mu.Lock(); defer s.mu.Unlock()
	if _,ok:=s.sets[name];!ok{return 0,errors.New("ruleset not found")}
	currentRevision:=s.revisions[name]; if expected!=nil && *expected!=currentRevision { return 0,revisionConflict(*expected,currentRevision) }
	current,err:=s.readFacts(name);if err!=nil{return 0,err}; next:=cloneFacts(current)
	for _,op:=range ops { if op.Op=="assert" { next[op.Fact.key()]=op.Fact } else { delete(next,op.Fact.key()) } }
	e,err:=s.rebuild(name,next);if err!=nil{return 0,err}
	for _,stub:=range emptyPredicateStubs(current,next) { if err:=e.Consult(stub);err!=nil{return 0,err} }
	if err:=s.writeFacts(name,next);err!=nil{return err}
	s.sets[name]=e
	return nil
}

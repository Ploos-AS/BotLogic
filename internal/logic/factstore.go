package logic

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

func (s *Store) factsPath(name string) string { return filepath.Join(s.dir,name+".facts.json") }

func (s *Store) readFacts(name string) (map[string]Fact,error) {
	out:=map[string]Fact{}; if s.dir=="" { return out,nil }
	b,err:=os.ReadFile(s.factsPath(name)); if errors.Is(err,os.ErrNotExist){return out,nil}; if err!=nil{return nil,err}
	var facts []Fact; if err:=json.Unmarshal(b,&facts); err!=nil{return nil,err}
	for _,f:=range facts { if err:=f.validate(); err!=nil{return nil,err}; out[f.key()]=f }; return out,nil
}

func (s *Store) writeFacts(name string,m map[string]Fact) error {
	if s.dir=="" { return nil }; b,err:=json.MarshalIndent(sortedFacts(m),"","  "); if err!=nil{return err}; b=append(b,'\n')
	tmp,err:=os.CreateTemp(s.dir,"."+name+"-facts-*.tmp"); if err!=nil{return err}; tmpName:=tmp.Name(); defer os.Remove(tmpName)
	if err:=tmp.Chmod(0600);err!=nil{tmp.Close();return err}; if _,err:=tmp.Write(b);err!=nil{tmp.Close();return err}; if err:=tmp.Sync();err!=nil{tmp.Close();return err}; if err:=tmp.Close();err!=nil{return err}
	return os.Rename(tmpName,s.factsPath(name))
}

func (s *Store) rebuild(name string, facts map[string]Fact) (*Engine,error) {
	if s.dir=="" { return nil,errors.New("persistent ruleset source unavailable") }
	src,err:=os.ReadFile(filepath.Join(s.dir,name+".pl")); if err!=nil{return nil,err}; e:=New(); if err:=e.Consult(string(src));err!=nil{return nil,err}
	preds:=map[string]int{}; for _,f:=range sortedFacts(facts){ preds[f.Predicate]=len(f.Args); p,err:=f.prolog();if err!=nil{return nil,err};if err:=e.Consult(p);err!=nil{return nil,err} }; return e,nil
}

func (s *Store) AssertFact(name string,f Fact) error {
	if err:=f.validate();err!=nil{return err}; s.mu.Lock(); defer s.mu.Unlock(); if _,ok:=s.sets[name];!ok{return errors.New("ruleset not found")}
	facts,err:=s.readFacts(name);if err!=nil{return err};facts[f.key()]=f;e,err:=s.rebuild(name,facts);if err!=nil{return err};if err:=s.writeFacts(name,facts);err!=nil{return err};s.sets[name]=e;return nil
}
func (s *Store) RetractFact(name string,f Fact) error {
	if err:=f.validate();err!=nil{return err};s.mu.Lock();defer s.mu.Unlock();if _,ok:=s.sets[name];!ok{return errors.New("ruleset not found")}
	facts,err:=s.readFacts(name);if err!=nil{return err}; oldPred:=f.Predicate; oldArity:=len(f.Args); delete(facts,f.key());e,err:=s.rebuild(name,facts);if err!=nil{return err}; still:=false; for _,x:=range facts { if x.Predicate==oldPred && len(x.Args)==oldArity { still=true; break } }; if !still { vars:=make([]string,oldArity); for i:=range vars { vars[i]="_" }; stub:=oldPred; if oldArity>0 { stub+="("+strings.Join(vars,",")+")" }; stub+=" :- fail."; if err:=e.Consult(stub);err!=nil{return err} };if err:=s.writeFacts(name,facts);err!=nil{return err};s.sets[name]=e;return nil
}
func (s *Store) Facts(name string)([]Fact,error){s.mu.RLock();defer s.mu.RUnlock();if _,ok:=s.sets[name];!ok{return nil,errors.New("ruleset not found")};m,err:=s.readFacts(name);if err!=nil{return nil,err};return sortedFacts(m),nil}


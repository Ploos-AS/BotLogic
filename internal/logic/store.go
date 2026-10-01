package logic

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"sync"
)

var validRuleset = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)

type Store struct {
	mu sync.RWMutex
	dir string
	sets map[string]*Engine
}

func NewStore(dir string) (*Store, error) {
	s := &Store{dir: dir, sets: map[string]*Engine{}}
	if dir == "" { return s, nil }
	if err := os.MkdirAll(dir, 0700); err != nil { return nil, err }
	entries, err := os.ReadDir(dir); if err != nil { return nil, err }
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".pl" { continue }
		name := entry.Name()[:len(entry.Name())-3]
		if !validRuleset.MatchString(name) { continue }
		src, err := os.ReadFile(filepath.Join(dir, entry.Name())); if err != nil { return nil, err }
		e := New(); if err := e.Consult(string(src)); err != nil { return nil, fmt.Errorf("load ruleset %q: %w", name, err) }
		facts, err := s.readFacts(name); if err != nil { return nil, fmt.Errorf("load facts %q: %w", name, err) }
		for _, fact := range sortedFacts(facts) { p, err := fact.prolog(); if err != nil { return nil, err }; if err := e.Consult(p); err != nil { return nil, err } }
		s.sets[name] = e
	}
	return s, nil
}

func (s *Store) Names() []string {
	s.mu.RLock(); defer s.mu.RUnlock()
	out := make([]string,0,len(s.sets)); for name := range s.sets { out=append(out,name) }; sort.Strings(out); return out
}

func (s *Store) Put(name, source string) error {
	if !validRuleset.MatchString(name) { return errors.New("invalid ruleset name") }
	e := New(); if err := e.Consult(source); err != nil { return err }
	if s.dir != "" {
		tmp, err := os.CreateTemp(s.dir, "."+name+"-*.tmp"); if err != nil { return err }
		tmpName:=tmp.Name(); defer os.Remove(tmpName)
		if err:=tmp.Chmod(0600); err!=nil { tmp.Close(); return err }
		if _,err:=tmp.WriteString(source); err!=nil { tmp.Close(); return err }
		if err:=tmp.Sync(); err!=nil { tmp.Close(); return err }
		if err:=tmp.Close(); err!=nil { return err }
		if err:=os.Rename(tmpName,filepath.Join(s.dir,name+".pl")); err!=nil { return err }
	}
	s.mu.Lock(); s.sets[name]=e; s.mu.Unlock(); return nil
}

func (s *Store) Query(ctx context.Context, name, query string) ([]map[string]string,error) {
	s.mu.RLock(); e,ok:=s.sets[name]; s.mu.RUnlock()
	if !ok { return nil, errors.New("ruleset not found") }
	return e.Query(ctx,query)
}

func (s *Store) Explain(ctx context.Context, name, query string) (Explanation,error) {
	s.mu.RLock(); e,ok:=s.sets[name]; s.mu.RUnlock()
	if !ok { return Explanation{}, errors.New("ruleset not found") }
	return e.Explain(ctx,query)
}

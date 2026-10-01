package logic

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var validPredicate = regexp.MustCompile(`^[a-z][a-zA-Z0-9_]{0,63}$`)

type Fact struct {
	Predicate string   `json:"predicate"`
	Args      []string `json:"args"`
}

func (f Fact) validate() error {
	if !validPredicate.MatchString(f.Predicate) { return errors.New("invalid predicate") }
	if len(f.Args)>16 { return errors.New("too many fact arguments") }
	for _,a:=range f.Args { if len(a)>1024 { return errors.New("fact argument too long") } }
	return nil
}

func (f Fact) key() string {
	return fmt.Sprintf("%s/%d\x00%s", f.Predicate, len(f.Args), strings.Join(f.Args,"\x00"))
}

func (f Fact) prolog() (string,error) {
	if err:=f.validate(); err!=nil { return "",err }
	args:=make([]string,len(f.Args))
	for i,a:=range f.Args { args[i]=quoteAtom(a) }
	if len(args)==0 { return f.Predicate+".",nil }
	return fmt.Sprintf("%s(%s).",f.Predicate,strings.Join(args,",")),nil
}

func quoteAtom(s string) string {
	// ISO-style quoted atoms: a single quote inside an atom is represented
	// by two single quotes. Keeping every argument quoted also prevents
	// variables, operators, comments, or clauses from becoming source.
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

func sortedFacts(m map[string]Fact) []Fact {
	out:=make([]Fact,0,len(m)); for _,f:=range m { out=append(out,f) }
	sort.Slice(out,func(i,j int)bool{return out[i].key()<out[j].key()}); return out
}

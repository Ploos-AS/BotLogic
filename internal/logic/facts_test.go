package logic

import (
	"context"
	"testing"
)

func TestStructuredFactsAssertRetractAndPersist(t *testing.T){
	dir:=t.TempDir();s,err:=NewStore(dir);if err!=nil{t.Fatal(err)}
	if err:=s.Put("policy",`may_voice(X) :- trusted(X).`);err!=nil{t.Fatal(err)}
	f:=Fact{Predicate:"trusted",Args:[]string{"alice"}}
	if err:=s.AssertFact("policy",f);err!=nil{t.Fatal(err)}
	rows,err:=s.Query(context.Background(),"policy",`may_voice(alice).`);if err!=nil||len(rows)!=1{t.Fatalf("assert rows=%v err=%v",rows,err)}
	facts,err:=s.Facts("policy");if err!=nil||len(facts)!=1{t.Fatalf("facts=%v err=%v",facts,err)}
	s2,err:=NewStore(dir);if err!=nil{t.Fatal(err)}
	rows,err=s2.Query(context.Background(),"policy",`may_voice(alice).`);if err!=nil||len(rows)!=1{t.Fatalf("restart rows=%v err=%v",rows,err)}
	if err:=s2.RetractFact("policy",f);err!=nil{t.Fatal(err)}
	rows,err=s2.Query(context.Background(),"policy",`may_voice(alice).`);if err!=nil||len(rows)!=0{t.Fatalf("retract rows=%v err=%v",rows,err)}
}
func TestFactArgumentsAreDataNotProlog(t *testing.T){
	s,err:=NewStore(t.TempDir());if err!=nil{t.Fatal(err)};if err:=s.Put("safe",`seen(X) :- payload(X).`);err!=nil{t.Fatal(err)}
	f:=Fact{Predicate:"payload",Args:[]string{"x\"). injected(evil). %"}}
	if err:=s.AssertFact("safe",f);err!=nil{t.Fatal(err)}
	rows,err:=s.Query(context.Background(),"safe",`injected(evil).`);if err==nil&&len(rows)>0{t.Fatal("fact argument escaped into Prolog program")}
}
func TestFactPredicateValidation(t *testing.T){
	s,err:=NewStore(t.TempDir());if err!=nil{t.Fatal(err)};if err:=s.Put("safe",`ok.`);err!=nil{t.Fatal(err)}
	if err:=s.AssertFact("safe",Fact{Predicate:"bad-name",Args:[]string{"x"}});err==nil{t.Fatal("expected invalid predicate")}
}

func TestFactArgumentsUnifyWithQuotedAtoms(t *testing.T) {
	s, err := NewStore(t.TempDir())
	if err != nil { t.Fatal(err) }
	if err := s.Put("policy", `allowed(X) :- grant(X).`); err != nil { t.Fatal(err) }
	if err := s.AssertFact("policy", Fact{Predicate:"grant", Args:[]string{"Alice Nick"}}); err != nil { t.Fatal(err) }
	rows, err := s.Query(context.Background(), "policy", `allowed('Alice Nick').`)
	if err != nil || len(rows) != 1 { t.Fatalf("rows=%v err=%v", rows, err) }
}

func TestFactAtomEscapingCannotInjectClause(t *testing.T) {
	s, err := NewStore(t.TempDir())
	if err != nil { t.Fatal(err) }
	if err := s.Put("safe", `seen(X) :- payload(X).`); err != nil { t.Fatal(err) }
	value := "x'). injected(evil). % '"
	if err := s.AssertFact("safe", Fact{Predicate:"payload", Args:[]string{value}}); err != nil { t.Fatal(err) }
	rows, err := s.Query(context.Background(), "safe", `injected(evil).`)
	if err == nil && len(rows) > 0 { t.Fatal("quoted atom escaped into Prolog program") }
	rows, err = s.Query(context.Background(), "safe", `seen('x''). injected(evil). % ''').`)
	if err != nil || len(rows) != 1 { t.Fatalf("payload did not round-trip safely: rows=%v err=%v", rows, err) }
}

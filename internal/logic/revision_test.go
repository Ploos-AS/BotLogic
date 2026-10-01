package logic

import (
	"context"
	"errors"
	"testing"
)

func TestRevisionIncrementsAndPersists(t *testing.T){
	dir:=t.TempDir();s,err:=NewStore(dir);if err!=nil{t.Fatal(err)};if err:=s.Put("state",`online(X) :- present(X).`);err!=nil{t.Fatal(err)}
	if r,_:=s.Revision("state");r!=0{t.Fatalf("initial revision=%d",r)}
	r,err:=s.ApplyFactsExpected("state",[]FactOperation{{Op:"assert",Fact:Fact{Predicate:"present",Args:[]string{"alice"}}}},nil);if err!=nil||r!=1{t.Fatalf("revision=%d err=%v",r,err)}
	s2,err:=NewStore(dir);if err!=nil{t.Fatal(err)};if r,_:=s2.Revision("state");r!=1{t.Fatalf("reloaded revision=%d",r)}
}
func TestExpectedRevisionConflictDoesNotCommit(t *testing.T){
	s,err:=NewStore(t.TempDir());if err!=nil{t.Fatal(err)};if err:=s.Put("state",`ok.`);err!=nil{t.Fatal(err)}
	expected:=uint64(1);_,err=s.ApplyFactsExpected("state",[]FactOperation{{Op:"assert",Fact:Fact{Predicate:"present",Args:[]string{"alice"}}}},&expected)
	if !errors.Is(err,ErrRevisionConflict){t.Fatalf("err=%v",err)}
	facts,_:=s.Facts("state");if len(facts)!=0{t.Fatalf("partial commit: %+v",facts)};if r,_:=s.Revision("state");r!=0{t.Fatalf("revision=%d",r)}
}
func TestQueryAtReportsSnapshotRevision(t *testing.T){
	s,err:=NewStore(t.TempDir());if err!=nil{t.Fatal(err)};if err:=s.Put("state",`seen(X) :- present(X).`);err!=nil{t.Fatal(err)}
	if err:=s.AssertFact("state",Fact{Predicate:"present",Args:[]string{"alice"}});err!=nil{t.Fatal(err)}
	rows,r,err:=s.QueryAt(context.Background(),"state",`seen('alice').`);if err!=nil||len(rows)!=1||r!=1{t.Fatalf("rows=%v revision=%d err=%v",rows,r,err)}
}

package logic

import (
	"context"
	"testing"
)

func TestFactBatchCommitsAtomically(t *testing.T){
	s,err:=NewStore(t.TempDir());if err!=nil{t.Fatal(err)}
	if err:=s.Put("state",`ready(X) :- online(X), authenticated(X).`);err!=nil{t.Fatal(err)}
	err=s.ApplyFacts("state",[]FactOperation{
		{Op:"assert",Fact:Fact{Predicate:"online",Args:[]string{"alice"}}},
		{Op:"assert",Fact:Fact{Predicate:"authenticated",Args:[]string{"alice"}}},
	});if err!=nil{t.Fatal(err)}
	rows,err:=s.Query(context.Background(),"state",`ready("alice").`);if err!=nil||len(rows)!=1{t.Fatalf("rows=%v err=%v",rows,err)}
}
func TestInvalidFactBatchRollsBack(t *testing.T){
	s,err:=NewStore(t.TempDir());if err!=nil{t.Fatal(err)}
	if err:=s.Put("state",`online(X) :- present(X).`);err!=nil{t.Fatal(err)}
	err=s.ApplyFacts("state",[]FactOperation{
		{Op:"assert",Fact:Fact{Predicate:"present",Args:[]string{"alice"}}},
		{Op:"explode",Fact:Fact{Predicate:"present",Args:[]string{"bob"}}},
	})
	if err==nil{t.Fatal("expected batch failure")}
	facts,err:=s.Facts("state");if err!=nil{t.Fatal(err)};if len(facts)!=0{t.Fatalf("partial commit: %+v",facts)}
}
func TestFactBatchMixedAssertRetract(t *testing.T){
	s,err:=NewStore(t.TempDir());if err!=nil{t.Fatal(err)}
	if err:=s.Put("state",`active(X) :- member(X).`);err!=nil{t.Fatal(err)}
	alice:=Fact{Predicate:"member",Args:[]string{"alice"}};bob:=Fact{Predicate:"member",Args:[]string{"bob"}}
	if err:=s.AssertFact("state",alice);err!=nil{t.Fatal(err)}
	if err:=s.ApplyFacts("state",[]FactOperation{{Op:"retract",Fact:alice},{Op:"assert",Fact:bob}});err!=nil{t.Fatal(err)}
	a,_:=s.Query(context.Background(),"state",`active("alice").`);b,_:=s.Query(context.Background(),"state",`active("bob").`)
	if len(a)!=0||len(b)!=1{t.Fatalf("alice=%v bob=%v",a,b)}
}

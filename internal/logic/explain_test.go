package logic

import (
	"context"
	"testing"
)

func TestExplainReturnsRuleSuppliedProof(t *testing.T) {
	e:=New()
	if err:=e.Consult(`trusted(alice). may_voice(X, fact(trusted(X))) :- trusted(X).`); err!=nil { t.Fatal(err) }
	got,err:=e.Explain(context.Background(),`may_voice(alice, Proof).`); if err!=nil { t.Fatal(err) }
	if !got.Result || len(got.Proofs)!=1 { t.Fatalf("explanation=%+v",got) }
	if got.Proofs[0]["Proof"]=="" { t.Fatalf("missing Proof: %+v",got) }
}

func TestExplainFalseHasNoProofs(t *testing.T) {
	e:=New()
	if err:=e.Consult(`trusted(alice). may_voice(X, fact(trusted(X))) :- trusted(X).`); err!=nil { t.Fatal(err) }
	got,err:=e.Explain(context.Background(),`may_voice(bob, Proof).`); if err!=nil { t.Fatal(err) }
	if got.Result || len(got.Proofs)!=0 { t.Fatalf("explanation=%+v",got) }
}

func TestExplainRequiresProofBinding(t *testing.T) {
	e:=New(); if err:=e.Consult(`trusted(alice).`); err!=nil { t.Fatal(err) }
	if _,err:=e.Explain(context.Background(),`trusted(X).`); err==nil { t.Fatal("expected missing Proof error") }
}

package logic

import (
    "context"
    "testing"
)

func TestConsultAndQuery(t *testing.T) {
    e := New()
    if err := e.Consult(`trusted(alice). trusted(bob). may_voice(X) :- trusted(X).`); err != nil {
        t.Fatal(err)
    }
    rows, err := e.Query(context.Background(), `may_voice(X).`)
    if err != nil {
        t.Fatal(err)
    }
    if len(rows) != 2 {
        t.Fatalf("solutions=%d, want 2", len(rows))
    }
}

func TestDeterministicFalseQuery(t *testing.T) {
    e := New()
    if err := e.Consult(`trusted(alice).`); err != nil {
        t.Fatal(err)
    }
    rows, err := e.Query(context.Background(), `trusted(carol).`)
    if err != nil {
        t.Fatal(err)
    }
    if len(rows) != 0 {
        t.Fatalf("solutions=%d, want 0", len(rows))
    }
}

package logic

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestRulesetsAreIsolated(t *testing.T) {
	s,err:=NewStore(t.TempDir()); if err!=nil { t.Fatal(err) }
	if err:=s.Put("ops",`trusted(alice).`); err!=nil { t.Fatal(err) }
	if err:=s.Put("games",`trusted(bob).`); err!=nil { t.Fatal(err) }
	a,err:=s.Query(context.Background(),"ops",`trusted(alice).`); if err!=nil || len(a)!=1 { t.Fatalf("ops=%v err=%v",a,err) }
	b,err:=s.Query(context.Background(),"games",`trusted(alice).`); if err!=nil || len(b)!=0 { t.Fatalf("games=%v err=%v",b,err) }
}

func TestRulesetsPersistAcrossRestart(t *testing.T) {
	dir:=t.TempDir(); s,err:=NewStore(dir); if err!=nil { t.Fatal(err) }
	if err:=s.Put("irc-policy",`may_voice(alice).`); err!=nil { t.Fatal(err) }
	if _,err:=os.Stat(filepath.Join(dir,"irc-policy.pl")); err!=nil { t.Fatal(err) }
	s2,err:=NewStore(dir); if err!=nil { t.Fatal(err) }
	rows,err:=s2.Query(context.Background(),"irc-policy",`may_voice(alice).`); if err!=nil || len(rows)!=1 { t.Fatalf("rows=%v err=%v",rows,err) }
}

func TestInvalidRulesetNameRejected(t *testing.T) {
	s,err:=NewStore(t.TempDir()); if err!=nil { t.Fatal(err) }
	if err:=s.Put("../escape",`x.`); err==nil { t.Fatal("expected invalid name error") }
}

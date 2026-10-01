package api

import "testing"

func TestStableM1VersionContract(t *testing.T){
	got:=VersionResponse()
	if got.Service!="botlogic"{t.Fatalf("service=%q",got.Service)}
	if got.Version!="1.0.0"{t.Fatalf("version=%q",got.Version)}
	if got.API!="v1"{t.Fatalf("api=%q",got.API)}
}

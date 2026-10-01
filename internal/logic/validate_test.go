package logic

import "testing"

func TestValidateDoesNotNeedStore(t *testing.T){
	if err:=Validate("allowed(alice).");err!=nil{t.Fatal(err)}
}
func TestValidateRejectsInvalidProlog(t *testing.T){
	if err:=Validate("allowed(");err==nil{t.Fatal("expected syntax error")}
}
func TestValidateRejectsEmptySource(t *testing.T){
	if err:=Validate("");err==nil{t.Fatal("expected empty source error")}
}

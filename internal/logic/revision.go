package logic

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var ErrRevisionConflict = errors.New("revision conflict")

type revisionFile struct { Revision uint64 `json:"revision"` }

func (s *Store) revisionPath(name string) string { return filepath.Join(s.dir,name+".revision.json") }

func (s *Store) readRevision(name string)(uint64,error){
	if s.dir=="" { return 0,nil }
	b,err:=os.ReadFile(s.revisionPath(name));if errors.Is(err,os.ErrNotExist){return 0,nil};if err!=nil{return 0,err}
	var v revisionFile;if err:=json.Unmarshal(b,&v);err!=nil{return 0,err};return v.Revision,nil
}
func (s *Store) writeRevision(name string,revision uint64) error {
	if s.dir=="" { return nil };b,err:=json.Marshal(revisionFile{Revision:revision});if err!=nil{return err};b=append(b,'\n')
	tmp,err:=os.CreateTemp(s.dir,"."+name+"-revision-*.tmp");if err!=nil{return err};tmpName:=tmp.Name();defer os.Remove(tmpName)
	if err:=tmp.Chmod(0600);err!=nil{tmp.Close();return err};if _,err:=tmp.Write(b);err!=nil{tmp.Close();return err};if err:=tmp.Sync();err!=nil{tmp.Close();return err};if err:=tmp.Close();err!=nil{return err}
	return os.Rename(tmpName,s.revisionPath(name))
}
func revisionConflict(expected,current uint64) error { return fmt.Errorf("%w: expected %d, current %d",ErrRevisionConflict,expected,current) }

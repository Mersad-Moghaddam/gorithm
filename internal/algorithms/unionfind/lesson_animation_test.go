package unionfind

import (
 "encoding/json"
 "reflect"
 "testing"
 "github.com/Mersad-Moghaddam/gorithm/internal/catalog"
 "github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestUnionFindLessonCompression(t *testing.T) {
 steps,_ := (Lesson{}).Build(nil)
 found := false
 for k,step := range steps {
  s := step.State.(renderers.UnionFindLessonState)
  if s.Phase == "compress" && s.ChangedNode == 4 {
   before := steps[k-1].State.(renderers.UnionFindLessonState)
   if step.LineID != "L7" || s.OldParent != 3 || s.NewParent != 0 || before.Parent[4] != 3 || s.Parent[4] != 0 || s.Components != 1 || s.Rank[0] != 2 || !reflect.DeepEqual(s.Path,[]int{4,3,0}) { t.Fatalf("wrong compression: %+v",s) }
   found = true
  }
 }
 if !found { t.Fatal("missing compression") }
 if !reflect.DeepEqual(steps[len(steps)-1].State.(renderers.UnionFindLessonState).Parent,[]int{0,0,0,0,0}) { t.Fatal("wrong final parents") }
}

func TestUnionFindLessonNestedFind(t *testing.T) {
 for _,caller := range []string{"union","connected"} {
  steps := Trace(4,[]Operation{{"union",0,1},{"union",2,3},{"union",0,2},{caller,3,1}})
  found := false
  for _,step := range steps { s := step.State.(renderers.UnionFindLessonState); if s.Operation == caller && s.Phase == "compress" && s.ChangedNode == 3 && s.OldParent == 2 && s.NewParent == 0 && step.LineID == "L7" { found = true } }
  if !found { t.Fatalf("missing nested compression: %s",caller) }
 }
}

func TestUnionFindLessonRanksAndRedundant(t *testing.T) {
 steps := Trace(4,[]Operation{{"connected",0,1},{"union",0,1},{"union",2,0},{"union",0,3},{"union",0,1},{"union",0,0},{"find",2,0},{"connected",2,3}})
 seen := map[string]bool{}
 prev := steps[0].State.(renderers.UnionFindLessonState)
 for _,step := range steps[1:] {
  s := step.State.(renderers.UnionFindLessonState); seen[step.LineID] = true
  for i,rank := range s.Rank { if rank != prev.Rank[i] && (step.LineID != "L16" || rank != prev.Rank[i]+1) { t.Fatal("rank changed outside tie") } }
  if s.Components != prev.Components && (s.Phase != "link" || s.Components != prev.Components-1) { t.Fatal("wrong component count") }
  if s.Redundant { seen["redundant"] = true }; if s.Operation == "connected" && s.ResultKnown { if s.Connected { seen["true"] = true } else { seen["false"] = true } }; prev = s
 }
 for _,key := range []string{"L14","L15","L16","redundant","true","false"} { if !seen[key] { t.Errorf("missing %s",key) } }
}

func TestUnionFindLessonSnapshotsAndGuards(t *testing.T) {
 if len(Trace(0,nil)) != 0 || len(Trace(-1,nil)) != 0 { t.Fatal("nonpositive n") }
 if len(Trace(1,[]Operation{{"find",-1,0},{"union",0,1},{"connected",2,0},{"unknown",0,0}})) != 2 { t.Fatal("invalid operations not skipped") }
 ops := []Operation{{"union",0,1},{"find",1,0}}; steps := Trace(2,ops)
 before,_ := json.Marshal(steps); ops[0].I = 1; Trace(2,ops); after,_ := json.Marshal(steps)
 if string(before) != string(after) { t.Fatal("input alias") }
 first := steps[0].State.(renderers.UnionFindLessonState); last := steps[len(steps)-1].State.(renderers.UnionFindLessonState)
 last.Parent[0],last.Rank[0] = 99,99; steps[len(steps)-1].Variables["parent"].([]int)[0] = 88
 if first.Parent[0] != 0 || first.Rank[0] != 0 || steps[0].Variables["parent"].([]int)[0] != 0 { t.Fatal("snapshot alias") }
 if err := catalog.ValidateSourceLineIDs(Lesson{},steps); err != nil { t.Fatal(err) }
 paired := false
 for i,step := range steps { if step.HoldMS != 0 && step.HoldMS != 350 && step.HoldMS != 600 && step.HoldMS != 1200 && step.HoldMS != 1800 { t.Fatal("invalid hold") }; if i>0 && step.State.(renderers.UnionFindLessonState).Phase == "link" && steps[i-1].HoldMS+step.HoldMS == 1800 { paired = true } }
 if !paired { t.Fatal("missing compare/commit budget") }
}

func TestUnionFindLessonValidation(t *testing.T) {
 for _,input := range []string{`null`,`{`,`[]`,`{"n":null}`,`{"n":0}`,`{"n":65}`,`{"operations":null}`,`{"operations":[{"op":"oops","i":0}]}`,`{"n":2,"operations":[{"op":"find","i":2}]}`,`{"operations":[{"op":"union","i":0}]}`,`{"operations":[{"op":"find"}]}`,`{"operations":[{"op":"find","i":null}]}`} { if _,err := (Lesson{}).Build(json.RawMessage(input)); err == nil { t.Errorf("accepted %s",input) } }
 for _,input := range []string{`{}`,`{"n":1,"operations":[]}`,`{"n":64,"operations":[{"op":"find","i":63}]}`} { if _,err := (Lesson{}).Build(json.RawMessage(input)); err != nil { t.Errorf("rejected %s: %v",input,err) } }
 ops := make([]Operation,128); for i := range ops { ops[i] = Operation{Op:"find",I:0} }
 input,_ := json.Marshal(LessonInput{N:1,Operations:ops}); if _,err := (Lesson{}).Build(input); err != nil { t.Fatal(err) }
 ops = append(ops,Operation{Op:"find",I:0}); input,_ = json.Marshal(LessonInput{N:1,Operations:ops}); if _,err := (Lesson{}).Build(input); err == nil { t.Fatal("accepted 129 operations") }
}

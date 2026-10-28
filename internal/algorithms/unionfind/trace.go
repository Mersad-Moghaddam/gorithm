package unionfind

import (
 "fmt"
 "slices"

 "github.com/Mersad-Moghaddam/gorithm/internal/trace"
 "github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// Operation represents an action to execute in union-find trace.
type Operation struct {
 Op string `json:"op"`
 I int `json:"i"`
 J int `json:"j,omitempty"`
}

// Trace records actual parent pointers, including compression within every caller.
func Trace(n int, ops []Operation) []trace.Step {
 b := trace.NewBuilder()
 if n <= 0 { return b.Steps() }
 d := New(n)
 s := renderers.UnionFindLessonState{I:-1,J:-1,RootI:-1,RootJ:-1,Active:-1,ChangedNode:-1,OldParent:-1,NewParent:-1}
 emit := func(line string, action trace.ActionType, phase, explanation string) {
  s.Phase = phase
  snapshot := s
  snapshot.Parent, snapshot.Rank, snapshot.Path = slices.Clone(d.parent), slices.Clone(d.rank), slices.Clone(s.Path)
  snapshot.Components = d.count
  b.AddStep(line,action,explanation,snapshot,map[string]any{
   "parent":slices.Clone(d.parent),"rank":slices.Clone(d.rank),"path":slices.Clone(s.Path),"components":d.count,
   "i":s.I,"j":s.J,"root_i":s.RootI,"root_j":s.RootJ,"result_known":s.ResultKnown,"connected":s.Connected,"redundant":s.Redundant,
  })
 }
 emit("L1",trace.ActionInit,"init",fmt.Sprintf("Each of %d elements starts as its own root. Rank starts at zero.",n))
 var find func(int) int
 find = func(i int) int {
  s.Active = i
  s.Path = append(s.Path,i)
  emit("L5",trace.ActionHighlight,"walk",fmt.Sprintf("Find visits %d; parent[%d] = %d.",i,i,d.parent[i]))
  b.IncComparisons()
  emit("L6",trace.ActionCompare,"walk",fmt.Sprintf("Check %d against parent %d: root = %t.",i,d.parent[i],d.parent[i] == i))
  if d.parent[i] != i {
   old := d.parent[i]
   root := find(old)
   if old != root {
    d.parent[i] = root
    b.IncOperations()
    s.Active,s.ChangedNode,s.OldParent,s.NewParent = i,i,old,root
    emit("L7",trace.ActionSet,"compress",fmt.Sprintf("Compress parent[%d]: %d becomes %d. Components and rank stay unchanged.",i,old,root))
    s.ChangedNode,s.OldParent,s.NewParent = -1,-1,-1
   }
  }
  s.Active = i
  emit("L9",trace.ActionHighlight,"roots",fmt.Sprintf("Find(%d) returns root %d.",i,d.parent[i]))
  return d.parent[i]
 }
 for _,op := range ops {
  if op.I < 0 || op.I >= n { continue }
  if op.Op != "find" && (op.J < 0 || op.J >= n) { continue }
  if op.Op != "find" && op.Op != "union" && op.Op != "connected" { continue }
  s = renderers.UnionFindLessonState{Operation:op.Op,I:op.I,J:-1,RootI:-1,RootJ:-1,Active:-1,ChangedNode:-1,OldParent:-1,NewParent:-1}
  if op.Op == "find" {
   s.RootI = find(op.I)
   s.ResultKnown = true
   emit("L9",trace.ActionDone,"done",fmt.Sprintf("Find(%d) = %d. Component count stays %d.",op.I,s.RootI,d.count))
   continue
  }
  s.J = op.J
  line := "L11"; if op.Op == "connected" { line = "L19" }
  emit(line,trace.ActionHighlight,"walk",fmt.Sprintf("%s(%d, %d): find both representative roots.",op.Op,op.I,op.J))
  s.RootI = find(op.I)
  s.Path = nil
  s.RootJ = find(op.J)
  s.Active = -1
  if op.Op == "connected" {
   b.IncComparisons()
   s.Connected,s.ResultKnown = s.RootI == s.RootJ,true
   emit("L20",trace.ActionCompare,"compare",fmt.Sprintf("Roots %d and %d: connected = %t.",s.RootI,s.RootJ,s.Connected))
   continue
  }
  emit("L12",trace.ActionHighlight,"roots",fmt.Sprintf("Roots found: %d and %d.",s.RootI,s.RootJ))
  b.IncComparisons()
  s.Redundant = s.RootI == s.RootJ
  emit("L13",trace.ActionCompare,"compare",fmt.Sprintf("Roots equal = %t. Redundant unions do not change rank or component count.",s.Redundant))
  if s.Redundant { continue }
  ri,rj := s.RootI,s.RootJ
  s.OldParent = -1
  if d.rank[ri] < d.rank[rj] { line = "L14"; s.ChangedNode,s.NewParent = ri,rj
  } else if d.rank[ri] > d.rank[rj] { line = "L15"; s.ChangedNode,s.NewParent = rj,ri
  } else { line = "L16"; s.ChangedNode,s.NewParent = rj,ri }
  b.IncComparisons()
  emit(line,trace.ActionCompare,"compare",fmt.Sprintf("Compare rank[%d] = %d and rank[%d] = %d; link root %d under %d.",ri,d.rank[ri],rj,d.rank[rj],s.ChangedNode,s.NewParent))
  s.OldParent = d.parent[s.ChangedNode]
  d.parent[s.ChangedNode] = s.NewParent
  if line == "L16" { d.rank[ri]++ }
  d.count--
  b.IncOperations()
  emit(line,trace.ActionSet,"link",fmt.Sprintf("parent[%d] = %d. Components: %d; root rank %d%s.",s.ChangedNode,s.NewParent,d.count,d.rank[s.NewParent],map[bool]string{true:" (tie increment)",false:" (unchanged)"}[line == "L16"]))
  s.ChangedNode,s.OldParent,s.NewParent = -1,-1,-1
  s.ResultKnown = true
  emit("L17",trace.ActionDone,"done",fmt.Sprintf("Union(%d, %d) merged two sets. Components: %d.",op.I,op.J,d.count))
 }
 s.Operation = ""
 s.Active,s.ChangedNode,s.OldParent,s.NewParent = -1,-1,-1,-1
 s.Path = nil
 emit("L17",trace.ActionDone,"done",fmt.Sprintf("All operations complete. Components: %d.",d.count))
 steps := b.Steps()
 for i := range steps {
  state := steps[i].State.(renderers.UnionFindLessonState)
  if state.Phase == "link" && i>0 { steps[i-1].HoldMS,steps[i].HoldMS = 600,1200 }
 }
 return steps
}

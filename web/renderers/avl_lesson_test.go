package renderers_test

import (
 "encoding/xml"
 "io"
 "reflect"
 "strings"
 "testing"

 "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/avl"
 "github.com/Mersad-Moghaddam/gorithm/internal/trace"
 "github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestAVLLessonSVG(t *testing.T) {
 for _,step:=range avl.Trace([]int{30,20,40,10,25,5}) {
  before:=renderers.RenderAVLLessonSVG(step,320,420)
  state:=step.State.(renderers.AVLLessonState)
  clone:=state; clone.Nodes=append([]renderers.TreeNode(nil),state.Nodes...)
  for _,size:=range [][2]int{{320,420},{800,480},{0,-1}} {
   svg:=renderers.RenderAVLLessonSVG(step,size[0],size[1])
   if strings.Contains(svg,"NaN")||strings.Contains(svg,"Inf") {t.Fatal("nonfinite SVG")}
   d:=xml.NewDecoder(strings.NewReader(`<svg xmlns="http://www.w3.org/2000/svg">`+svg+`</svg>`)); ids:=map[string]bool{}
   for {tok,err:=d.Token();if err==io.EOF {break};if err!=nil {t.Fatal(err)};if el,ok:=tok.(xml.StartElement);ok {for _,a:=range el.Attr {if a.Name.Local=="data-entity" {if ids[a.Value] {t.Fatal("duplicate entity")};ids[a.Value]=true}}}}
   if len(ids)!=len(state.Nodes) {t.Fatal("missing movable nodes")}
   if len(state.Nodes)>0 && (!strings.Contains(svg,"h=")||!strings.Contains(svg,"bf=")) {t.Fatal("missing height/balance badges")}
   if state.Rotation!="" && (!strings.Contains(svg,`data-rotation-arc`)||!strings.Contains(svg,"marker-end")) {t.Fatal("missing rotation arc")}
  }
  if before!=renderers.RenderAVLLessonSVG(step,320,420)||!reflect.DeepEqual(state,clone) {t.Fatal("render mutates snapshot")}
 }
 for _,step:=range []trace.Step{{},{State:42}} {if !strings.Contains(renderers.RenderAVLLessonSVG(step,320,420),"No AVL state") {t.Fatal("missing fallback")}}
 s:=renderers.AVLLessonState{TreeState:renderers.TreeState{Nodes:[]renderers.TreeNode{{ID:`node-1`,Label:`<>&"'`,Val:1}}},Heights:[]int{1},Balances:[]int{0}}
 svg:=renderers.RenderAVLLessonSVG(trace.Step{State:s},320,420); if strings.Contains(svg,`<>&`) || !strings.Contains(svg,"&lt;") {t.Fatal("unescaped label")}
}

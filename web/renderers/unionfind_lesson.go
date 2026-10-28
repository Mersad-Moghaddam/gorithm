package renderers

import (
 "fmt"
 "html"
 "math"
 "strings"
 "github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

// UnionFindLessonState is an immutable parent forest, never a union-request graph.
type UnionFindLessonState struct {
 Parent,Rank []int
 Phase,Operation string
 I,J,RootI,RootJ,Active int
 Path []int
 ChangedNode,OldParent,NewParent int
 Components int
 Connected,ResultKnown,Redundant bool
}

func RenderUnionFindLessonSVG(step trace.Step,width,height int) string {
 fallback := `<g data-lesson="unionfind" data-phase="empty" data-scene-width="800" data-scene-height="450"><title>Union-find state unavailable</title><text x="32" y="48" fill="#eff5ed" font-size="14">No union-find state available</text></g>`
 s,ok := step.State.(UnionFindLessonState)
 n := len(s.Parent)
 if !ok || n == 0 || n>256 || len(s.Rank) != n || width<=0 || height<=0 { return fallback }
 children := make([][]int,n); roots := []int{}
 for i,p := range s.Parent { if p<0 || p>=n { return fallback }; if p==i { roots=append(roots,i) } else { children[p]=append(children[p],i) } }
 // Validate all chains before recursion, including cycles without roots.
 for i := range s.Parent { p := i; for count:=0; s.Parent[p]!=p; count++ { if count>=n { return fallback }; p=s.Parent[p] } }
 spans := make([]float64,n)
 var span func(int) float64
 span = func(i int) float64 { total:=0.0; for _,child := range children[i] { total+=span(child) }; if total==0 { total=76 }; spans[i]=total; return total }
 forestWidth:=64.0; for _,root := range roots { forestWidth+=span(root)+76 }; forestWidth-=76
 w:=math.Max(800,math.Max(forestWidth,float64(96+56*n)))
 x,y := make([]float64,n),make([]float64,n)
 maxDepth:=0
 var place func(int,float64,int)
 place = func(i int,left float64,depth int) { x[i],y[i]=left+spans[i]/2,88+80*float64(depth); if depth>maxDepth { maxDepth=depth }; for _,child := range children[i] { place(child,left,depth+1); left+=spans[child] } }
 left:=32+(w-forestWidth)/2; for _,root := range roots { place(root,left,0); left+=spans[root]+76 }
 tableY:=168+80*maxDepth; h:=math.Max(450,float64(tableY+128))
 var b strings.Builder
 fmt.Fprintf(&b,`<g data-lesson="unionfind" data-phase="%s" data-scene-width="%.0f" data-scene-height="%.0f" data-min-width="%.0f" font-family="system-ui, sans-serif" font-size="14" fill="#eff5ed"><title>Union-find %s: %s; %d components</title>`,html.EscapeString(s.Phase),w,h,w,html.EscapeString(s.Operation),html.EscapeString(s.Phase),s.Components)
 fmt.Fprintf(&b,`<text x="32" y="30">Components: %d · %s</text>`,s.Components,html.EscapeString(s.Phase))
 status:="Parent arrows point toward roots; rank is not height."
 if s.Redundant { status="Already connected: redundant union; no change." } else if s.Operation=="connected" && s.ResultKnown { status=fmt.Sprintf("Roots %d and %d: connected = %t",s.RootI,s.RootJ,s.Connected) } else if s.Phase=="compare" { status=fmt.Sprintf("Compare roots %d and %d",s.RootI,s.RootJ) } else if s.Phase=="done" { status="✓ Checked; parent forest shown below." }
 fmt.Fprintf(&b,`<text x="32" y="52" fill="#bbcab9">%s</text>`,html.EscapeString(status))
 b.WriteString(`<defs><marker id="unionfind-arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto"><path d="M0 0L10 5L0 10Z" fill="#8bd5ff"/></marker></defs>`)
 edge := func(child,parent int,old bool) {
  dx,dy:=x[parent]-x[child],y[parent]-y[child]; length:=math.Hypot(dx,dy); if length==0 { return }
  attrs:=fmt.Sprintf(`data-edge="parent" data-child="unionfind-node-%d" data-parent="unionfind-node-%d"`,child,parent)
  if old { attrs=`data-edge="old-parent" stroke-dasharray="6 5"` }
  fmt.Fprintf(&b,`<line %s x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f" stroke="#8bd5ff" stroke-width="2" marker-end="url(#unionfind-arrow)"/>`,attrs,x[child]+22*dx/length,y[child]+22*dy/length,x[parent]-26*dx/length,y[parent]-26*dy/length)
 }
 for i,p:=range s.Parent { if i!=p { edge(i,p,false) } }
 if s.Phase=="compress" && s.ChangedNode>=0 && s.ChangedNode<n && s.OldParent>=0 && s.OldParent<n && s.OldParent!=s.ChangedNode { edge(s.ChangedNode,s.OldParent,true) }
 palette:=[]string{"#c45d75","#387faa","#8763bb","#b67b30","#258b80"}
 for i,p:=range s.Parent {
  outline:="#bbcab9"; cue:=""
  for _,visited:=range s.Path { if visited==i { outline="#9bf0bb"; cue="checked" } }
  if i==s.Active { outline="#ffe29a"; cue="active" }
  if s.Phase=="compare" && (i==s.RootI || i==s.RootJ) { outline="#ffe29a"; cue="compare" }
  if p==i { cue="root" }
  fmt.Fprintf(&b,`<g id="unionfind-node-%d" data-entity="unionfind-node-%d" data-x="%.2f" data-y="%.2f" data-motion="linear" transform="translate(%.2f %.2f)"><circle r="22" fill="%s" stroke="%s" stroke-width="3"/><text y="6" text-anchor="middle" fill="#ffffff" font-size="16">%d</text><text y="42" text-anchor="middle" fill="%s">%s</text></g>`,i,i,x[i],y[i],x[i],y[i],palette[i%len(palette)],outline,i,outline,cue)
 }
 for row,label:=range []string{"index","parent","rank"} {
  yy:=tableY+row*32
  fmt.Fprintf(&b,`<text x="32" y="%d" fill="#bbcab9">%s</text>`,yy+21,label)
  for i:=0;i<n;i++ { value:=i; if row==1 { value=s.Parent[i] }; if row==2 { value=s.Rank[i] }; fmt.Fprintf(&b,`<rect x="%d" y="%d" width="56" height="32" fill="#192923" stroke="#52634e"/><text x="%d" y="%d" text-anchor="middle" font-size="16">%d</text>`,88+56*i,yy,116+56*i,yy+22,value) }
 }
 b.WriteString(`</g>`)
 return b.String()
}

package renderers

import (
	"fmt"
	"html"
	"strconv"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type AVLLessonState struct {
	TreeState
	Phase string
	Target int
	ActiveID, PivotID, NewRootID, TransferID string
	PathIDs, LeftIDs, RightIDs []string
	Heights, Balances []int
	Rotation, Case string
	RotationPart int
	HeightsPending bool
}

func RenderAVLLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(AVLLessonState)
	if !ok { return `<g data-lesson="avl" data-phase="empty" data-scene-width="800" data-scene-height="440"><text x="32" y="48" fill="#eff5ed" font-size="14">No AVL state available</text></g>` }
	w,h:=800.0,440.0
	if x:=float64(88*(len(s.Nodes)+1));x>w {w=x}
	for _,n:=range s.Nodes {if n.Y+88>h {h=n.Y+88}}
	var b strings.Builder
	fmt.Fprintf(&b,`<g data-lesson="avl" data-phase="%s" data-scene-width="%.0f" data-scene-height="%.0f" data-min-width="%.0f" font-family="system-ui,sans-serif" font-size="14"><title>AVL tree: %s</title>`,html.EscapeString(s.Phase),w,h,w,html.EscapeString(s.Phase))
	label:=s.Phase
	if s.Case!="" {label=fmt.Sprintf("%s · %s rotation %d",s.Case,s.Rotation,s.RotationPart); if s.Case=="LR"||s.Case=="RL" {label+="/2"}}
	if s.HeightsPending {label+=" · heights pending recomputation"}
	if s.Phase=="done" {label="✓ checked · balanced"}
	if s.Phase=="compare" {label=fmt.Sprintf("Compare target %d",s.Target)}
	fmt.Fprintf(&b,`<text x="32" y="28" fill="#eff5ed">%s</text>`,html.EscapeString(label))
	nodes:=map[string]TreeNode{}; for _,n:=range s.Nodes {nodes[n.ID]=n}
	for i,n:=range s.Nodes {if n.ID==s.ActiveID && i<len(s.LeftIDs) && i<len(s.RightIDs) && i<len(s.Balances) {lh,rh:=0,0;for j,c:=range s.Nodes {if j>=len(s.Heights){continue};if c.ID==s.LeftIDs[i]{lh=s.Heights[j]};if c.ID==s.RightIDs[i]{rh=s.Heights[j]}};fmt.Fprintf(&b,`<text x="32" y="52" fill="#ffe29a">bf = %d − %d = %d · null height = 0</text>`,lh,rh,s.Balances[i])}}
	for _,n:=range s.Nodes {if p,exists:=nodes[n.ParentID];exists {fmt.Fprintf(&b,`<path id="avl-edge-%s" data-parent="%s" data-child="%s" d="M %.0f %.0f L %.0f %.0f" stroke="#bbcab9" stroke-width="2" fill="none"/>`,html.EscapeString(n.ID),html.EscapeString(p.ID),html.EscapeString(n.ID),p.X,p.Y+22,n.X,n.Y-22)}}
	if p,pok:=nodes[s.PivotID];pok && s.Rotation!="" {if n,nok:=nodes[s.NewRootID];nok {fmt.Fprintf(&b,`<defs><marker id="avl-rotation-arrow" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="6" markerHeight="6" orient="auto"><path d="M0 0 L10 5 L0 10 Z" fill="#8bd5ff"/></marker></defs><path data-rotation-arc="%s" d="M %.0f %.0f Q %.0f %.0f %.0f %.0f" fill="none" stroke="#8bd5ff" stroke-width="3" marker-end="url(#avl-rotation-arrow)"/>`,html.EscapeString(s.Rotation),p.X,p.Y-28,(p.X+n.X)/2,64.0,n.X,n.Y-28)}}
	palette:=[]string{"#c45d75","#387faa","#8763bb","#b67b30","#258b80"}
	for i,n:=range s.Nodes {
		ordinal,_:=strconv.Atoi(strings.TrimPrefix(n.ID,"node-"));if ordinal<1 {ordinal=1}
		stroke:="#bbcab9"; if n.ID==s.ActiveID {stroke="#ffe29a"};if s.Phase=="done" {stroke="#9bf0bb"}
		text:=strconv.Itoa(n.Val);if n.Label!="" {text=n.Label}
		fmt.Fprintf(&b,`<g id="avl-%s" data-entity="avl:%s" data-x="%.0f" data-y="%.0f" data-motion="linear" transform="translate(%.0f %.0f)"><circle r="22" fill="%s" stroke="%s" stroke-width="3"/><text y="6" text-anchor="middle" fill="#ffffff" font-size="16">%s</text><text y="40" text-anchor="middle" fill="#eff5ed">%s</text>`,html.EscapeString(n.ID),html.EscapeString(n.ID),n.X,n.Y,n.X,n.Y,palette[(ordinal-1)%5],stroke,html.EscapeString(text),html.EscapeString(n.ID))
		if i<len(s.Heights)&&i<len(s.Balances) {fmt.Fprintf(&b,`<text y="58" text-anchor="middle" fill="#eff5ed">h=%d</text><text y="76" text-anchor="middle" fill="#eff5ed">bf=%d</text>`,s.Heights[i],s.Balances[i])}
		if n.ID==s.TransferID {b.WriteString(`<rect x="-38" y="-30" width="76" height="112" rx="10" fill="none" stroke="#8bd5ff" stroke-dasharray="5 4"/><text x="40" y="0" fill="#8bd5ff">transfer</text>`)}
		b.WriteString(`</g>`)
	}
	if len(s.Nodes)==0 {b.WriteString(`<text x="32" y="96" fill="#bbcab9">Empty tree · ready to insert</text>`)}
	b.WriteString(`</g>`);return b.String()
}

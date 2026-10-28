package avl

import (
	"math/rand"
	"reflect"
	"sort"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func isAVL(t *testing.T, n *Node) (int, bool) {
	if n == nil {
		return 0, true
	}

	lh, lok := isAVL(t, n.Left)
	rh, rok := isAVL(t, n.Right)

	if !lok || !rok {
		return 0, false
	}

	if n.Left != nil && n.Left.Val >= n.Val {
		t.Errorf("BST invariant violated: left child %d >= node %d", n.Left.Val, n.Val)
		return 0, false
	}
	if n.Right != nil && n.Right.Val <= n.Val {
		t.Errorf("BST invariant violated: right child %d <= node %d", n.Right.Val, n.Val)
		return 0, false
	}

	diff := lh - rh
	if diff < -1 || diff > 1 {
		t.Errorf("AVL balance violated at node %d: left height %d, right height %d, diff %d", n.Val, lh, rh, diff)
		return 0, false
	}

	expectedHeight := 1
	if lh > rh {
		expectedHeight += lh
	} else {
		expectedHeight += rh
	}

	if n.Height != expectedHeight {
		t.Errorf("node %d stored height %d != expected height %d", n.Val, n.Height, expectedHeight)
		return 0, false
	}

	return expectedHeight, true
}

func TestAVLCorrectness(t *testing.T) {
	tree := New()

	if tree.Search(10) {
		t.Fatalf("expected Search(10) on empty tree to be false")
	}
	if len(tree.InOrder()) != 0 {
		t.Fatalf("expected empty InOrder slice, got %v", tree.InOrder())
	}

	values := []int{10, 20, 30, 40, 50, 25}
	for _, v := range values {
		tree.Insert(v)
	}

	// Duplicate insert should be no-op
	tree.Insert(30)
	tree.Insert(10)

	expectedOrder := []int{10, 20, 25, 30, 40, 50}
	if !reflect.DeepEqual(tree.InOrder(), expectedOrder) {
		t.Fatalf("InOrder got %v, want %v", tree.InOrder(), expectedOrder)
	}

	for _, v := range values {
		if !tree.Search(v) {
			t.Fatalf("expected Search(%d) to be true", v)
		}
	}

	for _, missing := range []int{5, 15, 35, 60, 100} {
		if tree.Search(missing) {
			t.Fatalf("expected Search(%d) to be false", missing)
		}
	}

	if _, ok := isAVL(t, tree.Root); !ok {
		t.Fatalf("tree failed AVL balance property after [10, 20, 30, 40, 50, 25]")
	}
}

func TestAVLBalancePropertySequentialAndRandom(t *testing.T) {
	// 1. Sequential ascending inserts (triggers Right-Right / Left rotations)
	treeAsc := New()
	for i := 1; i <= 100; i++ {
		treeAsc.Insert(i)
	}
	if _, ok := isAVL(t, treeAsc.Root); !ok {
		t.Fatalf("ascending sequential inserts failed AVL invariant")
	}
	if treeAsc.Root.Height > 8 {
		t.Fatalf("tree height %d exceeds O(log n) bound for 100 nodes", treeAsc.Root.Height)
	}

	// 2. Sequential descending inserts (triggers Left-Left / Right rotations)
	treeDesc := New()
	for i := 100; i >= 1; i-- {
		treeDesc.Insert(i)
	}
	if _, ok := isAVL(t, treeDesc.Root); !ok {
		t.Fatalf("descending sequential inserts failed AVL invariant")
	}
	if treeDesc.Root.Height > 8 {
		t.Fatalf("tree height %d exceeds O(log n) bound for 100 nodes", treeDesc.Root.Height)
	}

	// 3. Random inserts (triggers LL, RR, LR, RL rotations)
	rng := rand.New(rand.NewSource(42))
	nums := rng.Perm(200)
	treeRand := New()
	for _, n := range nums {
		treeRand.Insert(n)
	}
	if _, ok := isAVL(t, treeRand.Root); !ok {
		t.Fatalf("random inserts failed AVL invariant")
	}

	sorted := make([]int, len(nums))
	copy(sorted, nums)
	sort.Ints(sorted)
	if !reflect.DeepEqual(treeRand.InOrder(), sorted) {
		t.Fatalf("InOrder does not match sorted elements for random inserts")
	}
}

func TestAVLLessonValidation(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build trace: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("trace produced 0 steps")
	}

	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("line validation failed: %v", err)
	}

	meta := lesson.Meta()
	if meta.ID != "avl" {
		t.Errorf("meta.ID = %q; want 'avl'", meta.ID)
	}
	if meta.Name != "AVL Tree (Self-Balancing)" {
		t.Errorf("meta.Name = %q; want 'AVL Tree (Self-Balancing)'", meta.Name)
	}
	if meta.Category != catalog.CategoryTrees {
		t.Errorf("meta.Category = %q; want %q", meta.Category, catalog.CategoryTrees)
	}
	if meta.Renderer != "tree" {
		t.Errorf("meta.Renderer = %q; want 'tree'", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(log n)" {
		t.Errorf("TimeComplexity = %q; want 'O(log n)'", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(n)" {
		t.Errorf("SpaceComplexity = %q; want 'O(n)'", meta.Documentation.SpaceComplexity)
	}
	if meta.Documentation.Difficulty != "Hard" {
		t.Errorf("Difficulty = %q; want 'Hard'", meta.Documentation.Difficulty)
	}

	foundBF := false
	foundRot := false
	foundInv := false
	for _, c := range meta.Documentation.KeyConcepts {
		if c == "Balance Factor [-1, 0, 1]" {
			foundBF = true
		}
		if c == "Single & Double Rotations" {
			foundRot = true
		}
		if c == "Height-Balanced Invariant" {
			foundInv = true
		}
	}
	if !foundBF || !foundRot || !foundInv {
		t.Errorf("missing key concepts in %v", meta.Documentation.KeyConcepts)
	}
}

func TestAVLTraceProperties(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	seenRotationExpl := false
	seenBFInVars := false

	for i, step := range steps {
		state, ok := step.State.(renderers.AVLLessonState)
		if !ok {
			t.Fatalf("step %d state is %T, want renderers.TreeState", i, step.State)
		}

		if step.Variables != nil {
			if _, ok := step.Variables["balanceFactor"]; ok {
				seenBFInVars = true
			} else if _, ok := step.Variables["bf"]; ok {
				seenBFInVars = true
			}
		}

		if step.Action == trace.ActionSwap {
			seenRotationExpl = true
		}

		// Validate coordinates
		nodeMap := make(map[string]renderers.TreeNode)
		for _, n := range state.Nodes {
			nodeMap[n.ID] = n
		}

		for _, n := range state.Nodes {
			if n.ParentID == "" {
				if n.Y != 96 {
					t.Errorf("step %d: root node %s Y=%.1f, want 96", i, n.ID, n.Y)
				}
			} else {
				parent, exists := nodeMap[n.ParentID]
				if !exists {
					t.Errorf("step %d: node %s has unknown parent %s", i, n.ID, n.ParentID)
					continue
				}
				if n.Y <= parent.Y {
					t.Errorf("step %d: child Y (%.1f) not below parent Y (%.1f)", i, n.Y, parent.Y)
				}
			}
		}
	}

	if !seenRotationExpl {
		t.Errorf("trace did not exercise any rotation actions")
	}
	if !seenBFInVars {
		t.Errorf("trace steps variables did not include balance factor ('bf' or 'balanceFactor')")
	}
}

func TestAVLCustomInput(t *testing.T) {
	lesson := Lesson{}
	customInput := []byte(`{"values": [30, 20, 10, 25, 40, 50]}`)

	steps, err := lesson.Build(customInput)
	if err != nil {
		t.Fatalf("Build failed with custom input: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("custom input trace produced 0 steps")
	}

	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("line validation failed on custom input: %v", err)
	}
}

func TestAVLLessonRotationCases(t *testing.T) {
	for name, input := range map[string][]int{"LL": {30,20,10}, "RR": {10,20,30}, "LR": {30,10,20}, "RL": {10,30,20}} {
		t.Run(name, func(t *testing.T) {
			steps := Trace(input)
			if err := catalog.ValidateSourceLineIDs(Lesson{}, steps); err != nil { t.Fatal(err) }
			identities := map[int]string{}
			var rewires []renderers.AVLLessonState
			for i, step := range steps {
				s, ok := step.State.(renderers.AVLLessonState); if !ok { t.Fatalf("frame %d: %T",i,step.State) }
				for j,n := range s.Nodes {
					if old := identities[n.Val]; old != "" && old != n.ID { t.Fatal("identity changed") }; identities[n.Val]=n.ID
					if s.Phase == "height" || s.Phase == "balance" {
						if n.ID != s.ActiveID {continue}; lh,rh := 0,0
						for k,c := range s.Nodes { if c.ID==s.LeftIDs[j] {lh=s.Heights[k]}; if c.ID==s.RightIDs[j] {rh=s.Heights[k]} }
						if s.Heights[j]!=1+max(lh,rh) || s.Balances[j]!=lh-rh { t.Fatalf("bad height/balance: %+v",s) }
					}
				}
				if step.LineID=="L19" || step.LineID=="L23" { if !s.HeightsPending {t.Fatal("rewire heights not pending")}; rewires=append(rewires,s); t.Logf("rewire frame=%d source=%s case=%s part=%d",i,step.LineID,s.Case,s.RotationPart) }
			}
			last:=steps[len(steps)-1].State.(renderers.AVLLessonState)
			if len(identities)!=3 || len(last.Nodes)!=3 {t.Fatal("lost nodes")}
			for i,n:=range last.Nodes { if n.ParentID=="" && (n.Val!=20 || last.Heights[i]!=2) {t.Fatal("wrong root")}; if last.Balances[i] < -1 || last.Balances[i]>1 {t.Fatal("unbalanced")} }
			want:=1; if name=="LR"||name=="RL" {want=2}; if len(rewires)!=want {t.Fatalf("rewires %d want %d",len(rewires),want)}
			if want==2 && (rewires[0].RotationPart!=1 || rewires[1].RotationPart!=2 || reflect.DeepEqual(rewires[0].Nodes,rewires[1].Nodes)) {t.Fatal("double rotation missing intermediate topology")}
		})
	}
}

func TestAVLLessonTransferSubtree(t *testing.T) {
	steps:=Trace([]int{30,20,40,10,25,5}); found:=false
	for _,step:=range steps {s:=step.State.(renderers.AVLLessonState); if step.LineID!="L19" {continue}; for _,n:=range s.Nodes {if n.Val==25 && n.ID==s.TransferID {for _,p:=range s.Nodes {if p.ID==n.ParentID && p.Val==30 {found=true}}}}}
	if !found {t.Fatal("transfer 25 not preserved")}
}

func TestAVLLessonSnapshotsAndDefaults(t *testing.T) {
	if !reflect.DeepEqual(Trace(nil),Trace([]int{10,20,30,40,50,25})) {t.Fatal("default mismatch")}
	input:=[]int{5,5,5}; steps:=Trace(input); input[0]=999
	if steps[0].Variables["values"].([]int)[0]!=5 {t.Fatal("input aliases snapshot")}
	last:=steps[len(steps)-1].State.(renderers.AVLLessonState); if len(last.Nodes)!=1 {t.Fatal("duplicate allocated node")}
	first:=steps[2].State.(renderers.AVLLessonState); last.Nodes[0].Val=999; last.Heights[0]=99
	if first.Nodes[0].Val!=5 || first.Heights[0]!=1 {t.Fatal("snapshots alias")}
}

func TestAVLLessonInputBoundaries(t *testing.T) {
	for _,input:=range []string{`null`,`{"values":null}`,`{"values":[1000001]}`,`[true]`,`{`, `{"values":[-1000001]}`} {if _,err:=(Lesson{}).Build([]byte(input));err==nil {t.Errorf("accepted %s",input)}}
	for _,input:=range []string{`[-1000000,1000000]`,`{}`,`[]`,`{"values":[]}`} {if _,err:=(Lesson{}).Build([]byte(input));err!=nil {t.Errorf("rejected %s: %v",input,err)}}
	raw:=`[`; for i:=0;i<32;i++ {if i>0 {raw+=","};raw+="0"};raw+=`]`
	if _,err:=(Lesson{}).Build([]byte(raw));err!=nil {t.Fatal(err)}
	if _,err:=(Lesson{}).Build([]byte(raw[:len(raw)-1]+`,0]`));err==nil {t.Fatal("accepted 33 values")}
}

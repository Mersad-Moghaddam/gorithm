package kosaraju

// FindSCCs finds strongly connected components using Kosaraju's two-pass DFS.
// ponytail: recursive two-pass DFS; upgrade to iterative DFS if recursion depth exceeds call stack limits on large graphs.
func FindSCCs(nodes []string, edges [][2]string) [][]string {
	if len(nodes) == 0 {
		return nil
	}

	adj := make(map[string][]string, len(nodes))
	revAdj := make(map[string][]string, len(nodes))
	for _, n := range nodes {
		adj[n] = nil
		revAdj[n] = nil
	}
	for _, e := range edges {
		adj[e[0]] = append(adj[e[0]], e[1])
		revAdj[e[1]] = append(revAdj[e[1]], e[0])
	}

	visited := make(map[string]bool, len(nodes))
	finishStack := make([]string, 0, len(nodes))

	var dfs1 func(u string)
	dfs1 = func(u string) {
		visited[u] = true
		for _, v := range adj[u] {
			if !visited[v] {
				dfs1(v)
			}
		}
		finishStack = append(finishStack, u)
	}

	for _, u := range nodes {
		if !visited[u] {
			dfs1(u)
		}
	}

	visitedRev := make(map[string]bool, len(nodes))
	var sccs [][]string

	var dfs2 func(u string, comp *[]string)
	dfs2 = func(u string, comp *[]string) {
		visitedRev[u] = true
		*comp = append(*comp, u)
		for _, v := range revAdj[u] {
			if !visitedRev[v] {
				dfs2(v, comp)
			}
		}
	}

	for len(finishStack) > 0 {
		u := finishStack[len(finishStack)-1]
		finishStack = finishStack[:len(finishStack)-1]
		if !visitedRev[u] {
			var comp []string
			dfs2(u, &comp)
			sccs = append(sccs, comp)
		}
	}

	return sccs
}

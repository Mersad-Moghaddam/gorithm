package dfs

// Traverse performs depth-first search on adj starting at start node and returns visited node IDs in order.
// ponytail: recursive DFS with visited map; upgrade to explicit stack if recursion depth exceeds call stack limits.
func Traverse(adj map[string][]string, start string) []string {
	if start == "" {
		return nil
	}
	visited := make(map[string]bool)
	order := make([]string, 0)

	var dfs func(u string)
	dfs = func(u string) {
		visited[u] = true
		order = append(order, u)
		for _, v := range adj[u] {
			if !visited[v] {
				dfs(v)
			}
		}
	}

	dfs(start)
	return order
}

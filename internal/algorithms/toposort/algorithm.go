package toposort

// Sort performs Kahn's algorithm for topological sorting on a directed graph.
// Returns the topologically sorted node IDs and true if the graph is a valid DAG (no cycles).
// ponytail: map-based adjacency and in-degree tracking; dense integer arrays faster when node count exceeds 10k.
func Sort(nodes []string, edges [][2]string) ([]string, bool) {
	if len(nodes) == 0 {
		return []string{}, true
	}

	inDegree := make(map[string]int, len(nodes))
	adj := make(map[string][]string, len(nodes))
	for _, n := range nodes {
		inDegree[n] = 0
	}
	for _, e := range edges {
		u, v := e[0], e[1]
		adj[u] = append(adj[u], v)
		inDegree[v]++
	}

	var queue []string
	for _, n := range nodes {
		if inDegree[n] == 0 {
			queue = append(queue, n)
		}
	}

	order := make([]string, 0, len(nodes))
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		order = append(order, curr)

		for _, neighbor := range adj[curr] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	return order, len(order) == len(nodes)
}

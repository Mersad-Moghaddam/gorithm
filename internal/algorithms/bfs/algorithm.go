package bfs

// Traverse performs breadth-first search traversal on an adjacency list.
// ponytail: slice-backed queue; ring buffer queue faster when node count exceeds 10k.
func Traverse(adj map[string][]string, start string) []string {
	if start == "" {
		return nil
	}

	visited := make(map[string]bool)
	queue := []string{start}
	visited[start] = true
	var order []string

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		order = append(order, curr)

		for _, neighbor := range adj[curr] {
			if !visited[neighbor] {
				visited[neighbor] = true
				queue = append(queue, neighbor)
			}
		}
	}

	return order
}

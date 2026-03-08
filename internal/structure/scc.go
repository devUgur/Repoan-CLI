package structure

func stronglyConnectedComponents(nodes []string, edges map[string]map[string]int) [][]string {
	index := 0
	stack := []string{}
	onStack := map[string]bool{}
	idx := map[string]int{}
	low := map[string]int{}
	out := [][]string{}

	var visit func(v string)
	visit = func(v string) {
		idx[v] = index
		low[v] = index
		index++
		stack = append(stack, v)
		onStack[v] = true

		for w := range edges[v] {
			if _, ok := idx[w]; !ok {
				visit(w)
				if low[w] < low[v] {
					low[v] = low[w]
				}
			} else if onStack[w] {
				if idx[w] < low[v] {
					low[v] = idx[w]
				}
			}
		}

		if low[v] == idx[v] {
			component := []string{}
			for {
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[w] = false
				component = append(component, w)
				if w == v {
					break
				}
			}
			out = append(out, component)
		}
	}

	for _, n := range nodes {
		if _, ok := idx[n]; !ok {
			visit(n)
		}
	}
	return out
}

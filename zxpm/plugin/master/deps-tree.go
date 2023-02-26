package master

import (
	"fmt"
	"strings"

	"github.com/zostay/dev-tools/zxpm/plugin"
)

type DepNode struct {
	Path  string
	Tasks []plugin.TaskDescription
}

type DepsTree struct {
	root  string
	nodes map[string]*DepNode
	edges map[string][]string
}

func addEdge(edges map[string][]string, from, to string) {
	if _, hasEdgeFrom := edges[from]; hasEdgeFrom {
		edges[from] = append(edges[from], to)
	} else {
		edges[from] = []string{to}
	}
}

func edgeIndex(edges map[string][]string, from, to string) int {
	if _, hasEdgeFrom := edges[from]; !hasEdgeFrom {
		return -1
	}

	for i, toCheck := range edges[from] {
		if toCheck == to {
			return i
		}
	}

	return -1

}

func deleteEdge(edges map[string][]string, from, to string) {
	i := edgeIndex(edges, from, to)
	if i < 0 {
		return
	}

	if len(edges[from]) == 1 {
		delete(edges, from)
		return
	}

	copy(edges[from][i:], edges[from][i+1:])
	edges[from] = edges[from][:len(edges[from])-1]
	return
}

func copyEdges(dst, src map[string][]string) {
	for from, tos := range src {
		dst[from] = make([]string, len(tos))
		copy(dst[from], tos)
	}
}

func NewDepsTree(goal string, tasks []plugin.TaskDescription) *DepsTree {
	nodes := make(map[string]*DepNode, len(tasks))
	edges := make(map[string][]string, len(tasks))
	goalPath := "/" + goal

	establishNode := func(tree string) *DepNode {
		result := nodes[tree]
		if result == nil {
			result = &DepNode{
				Path:  tree,
				Tasks: make([]plugin.TaskDescription, 0, 1),
			}
			nodes[tree] = result
		}
		return result
	}

	for _, task := range tasks {
		parts := strings.Split(task.Name()[1:], "/")

		// skip [0], that's the goal path
		taskNames := parts[1:]

		parent := goalPath
		tree := parent
		for _, part := range taskNames {
			tree += "/" + part

			node := establishNode(tree)
			if tree == task.Name() {
				node.Tasks = append(node.Tasks, task)
			}

			addEdge(edges, tree, parent)
		}

		nodes[task.Name()].Tasks = append(nodes[task.Name()].Tasks, task)

		for _, req := range task.Requires() {
			addEdge(edges, task.Name(), req)
		}
	}

	return &DepsTree{goalPath, nodes, edges}
}

func (d *DepsTree) GroupOrder() ([][]plugin.TaskDescription, error) {
	workEdges := make(map[string][]string, len(d.edges))
	copyEdges(workEdges, d.edges)

	workNodes := make(map[string]*DepNode, len(d.nodes))
	for k, v := range d.nodes {
		workNodes[k] = v
	}

	out := make([][]plugin.TaskDescription, 0, len(d.nodes))
	foundNodes := make([]string, 0, len(d.nodes))
	for len(workEdges) > 0 {
		for from := range workNodes {
			if _, hasAnyEdges := workEdges[from]; !hasAnyEdges {
				foundNodes = append(foundNodes, from)
			}
		}

		// this would mean there's a cycle in the dependency graph
		if len(foundNodes) == 0 {
			return nil, fmt.Errorf("tasks contain a dependency graph, cannot run this goal")
		}

		// once used, we can remove these from the list
		thisPhase := make([]plugin.TaskDescription, 0, len(foundNodes))
		for _, used := range foundNodes {
			thisPhase = append(thisPhase, workNodes[used].Tasks...)
			delete(workNodes, used)
			for from := range workNodes {
				deleteEdge(workEdges, from, used)
			}
		}

		// some phases may have no tasks to execute, but we don't care about
		// those
		if len(thisPhase) > 0 {
			out = append(out, thisPhase)
		}

		// reuse the previous memory
		foundNodes = foundNodes[:0]
	}

	return out, nil
}

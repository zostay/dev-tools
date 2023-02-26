package master

import (
	"sort"
	"strings"

	"github.com/zostay/dev-tools/zxpm/plugin"
)

type TaskGroup struct {
	Tree  string
	Goal  plugin.GoalDescription
	Tasks []plugin.TaskDescription
}

// ExecutionGroups returns a prioritized list of tasks that can be executed in
// stages. Each item in the return slice is a group of tasks that can be
// executed concurrently as a unit. Each group must finish before the next group
// starts.
func (g *TaskGroup) ExecutionGroups() ([][]plugin.TaskDescription, error) {
	tree := NewDepsTree(g.Goal.Name(), g.Tasks)
	return tree.GroupOrder()
}

// SubTasks returns a complete list of sub-tasks for the given task group. This
// will include every possible task group under the given task group.
func (g *TaskGroup) SubTasks() []*TaskGroup {
	// find all the possible tree prefixes, excluding the goal itself
	// i.e., /release/mint/github becomes:
	// * /release/mint
	// * /release/mint/github
	trees := make(map[string]*TaskGroup, len(g.Tasks))
	for _, task := range g.Tasks {
		parts := strings.Split(task.Name()[1:], "/")

		// start with the goal prefix, since we don't want that one
		tree := "/" + parts[0]
		parts = parts[1:]

		for _, part := range parts {
			tree += "/" + part

			if sub, treeExists := trees[tree]; treeExists {
				sub.Tasks = append(sub.Tasks, task)
				continue
			}

			trees[tree] = &TaskGroup{
				Tree:  tree,
				Goal:  g.Goal,
				Tasks: []plugin.TaskDescription{task},
			}
		}
	}

	// convert the map to the output slice
	out := make([]*TaskGroup, 0, len(trees))
	for _, group := range trees {
		out = append(out, group)
	}

	// ensure the output slice is sorted by tree name
	sort.Slice(out, func(i, j int) bool {
		return out[i].Tree < out[j].Tree
	})

	return out
}

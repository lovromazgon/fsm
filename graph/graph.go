package graph

import (
	"fmt"
	"strconv"

	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
	"github.com/lovromazgon/fsm"
)

func DefToGraph[S fsm.State, O any](def fsm.FSM[S, O]) (*cgraph.Graph, error) {
	g := graphviz.New()
	graph, err := g.Graph(graphviz.Directed)
	if err != nil {
		return nil, fmt.Errorf("could not create graph: %w", err)
	}

	nodes := make(map[S]*cgraph.Node)
	for _, state := range def.States() {
		n, err := graph.CreateNode(string(state))
		if err != nil {
			return nil, fmt.Errorf("could not create node %v: %w", state, err)
		}
		nodes[state] = n
	}

	for i, transition := range def.Transitions() {
		name := strconv.Itoa(i) // TODO display nicer name
		e, err := graph.CreateEdge(name, nodes[transition.From], nodes[transition.To])
		if err != nil {
			return nil, fmt.Errorf("could not create edge %v: %w", name, err)
		}
		e.SetLabel(name)
	}

	return graph, nil
}

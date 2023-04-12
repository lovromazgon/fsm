package main

import (
	"log"

	"github.com/goccy/go-graphviz"
	"github.com/lovromazgon/fsm/example"
	"github.com/lovromazgon/fsm/graph"
)

func main() {
	// TODO make this utility work on any state machine (maybe use the same
	//  approach as gomock reflect mode https://github.com/golang/mock/blob/main/mockgen/reflect.go#L45)
	fooGraph, err := graph.DefToGraph[example.FooState, example.FooObservation](&example.FooFSM{})
	if err != nil {
		log.Fatal(err)
	}

	g := graphviz.New()

	if err := g.RenderFilename(fooGraph, graphviz.PNG, "./foograph.png"); err != nil {
		log.Fatal(err)
	}
}

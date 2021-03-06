package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/vetcher/pretty-deps/internal"
)

var (
	source = flag.String("src", "http://localhost:9411/api/v2", "URL of zipkin API")
	groups = flag.String("groups", "", "List of name service prefixes, separated by comma, that should be united to one cluster.")
)

func init() {
	flag.Parse()
}

func main() {
	core := internal.NewCore(*source)
	err := core.UpdateServicesList()
	if err != nil {
		panic(err)
	}
	now := time.Now()
	state := core.GetState(now.AddDate(0, 0, -1), now)
	gg := strings.Split(*groups, ",")
	if gg[0] == "" {
		gg = nil
	}
	dotData, err := internal.StateToGraph(state, internal.VisualizationParams{})
	if err != nil {
		panic(err)
	}
	fmt.Println(string(dotData))
}

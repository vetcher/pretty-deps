package internal

import (
	"fmt"
	"sort"
	"strings"

	"github.com/intel-go/bytebuf"
)

const (
	DarkZone      = "Dark Zone"
	MessageBroker = "Message Broker"
)

func kindToEdgeStyle(kind string) string {
	switch kind {
	case cs:
		return "solid"
	case cp:
		return "dashed"
	case other:
		return "dotted"
	default:
		return "solid"
	}
}

func extractAndRemoveStringFromSlice(ss *[]string, index int) string {
	s := (*ss)[index]
	(*ss)[index] = (*ss)[len(*ss)-1]
	*ss = (*ss)[:len(*ss)-1]
	return s
}

func findPref(ss []string, s string) (j int) {
	for i := range ss {
		if strings.HasPrefix(ss[i], s) {
			return i
		}
	}
	return -1
}

func findLinkPref(ss []Link, s string) (j int) {
	for i := range ss {
		if strings.HasPrefix(ss[i].From, s) {
			return i
		}
	}
	return -1
}

func extractAndRemoveLinkFromSlice(ss *[]Link, index int) Link {
	s := (*ss)[index]
	(*ss)[index] = (*ss)[len(*ss)-1]
	*ss = (*ss)[:len(*ss)-1]
	return s
}

type VisualizationParams struct {
	Prefixes     []string
	RemovePrefix bool
	StylingNodes map[string]StylingParams
}

func StateToGraph(s State, params VisualizationParams) ([]byte, error) {
	r := renderer{b: bytebuf.New()}
	services := s.Services
	links := s.Links
	r.Wln("digraph G {")
	r.Wln("\tnode [shape = box];")
	r.Wln("\tgraph [rankdir = \"LR\", overlap=false];")
	r.Wln("\tedge [dirType = forward];")
	tree := makePrefixTree(services, params.Prefixes)
	t := func(n int) string {
		return strings.Repeat("\t", n)
	}
	walker := func(node *sTree, level int) bool {
		if node == nil {
			return false
		}
		// root node, pass through
		if node.Value == "" {
			return true
		}
		name := node.Value
		// cluster group
		if node.Flag {
			r.Wln(t(level), `subgraph "cluster_`, name, `" {`)
			r.Wln(t(level+1), `label="`, name, `";`)
			r.Wln(t(level+1), parens(
				"node [",
				strings.Join(params.StylingNodes[name].GetPairs("="), ", "),
				"];"),
			)
			return true
		}
		r.Wln(t(level), `"`, name, `";`)
		return true
	}
	walkerDef := func(node *sTree, level int) bool {
		r.Wln(t(level), "}")
		return false
	}

	tree.WalkLevelDefer(walker, walkerDef, 1)

	r.Wln(t(1), `"`, DarkZone, "\" [style=filled,color=black];")
	r.Wln(t(1), `"`, MessageBroker, "\" [style=filled,color=purple];")
	for _, link := range links {
		link.fromToFillEmpty(DarkZone, DarkZone)
		r.Wln(t(1), `"`, link.From, `" -> "`, link.To, `" [style=`, kindToEdgeStyle(link.Kind), "];")
	}
	r.Wln("}")
	return r.b.Bytes(), nil
}

func parens(left, content, right string) string {
	if content == "" {
		return ""
	}
	return left + content + right
}

type StylingParams map[string]string

func (s StylingParams) Add(k, v string) {
	s[k] = v
}

func (s StylingParams) GetPairs(sep string) []string {
	i, x := 0, make([]string, len(s))
	for k, v := range s {
		x[i] = k + sep + v
		i++
	}
	sort.Strings(x)
	return x
}

type renderer struct {
	b bytebuf.Buffer
}

func (r *renderer) W(ss ...string) {
	for i := range ss {
		r.b.WriteString(ss[i])
	}
}

func (r *renderer) Wf(format string, a ...interface{}) {
	r.b.WriteString(fmt.Sprintf(format, a...))
}

func (r *renderer) Wln(ss ...string) {
	for i := range ss {
		r.b.WriteString(ss[i])
	}
	r.b.WriteRune('\n')
}

func (r renderer) String() string {
	return r.b.String()
}

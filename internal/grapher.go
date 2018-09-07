package internal

import (
	"fmt"
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

func StateToGraph(s State, prefs ...string) (string, error) {
	r := renderer{b: bytebuf.New()}
	services := s.Services
	links := s.Links
	r.Wln("digraph {")
	r.Wln("\tnode [shape = box];")
	r.Wln("\tgraph [rankdir = \"LR\", overlap=false];")
	r.Wln("\tedge [dirType = forward];")
	for _, pref := range prefs {
		r.Wln("\tsubgraph \"cluster_", pref, "\" {")
		for index := findPref(services, pref); index != -1; index = findPref(services, pref) {
			service := extractAndRemoveStringFromSlice(&services, index)
			r.Wln("\t\t\"", service, "\";")
		}
		/*
			for index := findLinkPref(links, pref); index != -1; index = findLinkPref(links, pref) {
				link := extractAndRemoveLinkFromSlice(&links, index)
				link.fromToFillEmpty(DarkZone, DarkZone)
				r.Wln("\t\t\"", link.From, "\"->\"", link.To, "\" [style=", kindToEdgeStyle(link.Kind), "];")
			}
		*/
		r.Wln("\t}")
	}

	for i := range services {
		r.Wln("\t\"", s.Services[i], "\";")
	}
	r.Wln("\t\"", DarkZone, "\" [style=filled,color=black];")
	r.Wln("\t\"", MessageBroker, "\" [style=filled,color=purple];")
	for _, link := range links {
		link.fromToFillEmpty(DarkZone, DarkZone)
		r.Wln("\t\"", link.From, "\"->\"", link.To, "\" [style=", kindToEdgeStyle(link.Kind), "];")
	}
	r.Wln("}")
	return r.String(), nil
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

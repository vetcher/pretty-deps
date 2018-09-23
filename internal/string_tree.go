package internal

import (
	"sort"
	"strings"
)

func makePrefixTree(input []string, prefixes []string) sTree {
	sort.Strings(prefixes)
	sort.Strings(input)
	root := sTree{Value: ""}
	for i := range prefixes {
		root.add(prefixes[i], true, strings.HasPrefix)
	}
	for i := range input {
		root.add(input[i], false, strings.HasPrefix)
	}
	return root
}

type sTree struct {
	Value    string
	Flag     bool
	Children []*sTree
}

func (s *sTree) add(value string, flag bool, selectorFunc func(string, string) bool) {
	for _, c := range s.Children {
		if c != nil && c.Flag && selectorFunc(value, c.Value) {
			c.add(value, flag, selectorFunc)
			return
		}
	}
	s.Children = append(s.Children, &sTree{Value: value, Flag: flag})
}

func (s *sTree) Walk(fn func(*sTree) bool) {
	if !fn(s) {
		return
	}
	for i := range s.Children {
		s.Children[i].Walk(fn)
	}
}

func (s *sTree) WalkLevel(fn func(*sTree, int) bool, level int) {
	if !fn(s, level) {
		return
	}
	level += 1
	for i := range s.Children {
		s.Children[i].WalkLevel(fn, level)
	}
}

func (s *sTree) WalkLevelDefer(fn, def func(*sTree, int) bool, level int) {
	if !fn(s, level) {
		return
	}
	level += 1
	for i := range s.Children {
		s.Children[i].WalkLevel(fn, level)
	}
	def(s, level)
}

func (s sTree) GoString() string {
	// code below is highly not optimized // todo
	x := make([]string, len(s.Children))
	for i := range s.Children {
		if s.Children[i] != nil {
			subTree := strings.Split(s.Children[i].GoString(), "\n")
			for j := range subTree {
				subTree[j] = "\t" + subTree[j]
			}
			x[i] = strings.Join(subTree, "\n")
		}
	}
	for i := 0; i < len(x); {
		if x[i] == "" {
			x[i], x[len(x)-1] = x[len(x)-1], x[i]
			x = x[:len(x)-1]
		}
		i++
	}
	return strings.Join(append([]string{s.Value + iss(s.Flag)}, x...), "\n")
}

func iss(f bool) string {
	if f {
		return " >"
	}
	return ""
}

package internal

import (
	"fmt"
	"strconv"
	"testing"
)

func Test_makePrefixTree(t *testing.T) {
	input := []string{"a", "b", "c", "aa", "abfds", "abc", "abd", "ca", "cb", "cd", "ce"}
	prefixes := []string{"ab", "a"}
	tree := makePrefixTree(input, prefixes)
	fmt.Println("root:", tree.GoString())
}

func TestSTree_Walk(t *testing.T) {
	input := []string{"a", "b", "c", "aa", "abfds", "abc", "abd", "ca", "cb", "cd", "ce"}
	prefixes := []string{"ab", "a", "c"}
	tree := makePrefixTree(input, prefixes)
	i := 0
	x := func(node *sTree, level int) bool {
		if node == nil {
			return false
		}
		node.Value = strconv.Itoa(level) + node.Value + strconv.Itoa(i)
		i++
		return true
	}
	tree.WalkLevel(x, 0)
	fmt.Println(tree.GoString())
}

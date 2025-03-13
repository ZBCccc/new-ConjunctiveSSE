package ggmtree

import (
	util "github.com/ZBCccc/Aura/Util"
	"math"
)

// GGMTree represents a tree structure in the GGM scheme.
type GGMTree struct {
	Level int
}

// NewGGMTree creates a new GGMTree with the given number of nodes.
func NewGGMTree(numNode int64) *GGMTree {
	return &GGMTree{
		Level: int(math.Ceil(math.Log2(float64(numNode)))),
	}
}

// DeriveKeyFromTree derives a key from the tree.
func DeriveKeyFromTree(currentKey []byte, offset uint, startLevel, targetLevel int) {
	if startLevel == targetLevel {
		return
	}
	for k := startLevel; k > targetLevel; k-- {
		kBit := (offset & (1 << (k - 1))) >> (k - 1)
		nextKey := util.KeyDerivation([]byte{byte(kBit)}, currentKey)
		copy(currentKey, nextKey)
	}
}

// MinCoverage calculates the minimum coverage of nodes.
func MinCoverage(nodeList []GGMNode) []GGMNode {
	nextLevelNode := make([]GGMNode, 0, len(nodeList))
	for i := 0; i < len(nodeList); i++ {
		node1 := nodeList[i]
		if i+1 == len(nodeList) {
			nextLevelNode = append(nextLevelNode, node1)
		} else {
			node2 := nodeList[i+1]
			if (node1.Index>>1 == node2.Index>>1) && (node1.Level == node2.Level) {
				nextLevelNode = append(nextLevelNode, GGMNode{Index: node1.Index >> 1, Level: node1.Level - 1})
				i++
			} else {
				nextLevelNode = append(nextLevelNode, node1)
			}
		}
	}
	if len(nextLevelNode) == len(nodeList) || len(nextLevelNode) == 0 {
		return nodeList
	}
	return MinCoverage(nextLevelNode)
}

// GetLevel returns the level of the tree.
func (tree *GGMTree) GetLevel() int {
	return tree.Level
}

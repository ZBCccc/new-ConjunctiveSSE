package ggmtree

import "crypto/aes"

// GGMNode represents a node in the GGM tree.
type GGMNode struct {
	Index int
	Level int
	Key   [aes.BlockSize]byte
}

// NewGGMNode creates a new GGMNode with the given index and level.
func NewGGMNode(index int, level int) *GGMNode {
	return &GGMNode{
		Index: index,
		Level: level,
	}
}

// NewGGMNodeWithKey creates a new GGMNode with the given index, level, and key.
func NewGGMNodeWithKey(index int, level int, key []byte) *GGMNode {
	node := &GGMNode{
		Index: index,
		Level: level,
	}
	copy(node.Key[:], key)
	return node
}
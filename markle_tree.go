package main

import (
	"crypto/sha256"
)

// MarkleTree ...
type MarkleTree struct {
	RootNode *MarkleNode
}

// MarkleNode ...
type MarkleNode struct {
	Left  *MarkleNode
	Right *MarkleNode
	Data  []byte
}

// NewMarkleNode ...
func NewMarkleNode(left, right *MarkleNode, data []byte) *MarkleNode {
	mNode := MarkleNode{}

	if left == nil && right == nil {
		hash := sha256.Sum256(data)
		mNode.Data = hash[:]
	} else {
		prevHashes := append(left.Data, right.Data...)
		hash := sha256.Sum256(prevHashes)
		mNode.Data = hash[:]
	}

	mNode.Left = left
	mNode.Right = right

	return &mNode
}

// NewMarkleTree ...
func NewMarkleTree(data [][]byte) *MarkleTree {
	var nodes []MarkleNode
	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}

	for _, item := range data {
		node := NewMarkleNode(nil, nil, item)
		nodes = append(nodes, *node)
	}

	for i := 0; i < len(data)/2; i++ {
		var newLevel []MarkleNode
		for j := 0; j < len(nodes); j += 2 {
			node := NewMarkleNode(&nodes[j], &nodes[j+1], nil)
			newLevel = append(newLevel, *node)
		}
		nodes = newLevel
	}

	tree := MarkleTree{&nodes[0]}

	return &tree
}

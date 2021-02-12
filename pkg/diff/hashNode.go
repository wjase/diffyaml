package diff

import (
	"crypto/sha1"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

// HashedNodes custom hash node collection
type HashedNodes []*HashedNode

// Remove remove from collection
func (h HashedNodes) Remove(toRemoveIndex int) HashedNodes {
	return append(h[:toRemoveIndex], h[toRemoveIndex+1:]...)
}

// Add add item to collection
func (h HashedNodes) Add(node *HashedNode) HashedNodes {
	return append(h, node)
}

// Find index of the keyed item or -1 if not there
func (h HashedNodes) Find(key string) (int, *HashedNode) {
	for index, element := range h {
		if element.Key == key {
			return index, element
		}
	}
	return -1, nil
}

// HashedNode wraps a node with a hash of the child nodes
type HashedNode struct {
	Parent   *HashedNode
	Node     *yaml.Node
	Key      string
	Hash     string
	Children HashedNodes
}

// IsScalar returns true if the Node is a scalar
func (h HashedNode) IsScalar() bool {
	return h.Node.Kind == yaml.ScalarNode
}

//HashNode calculate hash for node and children and build HashedNode structure for comparison
func HashNode(node *yaml.Node) *HashedNode {
	hashedNode := HashedNode{Node: node, Children: []*HashedNode{}, Parent: nil}

	switch node.Kind {
	case yaml.DocumentNode:
		buildChildren(&hashedNode)
		hashChildren(&hashedNode)
	case yaml.SequenceNode:
		buildChildren(&hashedNode)
		hashChildren(&hashedNode)
	case yaml.MappingNode:
		buildChildren(&hashedNode)
		hashChildren(&hashedNode)
	case yaml.ScalarNode:
		h := sha1.New()
		hashedNode.Hash = string(h.Sum([]byte(node.Value)))
	case yaml.AliasNode:
		// don't handle these yet
	}
	return &hashedNode
}

func hashChildren(hashedNode *HashedNode) {
	h := sha1.New()
	for _, eachChild := range hashedNode.Children {
		io.WriteString(h, string(eachChild.Hash))
	}
	hashedNode.Hash = string(h.Sum(nil))
}

func buildChildren(hashedNode *HashedNode) {
	var previousKeyNode string
	nodeKind := hashedNode.Node.Kind
	for ind, eachChild := range hashedNode.Node.Content {
		key := fmt.Sprintf("[%d]", ind)

		if nodeKind == yaml.MappingNode {
			if ind%2 == 0 {
				previousKeyNode = eachChild.Value
				continue
			} else {
				key = previousKeyNode
			}
		}
		// if eachChild.Kind == yaml.MappingNode && nodeKind == yaml.SequenceNode {
		// 	childCount := len(eachChild.Content)
		// 	if childCount > 2 {
		// 		key = eachChild.Content[0].Value
		// 		if len(eachChild.Content[1].Value) > 0 {
		// 			key = key + ":" + eachChild.Content[1].Value
		// 		}
		// 	}
		// }
		childHashedNode := HashNode(eachChild)
		childHashedNode.Key = key
		hashedNode.Children = append(hashedNode.Children, childHashedNode)
		childHashedNode.Parent = hashedNode
	}
}

// GetPath returns the nodes up to the root
func (h *HashedNode) GetPath() HashedNodes {
	if h.Parent == nil || h.Parent.Parent == nil {
		return HashedNodes{}
	}
	return append(h.Parent.GetPath(), h)
}

func (h HashedNodes) String() string {
	bld := strings.Builder{}
	bld.WriteString("doc.")
	finalDotIndex := len(h) - 1
	for ind, eachPart := range h {
		if ind <= finalDotIndex && ind > 0 {
			bld.WriteString(".")
		}
		bld.WriteString(eachPart.Key)
	}
	return bld.String()
}

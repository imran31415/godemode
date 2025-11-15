package graph

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dgraph-io/badger/v4"
)

// Node represents a node in the knowledge graph
type Node struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`       // issue, solution, ticket
	Data       map[string]interface{} `json:"data"`
	Neighbors  map[string][]string    `json:"neighbors"`  // edge_type -> []node_ids
}

// Edge represents a relationship between nodes
type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"`  // similar_to, solved_by, caused_by, etc.
}

// KnowledgeGraph wraps BadgerDB for graph operations
type KnowledgeGraph struct {
	db *badger.DB
	path string
}

// NewKnowledgeGraph creates or opens a knowledge graph
func NewKnowledgeGraph(path string) (*KnowledgeGraph, error) {
	opts := badger.DefaultOptions(path)
	opts.Logger = nil  // Disable logging

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open graph database: %w", err)
	}

	return &KnowledgeGraph{db: db, path: path}, nil
}

// AddNode adds a node to the graph
func (kg *KnowledgeGraph) AddNode(node *Node) error {
	if node.Neighbors == nil {
		node.Neighbors = make(map[string][]string)
	}

	data, err := json.Marshal(node)
	if err != nil {
		return fmt.Errorf("failed to marshal node: %w", err)
	}

	err = kg.db.Update(func(txn *badger.Txn) error {
		key := []byte("node:" + node.ID)
		return txn.Set(key, data)
	})

	if err != nil {
		return fmt.Errorf("failed to add node: %w", err)
	}

	return nil
}

// GetNode retrieves a node by ID
func (kg *KnowledgeGraph) GetNode(id string) (*Node, error) {
	var node Node

	err := kg.db.View(func(txn *badger.Txn) error {
		key := []byte("node:" + id)
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &node)
		})
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, fmt.Errorf("node not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	return &node, nil
}

// AddEdge creates a relationship between two nodes
func (kg *KnowledgeGraph) AddEdge(from, to, edgeType string) error {
	// Get the from node
	fromNode, err := kg.GetNode(from)
	if err != nil {
		return fmt.Errorf("from node not found: %w", err)
	}

	// Add to neighbors
	if fromNode.Neighbors == nil {
		fromNode.Neighbors = make(map[string][]string)
	}

	// Add edge if it doesn't exist
	neighbors := fromNode.Neighbors[edgeType]
	exists := false
	for _, n := range neighbors {
		if n == to {
			exists = true
			break
		}
	}

	if !exists {
		fromNode.Neighbors[edgeType] = append(fromNode.Neighbors[edgeType], to)
	}

	// Update the node
	return kg.AddNode(fromNode)
}

// GetNeighbors returns all neighbors of a node for a specific edge type
func (kg *KnowledgeGraph) GetNeighbors(nodeID, edgeType string) ([]string, error) {
	node, err := kg.GetNode(nodeID)
	if err != nil {
		return nil, err
	}

	neighbors := node.Neighbors[edgeType]
	if neighbors == nil {
		return []string{}, nil
	}

	return neighbors, nil
}

// FindSimilar finds nodes similar to the given description
// This is a simplified implementation using keyword matching
func (kg *KnowledgeGraph) FindSimilar(description string, nodeType string, topK int) ([]*Node, error) {
	keywords := strings.Fields(strings.ToLower(description))

	var candidates []*Node

	err := kg.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("node:")

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()

			err := item.Value(func(val []byte) error {
				var node Node
				if err := json.Unmarshal(val, &node); err != nil {
					return err
				}

				// Filter by type if specified
				if nodeType != "" && node.Type != nodeType {
					return nil
				}

				// Calculate similarity score (simple keyword matching)
				score := 0
				nodeText := strings.ToLower(fmt.Sprintf("%v", node.Data))

				for _, keyword := range keywords {
					if strings.Contains(nodeText, keyword) {
						score++
					}
				}

				if score > 0 {
					candidates = append(candidates, &node)
				}

				return nil
			})

			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to find similar nodes: %w", err)
	}

	// Return top K candidates (simplified - just return first K)
	if len(candidates) > topK {
		candidates = candidates[:topK]
	}

	return candidates, nil
}

// ListNodes returns all nodes of a specific type
func (kg *KnowledgeGraph) ListNodes(nodeType string) ([]*Node, error) {
	var nodes []*Node

	err := kg.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("node:")

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()

			err := item.Value(func(val []byte) error {
				var node Node
				if err := json.Unmarshal(val, &node); err != nil {
					return err
				}

				if nodeType == "" || node.Type == nodeType {
					nodes = append(nodes, &node)
				}

				return nil
			})

			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	return nodes, nil
}

// CountNodes returns the total number of nodes
func (kg *KnowledgeGraph) CountNodes() (int, error) {
	count := 0

	err := kg.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("node:")

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			count++
		}

		return nil
	})

	return count, err
}

// Reset clears all data from the graph
func (kg *KnowledgeGraph) Reset() error {
	return kg.db.DropAll()
}

// Close closes the graph database
func (kg *KnowledgeGraph) Close() error {
	return kg.db.Close()
}

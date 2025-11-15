package support

import (
	"fmt"

	"github.com/imran31415/godemode/benchmark/systems/graph"
	"github.com/imran31415/godemode/benchmark/tools"
)

// GraphTools provides knowledge graph tools
type GraphTools struct {
	kg *graph.KnowledgeGraph
}

// NewGraphTools creates graph tools
func NewGraphTools(kg *graph.KnowledgeGraph) *GraphTools {
	return &GraphTools{kg: kg}
}

// RegisterTools registers all graph tools with the registry
func (gt *GraphTools) RegisterTools(registry *tools.Registry) error {
	// FindSimilarIssues tool
	err := registry.Register(&tools.ToolInfo{
		Name:        "findSimilarIssues",
		Description: "Find similar issues in the knowledge graph",
		Parameters: []tools.ParamInfo{
			{Name: "description", Type: "string", Required: true},
			{Name: "topK", Type: "int", Required: false},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			description, ok := args["description"]
			if !ok {
				return nil, fmt.Errorf("findSimilarIssues requires description parameter")
			}

			descriptionStr := fmt.Sprintf("%v", description)

			topK := 5 // default
			if topKVal, ok := args["topK"]; ok {
				if k, ok := topKVal.(int); ok {
					topK = k
				}
			}

			return gt.kg.FindSimilar(descriptionStr, "issue", topK)
		},
	})
	if err != nil {
		return err
	}

	// LinkIssueInGraph tool
	err = registry.Register(&tools.ToolInfo{
		Name:        "linkIssueInGraph",
		Description: "Link two issues in the knowledge graph",
		Parameters: []tools.ParamInfo{
			{Name: "ticketID", Type: "string", Required: true},
			{Name: "relatedID", Type: "string", Required: true},
			{Name: "relationshipType", Type: "string", Required: false},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			ticketID, ok := args["ticketID"]
			if !ok {
				return nil, fmt.Errorf("linkIssueInGraph requires ticketID parameter")
			}

			ticketIDStr := fmt.Sprintf("%v", ticketID)

			relatedID, ok := args["relatedID"]
			if !ok {
				return nil, fmt.Errorf("linkIssueInGraph requires relatedID parameter")
			}

			relatedIDStr := fmt.Sprintf("%v", relatedID)

			relationshipType := "similar_to"
			if relType, ok := args["relationshipType"]; ok {
				relationshipType = fmt.Sprintf("%v", relType)
			}

			err := gt.kg.AddEdge(ticketIDStr, relatedIDStr, relationshipType)
			return err == nil, err
		},
	})
	if err != nil {
		return err
	}

	// AddIssueNode tool
	err = registry.Register(&tools.ToolInfo{
		Name:        "addIssueNode",
		Description: "Add a new issue node to the knowledge graph",
		Parameters: []tools.ParamInfo{
			{Name: "issueID", Type: "string", Required: true},
			{Name: "data", Type: "map", Required: true},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			issueID, ok := args["issueID"]
			if !ok {
				return nil, fmt.Errorf("addIssueNode requires issueID parameter")
			}

			issueIDStr := fmt.Sprintf("%v", issueID)

			data, ok := args["data"].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("data must be a map")
			}

			node := &graph.Node{
				ID:   issueIDStr,
				Type: "issue",
				Data: data,
			}

			err := gt.kg.AddNode(node)
			return err == nil, err
		},
	})
	if err != nil {
		return err
	}

	// GetIssueSolution tool
	err = registry.Register(&tools.ToolInfo{
		Name:        "getIssueSolution",
		Description: "Get the solution for an issue from the knowledge graph",
		Parameters: []tools.ParamInfo{
			{Name: "issueID", Type: "string", Required: true},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			issueID, ok := args["issueID"]
			if !ok {
				return nil, fmt.Errorf("getIssueSolution requires issueID parameter")
			}

			issueIDStr := fmt.Sprintf("%v", issueID)

			// Get neighbors with "solved_by" relationship
			neighbors, err := gt.kg.GetNeighbors(issueIDStr, "solved_by")
			if err != nil {
				return nil, err
			}

			if len(neighbors) == 0 {
				return nil, nil // No solution found
			}

			// Return the first solution
			return gt.kg.GetNode(neighbors[0])
		},
	})

	return err
}

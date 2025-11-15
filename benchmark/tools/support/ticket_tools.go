package support

import (
	"fmt"

	"github.com/imran31415/godemode/benchmark/systems/database"
	"github.com/imran31415/godemode/benchmark/tools"
)

// TicketTools provides ticket-related tools
type TicketTools struct {
	db *database.SQLiteDB
}

// NewTicketTools creates ticket tools
func NewTicketTools(db *database.SQLiteDB) *TicketTools {
	return &TicketTools{db: db}
}

// RegisterTools registers all ticket tools with the registry
func (tt *TicketTools) RegisterTools(registry *tools.Registry) error {
	// CreateTicket tool
	err := registry.Register(&tools.ToolInfo{
		Name:        "createTicket",
		Description: "Create a new support ticket",
		Parameters: []tools.ParamInfo{
			{Name: "customerID", Type: "string", Required: true},
			{Name: "subject", Type: "string", Required: true},
			{Name: "description", Type: "string", Required: true},
			{Name: "priority", Type: "int", Required: false},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			customerID, ok := args["customerID"]
			if !ok {
				return nil, fmt.Errorf("createTicket requires customerID parameter")
			}

			customerIDStr := fmt.Sprintf("%v", customerID)

			subject, ok := args["subject"]
			if !ok {
				return nil, fmt.Errorf("createTicket requires subject parameter")
			}

			subjectStr := fmt.Sprintf("%v", subject)

			description, ok := args["description"]
			if !ok {
				return nil, fmt.Errorf("createTicket requires description parameter")
			}

			descriptionStr := fmt.Sprintf("%v", description)

			priority := 3 // default
			if p, ok := args["priority"]; ok {
				if pInt, ok := p.(int); ok {
					priority = pInt
				}
			}

			ticket := &database.Ticket{
				CustomerID:  customerIDStr,
				Subject:     subjectStr,
				Description: descriptionStr,
				Priority:    priority,
				Status:      "new",
			}

			err := tt.db.CreateTicket(ticket)
			if err != nil {
				return nil, err
			}

			return ticket.ID, nil
		},
	})
	if err != nil {
		return err
	}

	// UpdateTicket tool
	err = registry.Register(&tools.ToolInfo{
		Name:        "updateTicket",
		Description: "Update a ticket's fields",
		Parameters: []tools.ParamInfo{
			{Name: "ticketID", Type: "string", Required: true},
			{Name: "updates", Type: "map", Required: true},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			ticketID, ok := args["ticketID"]
			if !ok {
				return nil, fmt.Errorf("updateTicket requires ticketID parameter")
			}

			ticketIDStr := fmt.Sprintf("%v", ticketID)

			updates, ok := args["updates"].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("updates must be a map")
			}

			err := tt.db.UpdateTicket(ticketIDStr, updates)
			return err == nil, err
		},
	})
	if err != nil {
		return err
	}

	// GetTicket tool
	err = registry.Register(&tools.ToolInfo{
		Name:        "getTicket",
		Description: "Get a ticket by ID",
		Parameters: []tools.ParamInfo{
			{Name: "ticketID", Type: "string", Required: true},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			ticketID, ok := args["ticketID"]
			if !ok {
				return nil, fmt.Errorf("getTicket requires ticketID parameter")
			}

			ticketIDStr := fmt.Sprintf("%v", ticketID)
			return tt.db.GetTicket(ticketIDStr)
		},
	})
	if err != nil {
		return err
	}

	// QueryTickets tool
	err = registry.Register(&tools.ToolInfo{
		Name:        "queryTickets",
		Description: "Query tickets with filters",
		Parameters: []tools.ParamInfo{
			{Name: "filters", Type: "map", Required: false},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			filters := make(map[string]interface{})

			if f, ok := args["filters"].(map[string]interface{}); ok {
				filters = f
			}

			return tt.db.QueryTickets(filters)
		},
	})
	if err != nil {
		return err
	}

	// SetPriority tool
	err = registry.Register(&tools.ToolInfo{
		Name:        "setPriority",
		Description: "Set a ticket's priority (1-5)",
		Parameters: []tools.ParamInfo{
			{Name: "ticketID", Type: "string", Required: true},
			{Name: "priority", Type: "int", Required: true},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			ticketID, ok := args["ticketID"]
			if !ok {
				return nil, fmt.Errorf("setPriority requires ticketID parameter")
			}

			ticketIDStr := fmt.Sprintf("%v", ticketID)

			priority, ok := args["priority"]
			if !ok {
				return nil, fmt.Errorf("setPriority requires priority parameter")
			}

			priorityInt, ok := priority.(int)
			if !ok {
				return nil, fmt.Errorf("priority must be an integer")
			}

			updates := map[string]interface{}{"priority": priorityInt}
			err := tt.db.UpdateTicket(ticketIDStr, updates)
			return err == nil, err
		},
	})
	if err != nil {
		return err
	}

	// AssignTicket tool
	err = registry.Register(&tools.ToolInfo{
		Name:        "assignTicket",
		Description: "Assign a ticket to an agent",
		Parameters: []tools.ParamInfo{
			{Name: "ticketID", Type: "string", Required: true},
			{Name: "agentID", Type: "string", Required: true},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			ticketID, ok := args["ticketID"]
			if !ok {
				return nil, fmt.Errorf("assignTicket requires ticketID parameter")
			}

			ticketIDStr := fmt.Sprintf("%v", ticketID)

			agentID, ok := args["agentID"]
			if !ok {
				return nil, fmt.Errorf("assignTicket requires agentID parameter")
			}

			agentIDStr := fmt.Sprintf("%v", agentID)

			updates := map[string]interface{}{
				"assigned_to": agentIDStr,
				"status":      "assigned",
			}
			err := tt.db.UpdateTicket(ticketIDStr, updates)
			return err == nil, err
		},
	})

	return err
}

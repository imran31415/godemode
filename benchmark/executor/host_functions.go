package executor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/imran31415/godemode/benchmark/scenarios"
	"github.com/imran31415/godemode/benchmark/systems/database"
	"github.com/imran31415/godemode/benchmark/systems/graph"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

// HostFunctions provides system access to WASM code
type HostFunctions struct {
	env *scenarios.TestEnvironment
}

// NewHostFunctions creates host functions for a test environment
func NewHostFunctions(env *scenarios.TestEnvironment) *HostFunctions {
	return &HostFunctions{env: env}
}

// RegisterHostFunctions registers all host functions with the WASM runtime
func (hf *HostFunctions) RegisterHostFunctions(ctx context.Context, r wazero.Runtime) (api.Module, error) {
	// Create the "env" module for host functions
	builder := r.NewHostModuleBuilder("env")

	// Register ticket operations
	builder.NewFunctionBuilder().
		WithFunc(hf.createTicket).
		Export("create_ticket")

	builder.NewFunctionBuilder().
		WithFunc(hf.getTicketCount).
		Export("get_ticket_count")

	// Register email operations
	builder.NewFunctionBuilder().
		WithFunc(hf.readEmail).
		Export("read_email")

	builder.NewFunctionBuilder().
		WithFunc(hf.sendEmail).
		Export("send_email")

	// Register log operations
	builder.NewFunctionBuilder().
		WithFunc(hf.searchLogs).
		Export("search_logs")

	// Register graph operations
	builder.NewFunctionBuilder().
		WithFunc(hf.linkToGraph).
		Export("link_to_graph")

	// Instantiate the host module
	return builder.Instantiate(ctx)
}

// createTicket creates a new support ticket
func (hf *HostFunctions) createTicket(ctx context.Context, m api.Module, priority uint32, tagsPtr, tagsLen uint32) uint32 {
	// Read tags from memory
	tags := []string{}
	if tagsLen > 0 {
		tagsData, ok := m.Memory().Read(tagsPtr, tagsLen)
		if ok {
			json.Unmarshal(tagsData, &tags)
		}
	}

	ticket := &database.Ticket{
		CustomerID:  "customer123",
		Subject:     "Support Ticket",
		Description: "Created from WASM code",
		Priority:    int(priority),
		Status:      "open",
		Tags:        tags,
	}

	err := hf.env.Database.CreateTicket(ticket)
	if err != nil {
		return 0 // failure
	}

	return 1 // success
}

// getTicketCount returns the number of tickets
func (hf *HostFunctions) getTicketCount(ctx context.Context, m api.Module) uint32 {
	count, err := hf.env.Database.CountTickets()
	if err != nil {
		return 0
	}
	return uint32(count)
}

// readEmail reads an email
func (hf *HostFunctions) readEmail(ctx context.Context, m api.Module, emailIDPtr, emailIDLen uint32) uint32 {
	emailIDData, ok := m.Memory().Read(emailIDPtr, emailIDLen)
	if !ok {
		return 0
	}

	emailID := string(emailIDData)
	_, err := hf.env.EmailSystem.ReadEmail(emailID)
	if err != nil {
		return 0
	}

	return 1
}

// sendEmail sends an email
func (hf *HostFunctions) sendEmail(ctx context.Context, m api.Module, toPtr, toLen, subjectPtr, subjectLen uint32) uint32 {
	toData, ok := m.Memory().Read(toPtr, toLen)
	if !ok {
		return 0
	}

	subjectData, ok := m.Memory().Read(subjectPtr, subjectLen)
	if !ok {
		return 0
	}

	to := string(toData)
	subject := string(subjectData)

	_, err := hf.env.EmailSystem.WriteEmail(to, subject, "Confirmation email")
	if err != nil {
		return 0
	}

	return 1
}

// searchLogs searches application logs
func (hf *HostFunctions) searchLogs(ctx context.Context, m api.Module, patternPtr, patternLen uint32) uint32 {
	patternData, ok := m.Memory().Read(patternPtr, patternLen)
	if !ok {
		return 0
	}

	pattern := string(patternData)
	entries, err := hf.env.LogSystem.SearchLogs(pattern, 0)
	if err != nil {
		return 0
	}

	if len(entries) > 0 {
		return 1
	}

	return 0
}

// linkToGraph links a ticket to the knowledge graph
func (hf *HostFunctions) linkToGraph(ctx context.Context, m api.Module, ticketIDPtr, ticketIDLen, issueIDPtr, issueIDLen uint32) uint32 {
	ticketIDData, ok := m.Memory().Read(ticketIDPtr, ticketIDLen)
	if !ok {
		return 0
	}

	issueIDData, ok := m.Memory().Read(issueIDPtr, issueIDLen)
	if !ok {
		return 0
	}

	ticketID := string(ticketIDData)
	issueID := string(issueIDData)

	// Add ticket node
	ticketNode := &graph.Node{
		ID:   ticketID,
		Type: "ticket",
		Data: map[string]interface{}{
			"id": ticketID,
		},
	}
	hf.env.Graph.AddNode(ticketNode)

	// Link to issue
	err := hf.env.Graph.AddEdge(ticketID, issueID, "similar_to")
	if err != nil {
		fmt.Printf("Failed to link: %v\n", err)
		return 0
	}

	return 1
}

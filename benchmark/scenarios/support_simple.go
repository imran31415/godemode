package scenarios

import (
	"fmt"
	"time"

	"github.com/imran31415/godemode/benchmark/systems/database"
	"github.com/imran31415/godemode/benchmark/systems/email"
	"github.com/imran31415/godemode/benchmark/systems/filesystem"
	"github.com/imran31415/godemode/benchmark/systems/graph"
	"github.com/imran31415/godemode/benchmark/systems/security"
)

// Task represents a benchmark task
type Task struct {
	Name              string
	Description       string
	Complexity        string // simple, medium, complex
	ExpectedOps       int
	SetupFunc         func(*TestEnvironment) error
	VerificationFunc  func(*TestEnvironment) (bool, []string)
}

// TestEnvironment contains all systems for testing
type TestEnvironment struct {
	EmailSystem     *email.EmailSystem
	Database        *database.SQLiteDB
	Graph           *graph.KnowledgeGraph
	LogSystem       *filesystem.LogSystem
	ConfigSystem    *filesystem.ConfigSystem
	SecurityMonitor *security.SecurityMonitor
}

// SupportScenario contains all support-related tasks
type SupportScenario struct {
	Name  string
	Tasks []Task
}

// NewSupportScenario creates the support ticket scenario
func NewSupportScenario() *SupportScenario {
	return &SupportScenario{
		Name: "IT Support Triage & Resolution",
		Tasks: []Task{
			createSimpleTask(),
			createMediumTask(),
			createComplexTask(),
		},
	}
}

// createSimpleTask: Email to Ticket creation
func createSimpleTask() Task {
	return Task{
		Name:        "email-to-ticket",
		Description: "Read email ID 'support_001' (a bug report about login failure), create a ticket with high priority (3-4) since it's a bug, and send confirmation",
		Complexity:  "simple",
		ExpectedOps: 3, // readEmail, createTicket, sendEmail

		SetupFunc: func(env *TestEnvironment) error {
			// Email file support_001.eml already exists in fixtures
			// No setup needed - using static fixture email
			return nil
		},

		VerificationFunc: func(env *TestEnvironment) (bool, []string) {
			var errors []string
			success := true

			// Check: 1 ticket created
			count, err := env.Database.CountTickets()
			if err != nil {
				errors = append(errors, fmt.Sprintf("Failed to count tickets: %v", err))
				return false, errors
			}
			if count != 1 {
				errors = append(errors, fmt.Sprintf("Expected 1 ticket, got %d", count))
				success = false
			}

			// Check: Ticket has priority set (3 or 4 for bug)
			tickets, err := env.Database.QueryTickets(map[string]interface{}{})
			if err != nil || len(tickets) == 0 {
				errors = append(errors, "No tickets found")
				return false, errors
			}

			ticket := tickets[0]
			if ticket.Priority < 3 || ticket.Priority > 4 {
				errors = append(errors, fmt.Sprintf("Expected priority 3-4, got %d", ticket.Priority))
				success = false
			}

			// Check: Confirmation email sent
			// Note: In real implementation, would check outbox

			if success {
				errors = append(errors, "✓ All checks passed")
			}

			return success, errors
		},
	}
}

// createMediumTask: Multi-system investigation
func createMediumTask() Task {
	return Task{
		Name:        "investigate-with-logs",
		Description: "Read email ID 'error_report_001' (about ERR-500-XYZ OutOfMemory error), search logs for 'ERR-500-XYZ', find similar issues in knowledge graph, create high priority ticket (priority 4-5) with tags=['memory', 'OutOfMemory'] and description including log analysis, link ticket to similar issues",
		Complexity:  "medium",
		ExpectedOps: 8, // readEmail, searchLogs, findSimilar, createTicket, linkIssue x2, sendEmail

		SetupFunc: func(env *TestEnvironment) error {
			// Create test email with error code
			emailContent := `From: admin@company.com
To: support@company.com
Subject: Error 500 on file upload
Date: ` + time.Now().Format(time.RFC1123Z) + `

Getting Error 500 when trying to upload files larger than 10MB.

Error ID: ERR-500-XYZ

This is blocking our entire team!`

			// Write log entry with this error
			err := env.LogSystem.WriteLog("app.log", "ERROR", "ERR-500-XYZ: OutOfMemory exception during file upload at line 234")
			if err != nil {
				return err
			}

			// Add historical issues to knowledge graph
			node1 := &graph.Node{
				ID:   "ISS-001",
				Type: "issue",
				Data: map[string]interface{}{
					"description": "OutOfMemory during large file processing",
					"solution":    "Increase heap size to 2GB",
				},
			}

			node2 := &graph.Node{
				ID:   "ISS-002",
				Type: "issue",
				Data: map[string]interface{}{
					"description": "Upload fails with OutOfMemory",
					"solution":    "Stream files instead of loading into memory",
				},
			}

			if err := env.Graph.AddNode(node1); err != nil {
				return err
			}
			if err := env.Graph.AddNode(node2); err != nil {
				return err
			}

			// Create the email with specific ID
			_, err = env.EmailSystem.WriteEmailWithID("error_report_001", "support@company.com", "Error 500 on file upload", emailContent)
			return err
		},

		VerificationFunc: func(env *TestEnvironment) (bool, []string) {
			var errors []string
			success := true

			// Check: Ticket created
			count, err := env.Database.CountTickets()
			if err != nil || count != 1 {
				errors = append(errors, fmt.Sprintf("Expected 1 ticket, got %d", count))
				return false, errors
			}

			tickets, err := env.Database.QueryTickets(map[string]interface{}{})
			if err != nil || len(tickets) == 0 {
				errors = append(errors, "No tickets found")
				return false, errors
			}

			ticket := tickets[0]

			// Check: Ticket has high priority (error 500 is serious)
			if ticket.Priority < 4 {
				errors = append(errors, fmt.Sprintf("Expected priority >= 4, got %d", ticket.Priority))
				success = false
			}

			// Check: Ticket is linked in graph to similar issues
			_, err = env.Graph.GetNode(ticket.ID)
			if err == nil {
				// Check if it has any similar_to relationships
				neighbors, _ := env.Graph.GetNeighbors(ticket.ID, "similar_to")
				if len(neighbors) < 1 {
					errors = append(errors, "Ticket not linked to similar issues in graph")
					success = false
				} else {
					errors = append(errors, fmt.Sprintf("✓ Ticket linked to %d similar issues", len(neighbors)))
				}
			}

			// Check: Ticket tags include memory-related keywords
			hasMemoryTag := false
			for _, tag := range ticket.Tags {
				if tag == "memory" || tag == "OutOfMemory" {
					hasMemoryTag = true
					break
				}
			}
			if !hasMemoryTag {
				errors = append(errors, "Expected memory-related tags")
				success = false
			} else {
				errors = append(errors, "✓ Ticket has appropriate tags")
			}

			if success {
				errors = append(errors, "✓ All verifications passed")
			}

			return success, errors
		},
	}
}

// createComplexTask: Full workflow with auto-resolution
func createComplexTask() Task {
	return Task{
		Name:        "auto-resolve-known-issue",
		Description: "Read email ID 'urgent_001' (urgent upload error with ERR-UPLOAD-500 OutOfMemory), search logs for 'ERR-UPLOAD-500', find similar issues in graph, read feature_flags.json and known_issues.yaml configs, create high priority ticket (priority 4-5) with tags=['memory', 'upload', 'urgent'] and description including detailed log analysis (OutOfMemory, upload errors, etc), link ticket to similar issues, add auto-suggested solution to ticket",
		Complexity:  "complex",
		ExpectedOps: 15, // readEmail, searchLogs, findSimilar, getSolution, readConfig x2, createTicket, linkIssue, updateTicket, sendEmail

		SetupFunc: func(env *TestEnvironment) error {
			// Create email about upload issue
			emailContent := `From: critical@company.com
To: support@company.com
Subject: URGENT: Upload feature broken
Date: ` + time.Now().Format(time.RFC1123Z) + `

Our upload feature is completely broken. Users getting error 500.

This is affecting all customers!

Error details:
- Error code: ERR-UPLOAD-500
- Started happening 2 hours ago
- Affects files > 5MB`

			// Write multiple log entries
			env.LogSystem.WriteLog("app.log", "ERROR", "ERR-UPLOAD-500: OutOfMemory exception in UploadHandler")
			env.LogSystem.WriteLog("app.log", "WARN", "Heap usage at 95% during file upload")
			env.LogSystem.WriteLog("app.log", "ERROR", "Failed to allocate 50MB for upload buffer")

			// Add known issue to graph with solution
			issueNode := &graph.Node{
				ID:   "ISS-UPLOAD-MEM",
				Type: "issue",
				Data: map[string]interface{}{
					"description": "OutOfMemory during file upload",
					"category":    "upload",
					"severity":    "high",
				},
			}

			solutionNode := &graph.Node{
				ID:   "SOL-001",
				Type: "solution",
				Data: map[string]interface{}{
					"title":         "Increase heap size and enable streaming",
					"steps":         []string{"1. Increase JVM heap to 2GB", "2. Enable streaming mode for uploads", "3. Set max file size to 100MB"},
					"effectiveness": 0.9,
				},
			}

			env.Graph.AddNode(issueNode)
			env.Graph.AddNode(solutionNode)
			env.Graph.AddEdge("ISS-UPLOAD-MEM", "SOL-001", "solved_by")

			// Create feature flags config
			featureFlags := map[string]interface{}{
				"upload_v2":      true,
				"streaming_mode": false,
				"max_file_size":  10485760, // 10MB
			}
			env.ConfigSystem.WriteConfig("feature_flags.json", featureFlags)

			// Create known issues config
			knownIssues := map[string]interface{}{
				"upload-memory": map[string]interface{}{
					"description": "Upload fails with OutOfMemory for large files",
					"workaround":  "Use upload_v2 with streaming enabled",
					"fix_version": "2.1.0",
				},
			}
			env.ConfigSystem.WriteConfig("known_issues.yaml", knownIssues)

			// Create the email with specific ID
			_, err := env.EmailSystem.WriteEmailWithID("urgent_001", "support@company.com", "URGENT: Upload feature broken", emailContent)
			return err
		},

		VerificationFunc: func(env *TestEnvironment) (bool, []string) {
			var errors []string
			success := true

			// Check: Ticket created with high priority
			tickets, err := env.Database.QueryTickets(map[string]interface{}{})
			if err != nil || len(tickets) == 0 {
				errors = append(errors, "No tickets found")
				return false, errors
			}

			ticket := tickets[0]

			// Should be priority 5 (urgent + critical customer)
			if ticket.Priority < 4 {
				errors = append(errors, fmt.Sprintf("Expected high priority (4-5), got %d", ticket.Priority))
				success = false
			} else {
				errors = append(errors, fmt.Sprintf("✓ High priority set: %d", ticket.Priority))
			}

			// Check: Ticket has appropriate tags
			expectedTags := []string{"memory", "upload", "urgent"}
			tagCount := 0
			for _, expectedTag := range expectedTags {
				for _, tag := range ticket.Tags {
					if tag == expectedTag {
						tagCount++
						break
					}
				}
			}
			if tagCount < 2 {
				errors = append(errors, fmt.Sprintf("Expected at least 2 relevant tags, got %d", tagCount))
				success = false
			} else {
				errors = append(errors, fmt.Sprintf("✓ Ticket tagged appropriately (%d relevant tags)", tagCount))
			}

			// Check: Ticket linked to known issue in graph
			_, err = env.Graph.GetNode(ticket.ID)
			if err == nil {
				neighbors, _ := env.Graph.GetNeighbors(ticket.ID, "similar_to")
				if len(neighbors) == 0 {
					errors = append(errors, "Ticket not linked to known issue")
					success = false
				} else {
					errors = append(errors, fmt.Sprintf("✓ Linked to %d known issues", len(neighbors)))
				}
			}

			// Check: Ticket description contains log analysis
			if !containsAny(ticket.Description, []string{"OutOfMemory", "ERR-UPLOAD-500", "upload"}) {
				errors = append(errors, "Ticket missing log analysis details")
				success = false
			} else {
				errors = append(errors, "✓ Ticket includes log analysis")
			}

			// Check: Related IDs populated (linked to historical issues)
			if len(ticket.RelatedIDs) > 0 {
				errors = append(errors, fmt.Sprintf("✓ %d related tickets linked", len(ticket.RelatedIDs)))
			}

			if success {
				errors = append(errors, "✓ Complex workflow completed successfully")
			}

			return success, errors
		},
	}
}

// Helper function
func containsAny(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if len(text) > 0 && len(keyword) > 0 {
			// Simple contains check
			if len(text) >= len(keyword) {
				for i := 0; i <= len(text)-len(keyword); i++ {
					match := true
					for j := 0; j < len(keyword); j++ {
						if text[i+j] != keyword[j] {
							match = false
							break
						}
					}
					if match {
						return true
					}
				}
			}
		}
	}
	return false
}

package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Ticket represents a support ticket
type Ticket struct {
	ID          string    `json:"id"`
	CustomerID  string    `json:"customer_id"`
	Subject     string    `json:"subject"`
	Description string    `json:"description"`
	Priority    int       `json:"priority"`     // 1-5
	Status      string    `json:"status"`       // new, open, assigned, resolved
	AssignedTo  string    `json:"assigned_to"`
	Tags        []string  `json:"tags"`
	RelatedIDs  []string  `json:"related_ids"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SQLiteDB wraps database operations
type SQLiteDB struct {
	db   *sql.DB
	path string
}

// NewSQLiteDB creates or opens a SQLite database
func NewSQLiteDB(path string) (*SQLiteDB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	sqlDB := &SQLiteDB{db: db, path: path}

	// Create schema
	if err := sqlDB.createSchema(); err != nil {
		return nil, err
	}

	return sqlDB, nil
}

// createSchema creates the tickets table
func (s *SQLiteDB) createSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS tickets (
		id TEXT PRIMARY KEY,
		customer_id TEXT NOT NULL,
		subject TEXT NOT NULL,
		description TEXT,
		priority INTEGER DEFAULT 3,
		status TEXT DEFAULT 'new',
		assigned_to TEXT,
		tags TEXT,  -- JSON array
		related_ids TEXT,  -- JSON array
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_tickets_status ON tickets(status);
	CREATE INDEX IF NOT EXISTS idx_tickets_priority ON tickets(priority);
	CREATE INDEX IF NOT EXISTS idx_tickets_assigned ON tickets(assigned_to);
	`

	_, err := s.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

// CreateTicket inserts a new ticket
func (s *SQLiteDB) CreateTicket(ticket *Ticket) error {
	if ticket.ID == "" {
		ticket.ID = generateTicketID()
	}

	if ticket.Status == "" {
		ticket.Status = "new"
	}

	if ticket.Priority == 0 {
		ticket.Priority = 3
	}

	ticket.CreatedAt = time.Now()
	ticket.UpdatedAt = time.Now()

	// Serialize tags and related_ids as JSON
	tagsJSON, _ := json.Marshal(ticket.Tags)
	relatedJSON, _ := json.Marshal(ticket.RelatedIDs)

	query := `
		INSERT INTO tickets (id, customer_id, subject, description, priority, status,
		                    assigned_to, tags, related_ids, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		ticket.ID, ticket.CustomerID, ticket.Subject, ticket.Description,
		ticket.Priority, ticket.Status, ticket.AssignedTo,
		string(tagsJSON), string(relatedJSON),
		ticket.CreatedAt, ticket.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create ticket: %w", err)
	}

	return nil
}

// GetTicket retrieves a ticket by ID
func (s *SQLiteDB) GetTicket(id string) (*Ticket, error) {
	query := `
		SELECT id, customer_id, subject, description, priority, status,
		       assigned_to, tags, related_ids, created_at, updated_at
		FROM tickets WHERE id = ?
	`

	row := s.db.QueryRow(query, id)

	var ticket Ticket
	var tagsJSON, relatedJSON string

	err := row.Scan(
		&ticket.ID, &ticket.CustomerID, &ticket.Subject, &ticket.Description,
		&ticket.Priority, &ticket.Status, &ticket.AssignedTo,
		&tagsJSON, &relatedJSON,
		&ticket.CreatedAt, &ticket.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("ticket not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}

	// Deserialize JSON fields
	json.Unmarshal([]byte(tagsJSON), &ticket.Tags)
	json.Unmarshal([]byte(relatedJSON), &ticket.RelatedIDs)

	return &ticket, nil
}

// UpdateTicket updates a ticket's fields
func (s *SQLiteDB) UpdateTicket(id string, updates map[string]interface{}) error {
	ticket, err := s.GetTicket(id)
	if err != nil {
		return err
	}

	// Apply updates
	for key, value := range updates {
		switch key {
		case "priority":
			if pInt, ok := value.(int); ok {
				ticket.Priority = pInt
			} else if pFloat, ok := value.(float64); ok {
				ticket.Priority = int(pFloat)
			}
		case "status":
			ticket.Status = value.(string)
		case "assigned_to":
			ticket.AssignedTo = value.(string)
		case "tags":
			// Handle both []string and []interface{}
			if tagSlice, ok := value.([]string); ok {
				ticket.Tags = tagSlice
			} else if tagIface, ok := value.([]interface{}); ok {
				tags := make([]string, len(tagIface))
				for i, v := range tagIface {
					if s, ok := v.(string); ok {
						tags[i] = s
					}
				}
				ticket.Tags = tags
			}
		case "related_ids":
			// Handle both []string and []interface{}
			if idSlice, ok := value.([]string); ok {
				ticket.RelatedIDs = idSlice
			} else if idIface, ok := value.([]interface{}); ok {
				ids := make([]string, len(idIface))
				for i, v := range idIface {
					if s, ok := v.(string); ok {
						ids[i] = s
					}
				}
				ticket.RelatedIDs = ids
			}
		}
	}

	ticket.UpdatedAt = time.Now()

	// Serialize tags and related_ids
	tagsJSON, _ := json.Marshal(ticket.Tags)
	relatedJSON, _ := json.Marshal(ticket.RelatedIDs)

	query := `
		UPDATE tickets
		SET priority = ?, status = ?, assigned_to = ?,
		    tags = ?, related_ids = ?, updated_at = ?
		WHERE id = ?
	`

	_, err = s.db.Exec(query,
		ticket.Priority, ticket.Status, ticket.AssignedTo,
		string(tagsJSON), string(relatedJSON),
		ticket.UpdatedAt, id)

	if err != nil {
		return fmt.Errorf("failed to update ticket: %w", err)
	}

	return nil
}

// QueryTickets retrieves tickets matching criteria
func (s *SQLiteDB) QueryTickets(filters map[string]interface{}) ([]*Ticket, error) {
	query := "SELECT id, customer_id, subject, description, priority, status, assigned_to, tags, related_ids, created_at, updated_at FROM tickets WHERE 1=1"
	args := []interface{}{}

	if status, ok := filters["status"]; ok {
		query += " AND status = ?"
		args = append(args, status)
	}

	if priority, ok := filters["priority"]; ok {
		query += " AND priority = ?"
		args = append(args, priority)
	}

	if assignedTo, ok := filters["assigned_to"]; ok {
		query += " AND assigned_to = ?"
		args = append(args, assignedTo)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tickets: %w", err)
	}
	defer rows.Close()

	var tickets []*Ticket
	for rows.Next() {
		var ticket Ticket
		var tagsJSON, relatedJSON string

		err := rows.Scan(
			&ticket.ID, &ticket.CustomerID, &ticket.Subject, &ticket.Description,
			&ticket.Priority, &ticket.Status, &ticket.AssignedTo,
			&tagsJSON, &relatedJSON,
			&ticket.CreatedAt, &ticket.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(tagsJSON), &ticket.Tags)
		json.Unmarshal([]byte(relatedJSON), &ticket.RelatedIDs)

		tickets = append(tickets, &ticket)
	}

	return tickets, nil
}

// CountTickets returns the total number of tickets
func (s *SQLiteDB) CountTickets() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM tickets").Scan(&count)
	return count, err
}

// Reset clears all data from the database
func (s *SQLiteDB) Reset() error {
	_, err := s.db.Exec("DELETE FROM tickets")
	return err
}

// Close closes the database connection
func (s *SQLiteDB) Close() error {
	return s.db.Close()
}

// generateTicketID generates a unique ticket ID
func generateTicketID() string {
	return fmt.Sprintf("TKT-%d", time.Now().UnixNano()%1000000)
}

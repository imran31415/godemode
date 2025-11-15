package email

import (
	"bytes"
	"fmt"
	"io"
	"net/mail"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Email represents a parsed email message
type Email struct {
	ID          string
	From        string
	To          string
	Subject     string
	Body        string
	Date        time.Time
	Attachments []string
}

// EmailSystem handles reading and writing .eml files
type EmailSystem struct {
	inboxPath  string
	outboxPath string
}

// NewEmailSystem creates a new email system
func NewEmailSystem(inboxPath, outboxPath string) *EmailSystem {
	return &EmailSystem{
		inboxPath:  inboxPath,
		outboxPath: outboxPath,
	}
}

// ReadEmail reads an email from the inbox
func (es *EmailSystem) ReadEmail(emailID string) (*Email, error) {
	return es.ReadEmailFromFolder(emailID, es.inboxPath)
}

// ReadSentEmail reads an email from the outbox
func (es *EmailSystem) ReadSentEmail(emailID string) (*Email, error) {
	return es.ReadEmailFromFolder(emailID, es.outboxPath)
}

// ReadEmailFromFolder reads an email from a specific folder
func (es *EmailSystem) ReadEmailFromFolder(emailID string, folderPath string) (*Email, error) {
	filename := filepath.Join(folderPath, emailID+".eml")

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open email %s: %w", emailID, err)
	}
	defer file.Close()

	msg, err := mail.ReadMessage(file)
	if err != nil {
		return nil, fmt.Errorf("failed to parse email: %w", err)
	}

	// Parse body
	bodyBytes, err := io.ReadAll(msg.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read email body: %w", err)
	}

	// Parse date
	dateStr := msg.Header.Get("Date")
	date, err := mail.ParseDate(dateStr)
	if err != nil {
		// Use current time if date parsing fails
		date = time.Now()
	}

	email := &Email{
		ID:      emailID,
		From:    msg.Header.Get("From"),
		To:      msg.Header.Get("To"),
		Subject: msg.Header.Get("Subject"),
		Body:    string(bodyBytes),
		Date:    date,
	}

	return email, nil
}

// WriteEmail writes an email to the outbox
func (es *EmailSystem) WriteEmail(to, subject, body string) (string, error) {
	// Generate email ID
	emailID := fmt.Sprintf("email_%d", time.Now().Unix())

	filename := filepath.Join(es.outboxPath, emailID+".eml")

	// Create email in RFC 822 format
	var buf bytes.Buffer

	headers := map[string]string{
		"From":    "support@company.com",
		"To":      to,
		"Subject": subject,
		"Date":    time.Now().Format(time.RFC1123Z),
	}

	for key, value := range headers {
		buf.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}

	buf.WriteString("\r\n")  // Blank line separates headers from body
	buf.WriteString(body)

	err := os.WriteFile(filename, buf.Bytes(), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write email: %w", err)
	}

	return emailID, nil
}

// WriteEmailWithID writes an email with a specific ID to the inbox
func (es *EmailSystem) WriteEmailWithID(emailID, to, subject, body string) (string, error) {
	filename := filepath.Join(es.inboxPath, emailID+".eml")

	// Create email in RFC 822 format
	var buf bytes.Buffer

	headers := map[string]string{
		"From":    "support@company.com",
		"To":      to,
		"Subject": subject,
		"Date":    time.Now().Format(time.RFC1123Z),
	}

	for key, value := range headers {
		buf.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}

	buf.WriteString("\r\n")  // Blank line separates headers from body
	buf.WriteString(body)

	err := os.WriteFile(filename, buf.Bytes(), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write email: %w", err)
	}

	return emailID, nil
}

// ListEmails returns all email IDs in the inbox
func (es *EmailSystem) ListEmails() ([]string, error) {
	return es.ListEmailsInFolder(es.inboxPath)
}

func (es *EmailSystem) ListSentEmails() ([]string, error) {
	return es.ListEmailsInFolder(es.outboxPath)
}

func (es *EmailSystem) ListEmailsInFolder(folderPath string) ([]string, error) {
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read folder %s: %w", folderPath, err)
	}

	var emailIDs []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".eml") {
			// Remove .eml extension
			emailID := strings.TrimSuffix(entry.Name(), ".eml")
			emailIDs = append(emailIDs, emailID)
		}
	}

	return emailIDs, nil
}

// ExtractErrorCode extracts error codes from email body (e.g., ERR-500-XYZ)
func (e *Email) ExtractErrorCode() string {
	// Pattern: ERR-{number}-{alphanumeric}
	pattern := regexp.MustCompile(`ERR-\d+-\w+`)
	match := pattern.FindString(e.Body)
	return match
}

// ExtractKeywords extracts important keywords from subject and body
func (e *Email) ExtractKeywords() []string {
	text := strings.ToLower(e.Subject + " " + e.Body)

	keywords := []string{}
	importantTerms := []string{
		"urgent", "critical", "error", "bug", "crash",
		"outofmemory", "timeout", "slow", "failed", "broken",
		"500", "404", "503", "login", "upload", "download",
	}

	for _, term := range importantTerms {
		if strings.Contains(text, term) {
			keywords = append(keywords, term)
		}
	}

	return keywords
}

// Reset clears all emails from outbox (useful for testing)
func (es *EmailSystem) Reset() error {
	entries, err := os.ReadDir(es.outboxPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			os.Remove(filepath.Join(es.outboxPath, entry.Name()))
		}
	}

	return nil
}

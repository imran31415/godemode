package security

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// SecurityEvent represents a security event
type SecurityEvent struct {
	ID          string
	Timestamp   time.Time
	EventType   string // login_failure, suspicious_activity, data_access, etc.
	Severity    string // low, medium, high, critical
	SourceIP    string
	UserID      string
	Resource    string
	Details     map[string]interface{}
	Flagged     bool
}

// ThreatIntel represents threat intelligence data
type ThreatIntel struct {
	IP           string
	ThreatLevel  string // low, medium, high, critical
	ThreatActor  string
	LastSeen     time.Time
	AttackType   string
	Confidence   float64
	Description  string
}

// SecurityMonitor manages security events and threat intelligence
type SecurityMonitor struct {
	mu            sync.RWMutex
	events        map[string]*SecurityEvent
	threats       map[string]*ThreatIntel
	blockedIPs    map[string]bool
	compromisedUsers map[string]bool
}

// NewSecurityMonitor creates a new security monitoring system
func NewSecurityMonitor() *SecurityMonitor {
	return &SecurityMonitor{
		events:        make(map[string]*SecurityEvent),
		threats:       make(map[string]*ThreatIntel),
		blockedIPs:    make(map[string]bool),
		compromisedUsers: make(map[string]bool),
	}
}

// LogSecurityEvent logs a security event
func (sm *SecurityMonitor) LogSecurityEvent(event *SecurityEvent) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if event.ID == "" {
		event.ID = fmt.Sprintf("SEC-%d", time.Now().UnixNano())
	}
	event.Timestamp = time.Now()

	sm.events[event.ID] = event
	return nil
}

// SearchEvents searches for security events matching criteria
func (sm *SecurityMonitor) SearchEvents(filters map[string]interface{}) ([]*SecurityEvent, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var results []*SecurityEvent

	for _, event := range sm.events {
		matches := true

		if eventType, ok := filters["event_type"].(string); ok {
			if event.EventType != eventType {
				matches = false
			}
		}

		if sourceIP, ok := filters["source_ip"].(string); ok {
			if event.SourceIP != sourceIP {
				matches = false
			}
		}

		if userID, ok := filters["user_id"].(string); ok {
			if event.UserID != userID {
				matches = false
			}
		}

		if severity, ok := filters["severity"].(string); ok {
			if event.Severity != severity {
				matches = false
			}
		}

		if matches {
			results = append(results, event)
		}
	}

	return results, nil
}

// AnalyzeSuspiciousActivity detects suspicious patterns
func (sm *SecurityMonitor) AnalyzeSuspiciousActivity(timeWindow time.Duration) ([]string, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	suspicious := []string{}
	ipCounts := make(map[string]int)
	userCounts := make(map[string]int)

	cutoff := time.Now().Add(-timeWindow)

	for _, event := range sm.events {
		if event.Timestamp.After(cutoff) {
			if event.EventType == "login_failure" {
				ipCounts[event.SourceIP]++
				userCounts[event.UserID]++
			}
		}
	}

	// Flag IPs with many failed logins
	for ip, count := range ipCounts {
		if count >= 5 {
			suspicious = append(suspicious, fmt.Sprintf("IP %s: %d failed logins", ip, count))
		}
	}

	// Flag users with many failed logins
	for user, count := range userCounts {
		if count >= 3 {
			suspicious = append(suspicious, fmt.Sprintf("User %s: %d failed logins", user, count))
		}
	}

	return suspicious, nil
}

// CheckThreatIntel checks if an IP is in threat intelligence database
func (sm *SecurityMonitor) CheckThreatIntel(ip string) (*ThreatIntel, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	threat, exists := sm.threats[ip]
	return threat, exists
}

// AddThreatIntel adds threat intelligence data
func (sm *SecurityMonitor) AddThreatIntel(intel *ThreatIntel) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.threats[intel.IP] = intel
	return nil
}

// BlockIP blocks an IP address
func (sm *SecurityMonitor) BlockIP(ip string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.blockedIPs[ip] = true
	return nil
}

// IsIPBlocked checks if an IP is blocked
func (sm *SecurityMonitor) IsIPBlocked(ip string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.blockedIPs[ip]
}

// MarkUserCompromised marks a user account as compromised
func (sm *SecurityMonitor) MarkUserCompromised(userID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.compromisedUsers[userID] = true
	return nil
}

// IsUserCompromised checks if a user is marked as compromised
func (sm *SecurityMonitor) IsUserCompromised(userID string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.compromisedUsers[userID]
}

// GetBlastRadius calculates the blast radius of an incident
func (sm *SecurityMonitor) GetBlastRadius(suspectIPs []string, timeWindow time.Duration) (map[string]interface{}, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	cutoff := time.Now().Add(-timeWindow)
	affectedUsers := make(map[string]bool)
	affectedResources := make(map[string]bool)
	eventCount := 0

	for _, event := range sm.events {
		if event.Timestamp.After(cutoff) {
			for _, ip := range suspectIPs {
				if event.SourceIP == ip {
					affectedUsers[event.UserID] = true
					affectedResources[event.Resource] = true
					eventCount++
				}
			}
		}
	}

	return map[string]interface{}{
		"affected_users":     len(affectedUsers),
		"affected_resources": len(affectedResources),
		"total_events":       eventCount,
		"user_list":          keys(affectedUsers),
		"resource_list":      keys(affectedResources),
	}, nil
}

// CalculateRiskScore calculates a risk score for an incident
func (sm *SecurityMonitor) CalculateRiskScore(factors map[string]interface{}) int {
	score := 0

	if failedLogins, ok := factors["failed_logins"].(int); ok {
		score += failedLogins * 2
	}

	if affectedUsers, ok := factors["affected_users"].(int); ok {
		score += affectedUsers * 10
	}

	if hasThreatIntel, ok := factors["has_threat_intel"].(bool); ok && hasThreatIntel {
		score += 25
	}

	if dataExfiltration, ok := factors["data_exfiltration"].(bool); ok && dataExfiltration {
		score += 50
	}

	if mfaBypassed, ok := factors["mfa_bypassed"].(bool); ok && mfaBypassed {
		score += 30
	}

	// Normalize to 0-100
	if score > 100 {
		score = 100
	}

	return score
}

// GetEventTimeline builds a timeline of events
func (sm *SecurityMonitor) GetEventTimeline(filters map[string]interface{}) ([]*SecurityEvent, error) {
	events, err := sm.SearchEvents(filters)
	if err != nil {
		return nil, err
	}

	// Sort by timestamp (bubble sort for simplicity)
	for i := 0; i < len(events); i++ {
		for j := i + 1; j < len(events); j++ {
			if events[i].Timestamp.After(events[j].Timestamp) {
				events[i], events[j] = events[j], events[i]
			}
		}
	}

	return events, nil
}

// DetectAttackPattern identifies attack patterns
func (sm *SecurityMonitor) DetectAttackPattern(events []*SecurityEvent) string {
	if len(events) == 0 {
		return "unknown"
	}

	failedLogins := 0
	successfulLogin := false
	dataAccess := 0

	for _, event := range events {
		switch event.EventType {
		case "login_failure":
			failedLogins++
		case "login_success":
			successfulLogin = true
		case "data_access":
			dataAccess++
		}
	}

	if failedLogins >= 5 && successfulLogin {
		return "credential_stuffing"
	}

	if failedLogins >= 10 {
		return "brute_force"
	}

	if successfulLogin && dataAccess > 0 {
		return "unauthorized_access"
	}

	return "suspicious_activity"
}

// GenerateIncidentReport generates a comprehensive incident report
func (sm *SecurityMonitor) GenerateIncidentReport(incidentID string, events []*SecurityEvent, blastRadius map[string]interface{}) string {
	var report strings.Builder

	report.WriteString(fmt.Sprintf("=== SECURITY INCIDENT REPORT ===\n"))
	report.WriteString(fmt.Sprintf("Incident ID: %s\n", incidentID))
	report.WriteString(fmt.Sprintf("Timestamp: %s\n\n", time.Now().Format(time.RFC3339)))

	report.WriteString(fmt.Sprintf("Total Events: %d\n", len(events)))
	report.WriteString(fmt.Sprintf("Affected Users: %v\n", blastRadius["affected_users"]))
	report.WriteString(fmt.Sprintf("Affected Resources: %v\n", blastRadius["affected_resources"]))

	attackPattern := sm.DetectAttackPattern(events)
	report.WriteString(fmt.Sprintf("\nAttack Pattern: %s\n", attackPattern))

	report.WriteString("\n=== Event Timeline ===\n")
	for _, event := range events {
		report.WriteString(fmt.Sprintf("[%s] %s from %s - %s\n",
			event.Timestamp.Format(time.RFC3339),
			event.EventType,
			event.SourceIP,
			event.UserID))
	}

	return report.String()
}

// Helper function to get map keys
func keys(m map[string]bool) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}

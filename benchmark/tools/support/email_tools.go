package support

import (
	"fmt"

	"github.com/imran31415/godemode/benchmark/systems/email"
	"github.com/imran31415/godemode/benchmark/tools"
)

// EmailTools provides email-related tools
type EmailTools struct {
	emailSystem *email.EmailSystem
}

// NewEmailTools creates email tools
func NewEmailTools(emailSystem *email.EmailSystem) *EmailTools {
	return &EmailTools{emailSystem: emailSystem}
}

// RegisterTools registers all email tools with the registry
func (et *EmailTools) RegisterTools(registry *tools.Registry) error {
	// ReadEmail tool
	err := registry.Register(&tools.ToolInfo{
		Name:        "readEmail",
		Description: "Read an email from the inbox by ID",
		Parameters: []tools.ParamInfo{
			{Name: "emailID", Type: "string", Required: true},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			emailID, ok := args["emailID"]
			if !ok {
				return nil, fmt.Errorf("readEmail requires emailID parameter")
			}

			emailIDStr, ok := emailID.(string)
			if !ok {
				return nil, fmt.Errorf("emailID must be a string")
			}

			return et.emailSystem.ReadEmail(emailIDStr)
		},
	})
	if err != nil {
		return err
	}

	// SendEmail tool
	err = registry.Register(&tools.ToolInfo{
		Name:        "sendEmail",
		Description: "Send an email to a recipient",
		Parameters: []tools.ParamInfo{
			{Name: "to", Type: "string", Required: true},
			{Name: "subject", Type: "string", Required: true},
			{Name: "body", Type: "string", Required: true},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			to, ok := args["to"]
			if !ok {
				return nil, fmt.Errorf("sendEmail requires to parameter")
			}

			toStr, ok := to.(string)
			if !ok {
				return nil, fmt.Errorf("to must be a string")
			}

			subject, ok := args["subject"]
			if !ok {
				return nil, fmt.Errorf("sendEmail requires subject parameter")
			}

			subjectStr, ok := subject.(string)
			if !ok {
				return nil, fmt.Errorf("subject must be a string")
			}

			body, ok := args["body"]
			if !ok {
				return nil, fmt.Errorf("sendEmail requires body parameter")
			}

			bodyStr, ok := body.(string)
			if !ok {
				return nil, fmt.Errorf("body must be a string")
			}

			return et.emailSystem.WriteEmail(toStr, subjectStr, bodyStr)
		},
	})
	if err != nil {
		return err
	}

	// ListEmails tool
	err = registry.Register(&tools.ToolInfo{
		Name:        "listEmails",
		Description: "List all emails in the inbox",
		Parameters:  []tools.ParamInfo{},
		Function: func(args map[string]interface{}) (interface{}, error) {
			return et.emailSystem.ListEmails()
		},
	})

	return err
}

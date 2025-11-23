package githubtools

import (
	"fmt"
	"sync"
)

// ToolParameter represents a parameter for a tool
type ToolParameter struct {
	Name        string
	Type        string
	Required    bool
	Description string
}

// ToolInfo contains information about a tool
type ToolInfo struct {
	Name        string
	Description string
	Parameters  []ToolParameter
	Function    func(args map[string]interface{}) (interface{}, error)
}

// Registry holds all available tools
type Registry struct {
	mu    sync.RWMutex
	tools map[string]*ToolInfo
}

// NewRegistry creates a new tool registry with all GitHub tools
func NewRegistry() *Registry {
	r := &Registry{
		tools: make(map[string]*ToolInfo),
	}
	r.registerTools()
	return r
}

// Register adds a tool to the registry
func (r *Registry) Register(tool *ToolInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[tool.Name] = tool
}

// Get retrieves a tool by name
func (r *Registry) Get(name string) (*ToolInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tool, found := r.tools[name]
	return tool, found
}

// Call executes a tool by name with the given arguments
func (r *Registry) Call(name string, args map[string]interface{}) (interface{}, error) {
	tool, found := r.Get(name)
	if !found {
		return nil, fmt.Errorf("tool not found: %s", name)
	}
	return tool.Function(args)
}

// ListTools returns all registered tools
func (r *Registry) ListTools() []*ToolInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]*ToolInfo, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// registerTools registers all GitHub tools
func (r *Registry) registerTools() {
	r.Register(&ToolInfo{
		Name:        "search_repositories",
		Description: "Search for GitHub repositories",
		Parameters: []ToolParameter{
			{Name: "query", Type: "string", Required: true, Description: "Search query"},
			{Name: "page", Type: "int", Required: false, Description: "Page number"},
			{Name: "per_page", Type: "int", Required: false, Description: "Results per page"},
		},
		Function: searchRepositories,
	})

	r.Register(&ToolInfo{
		Name:        "get_repository",
		Description: "Get repository details",
		Parameters: []ToolParameter{
			{Name: "owner", Type: "string", Required: true, Description: "Repository owner"},
			{Name: "repo", Type: "string", Required: true, Description: "Repository name"},
		},
		Function: getRepository,
	})

	r.Register(&ToolInfo{
		Name:        "list_issues",
		Description: "List and filter repository issues",
		Parameters: []ToolParameter{
			{Name: "owner", Type: "string", Required: true, Description: "Repository owner"},
			{Name: "repo", Type: "string", Required: true, Description: "Repository name"},
			{Name: "state", Type: "string", Required: false, Description: "Issue state: open, closed, all"},
			{Name: "labels", Type: "string", Required: false, Description: "Comma-separated labels"},
			{Name: "per_page", Type: "int", Required: false, Description: "Results per page"},
		},
		Function: listIssues,
	})

	r.Register(&ToolInfo{
		Name:        "get_issue",
		Description: "Get the contents of an issue within a repository",
		Parameters: []ToolParameter{
			{Name: "owner", Type: "string", Required: true, Description: "Repository owner"},
			{Name: "repo", Type: "string", Required: true, Description: "Repository name"},
			{Name: "issue_number", Type: "int", Required: true, Description: "Issue number"},
		},
		Function: getIssue,
	})

	// NOTE: Write operations (create_issue, update_issue, add_issue_comment)
	// have been intentionally removed for safety. This benchmark is READ-ONLY.

	r.Register(&ToolInfo{
		Name:        "list_issue_comments",
		Description: "List comments on an issue",
		Parameters: []ToolParameter{
			{Name: "owner", Type: "string", Required: true, Description: "Repository owner"},
			{Name: "repo", Type: "string", Required: true, Description: "Repository name"},
			{Name: "issue_number", Type: "int", Required: true, Description: "Issue number"},
		},
		Function: listIssueComments,
	})

	r.Register(&ToolInfo{
		Name:        "search_issues",
		Description: "Search for issues and pull requests",
		Parameters: []ToolParameter{
			{Name: "query", Type: "string", Required: true, Description: "Search query"},
			{Name: "per_page", Type: "int", Required: false, Description: "Results per page"},
		},
		Function: searchIssues,
	})

	r.Register(&ToolInfo{
		Name:        "list_pull_requests",
		Description: "List and filter repository pull requests",
		Parameters: []ToolParameter{
			{Name: "owner", Type: "string", Required: true, Description: "Repository owner"},
			{Name: "repo", Type: "string", Required: true, Description: "Repository name"},
			{Name: "state", Type: "string", Required: false, Description: "PR state"},
			{Name: "per_page", Type: "int", Required: false, Description: "Results per page"},
		},
		Function: listPullRequests,
	})

	r.Register(&ToolInfo{
		Name:        "get_pull_request",
		Description: "Get details of a specific pull request",
		Parameters: []ToolParameter{
			{Name: "owner", Type: "string", Required: true, Description: "Repository owner"},
			{Name: "repo", Type: "string", Required: true, Description: "Repository name"},
			{Name: "pull_number", Type: "int", Required: true, Description: "Pull request number"},
		},
		Function: getPullRequest,
	})

	r.Register(&ToolInfo{
		Name:        "list_commits",
		Description: "Get commits of a branch in a repository",
		Parameters: []ToolParameter{
			{Name: "owner", Type: "string", Required: true, Description: "Repository owner"},
			{Name: "repo", Type: "string", Required: true, Description: "Repository name"},
			{Name: "sha", Type: "string", Required: false, Description: "Branch or commit SHA"},
			{Name: "per_page", Type: "int", Required: false, Description: "Results per page"},
		},
		Function: listCommits,
	})

	r.Register(&ToolInfo{
		Name:        "get_file_contents",
		Description: "Get contents of a file or directory",
		Parameters: []ToolParameter{
			{Name: "owner", Type: "string", Required: true, Description: "Repository owner"},
			{Name: "repo", Type: "string", Required: true, Description: "Repository name"},
			{Name: "path", Type: "string", Required: true, Description: "File path"},
			{Name: "ref", Type: "string", Required: false, Description: "Branch or commit ref"},
		},
		Function: getFileContents,
	})

	r.Register(&ToolInfo{
		Name:        "search_code",
		Description: "Search for code across GitHub repositories",
		Parameters: []ToolParameter{
			{Name: "query", Type: "string", Required: true, Description: "Search query"},
			{Name: "per_page", Type: "int", Required: false, Description: "Results per page"},
		},
		Function: searchCode,
	})

	r.Register(&ToolInfo{
		Name:        "list_labels",
		Description: "List labels for a repository",
		Parameters: []ToolParameter{
			{Name: "owner", Type: "string", Required: true, Description: "Repository owner"},
			{Name: "repo", Type: "string", Required: true, Description: "Repository name"},
		},
		Function: listLabels,
	})
}

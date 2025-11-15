package spec

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseOpenAPISpecFromBytes(t *testing.T) {
	validSpec := []byte(`{
		"openapi": "3.0.0",
		"info": {
			"title": "Example API",
			"description": "An example API",
			"version": "1.0.0"
		},
		"servers": [
			{
				"url": "https://api.example.com",
				"description": "Production server"
			}
		],
		"paths": {
			"/users/{id}": {
				"get": {
					"operationId": "getUser",
					"summary": "Get user by ID",
					"parameters": [
						{
							"name": "id",
							"in": "path",
							"required": true,
							"schema": {
								"type": "string"
							}
						}
					],
					"responses": {
						"200": {
							"description": "Successful response"
						}
					}
				}
			},
			"/users": {
				"post": {
					"operationId": "createUser",
					"summary": "Create a new user",
					"requestBody": {
						"required": true,
						"content": {
							"application/json": {
								"schema": {
									"type": "object",
									"properties": {
										"name": {
											"type": "string"
										},
										"email": {
											"type": "string"
										}
									},
									"required": ["name", "email"]
								}
							}
						}
					},
					"responses": {
						"201": {
							"description": "User created"
						}
					}
				}
			}
		}
	}`)

	spec, err := ParseOpenAPISpecFromBytes(validSpec)
	if err != nil {
		t.Fatalf("Failed to parse valid OpenAPI spec: %v", err)
	}

	if spec.OpenAPI != "3.0.0" {
		t.Errorf("Expected openapi '3.0.0', got '%s'", spec.OpenAPI)
	}

	if spec.Info.Title != "Example API" {
		t.Errorf("Expected title 'Example API', got '%s'", spec.Info.Title)
	}

	if len(spec.Servers) != 1 {
		t.Errorf("Expected 1 server, got %d", len(spec.Servers))
	}

	if len(spec.Paths) != 2 {
		t.Fatalf("Expected 2 paths, got %d", len(spec.Paths))
	}

	// Check GET operation
	getUserPath := spec.Paths["/users/{id}"]
	if getUserPath.Get == nil {
		t.Fatal("Expected GET operation on /users/{id}")
	}

	if getUserPath.Get.OperationID != "getUser" {
		t.Errorf("Expected operationId 'getUser', got '%s'", getUserPath.Get.OperationID)
	}

	// Check POST operation
	createUserPath := spec.Paths["/users"]
	if createUserPath.Post == nil {
		t.Fatal("Expected POST operation on /users")
	}

	if createUserPath.Post.RequestBody == nil {
		t.Fatal("Expected request body on POST /users")
	}
}

func TestOpenAPIToToolDefinitions(t *testing.T) {
	spec := OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: OpenAPIInfo{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: map[string]OpenAPIPathItem{
			"/users": {
				Get: &OpenAPIOperation{
					OperationID: "listUsers",
					Summary:     "List all users",
					Parameters: []OpenAPIParameter{
						{
							Name:     "limit",
							In:       "query",
							Required: false,
							Schema: &OpenAPISchema{
								Type: "integer",
							},
						},
					},
				},
				Post: &OpenAPIOperation{
					OperationID: "createUser",
					Summary:     "Create user",
					RequestBody: &OpenAPIRequestBody{
						Required: true,
						Content: map[string]OpenAPIMediaType{
							"application/json": {
								Schema: &OpenAPISchema{
									Type: "object",
									Properties: map[string]*OpenAPISchema{
										"name": {
											Type: "string",
										},
										"email": {
											Type: "string",
										},
									},
									Required: []string{"name"},
								},
							},
						},
					},
				},
			},
			"/posts/{id}": {
				Delete: &OpenAPIOperation{
					OperationID: "deletePost",
					Summary:     "Delete a post",
					Parameters: []OpenAPIParameter{
						{
							Name:     "id",
							In:       "path",
							Required: true,
							Schema: &OpenAPISchema{
								Type: "string",
							},
						},
					},
				},
			},
		},
	}

	tools := spec.ToToolDefinitions()

	// Should have 3 tools: GET /users, POST /users, DELETE /posts/{id}
	if len(tools) != 3 {
		t.Fatalf("Expected 3 tool definitions, got %d", len(tools))
	}

	// Verify tool names and descriptions
	toolMap := make(map[string]ToolDefinition)
	for _, tool := range tools {
		toolMap[tool.Name] = tool
	}

	if _, ok := toolMap["listUsers"]; !ok {
		t.Error("Expected tool 'listUsers' not found")
	}

	if _, ok := toolMap["createUser"]; !ok {
		t.Error("Expected tool 'createUser' not found")
	}

	if _, ok := toolMap["deletePost"]; !ok {
		t.Error("Expected tool 'deletePost' not found")
	}

	// Check createUser has request body parameters
	createUser := toolMap["createUser"]
	if len(createUser.Parameters) != 2 {
		t.Errorf("Expected createUser to have 2 parameters, got %d", len(createUser.Parameters))
	}

	// Check that path parameters are marked as required
	deletePost := toolMap["deletePost"]
	foundPathParam := false
	for _, param := range deletePost.Parameters {
		if param.Name == "id" && param.Required {
			foundPathParam = true
			break
		}
	}
	if !foundPathParam {
		t.Error("Expected path parameter 'id' to be required")
	}
}

func TestGenerateOperationName(t *testing.T) {
	tests := []struct {
		method   string
		path     string
		expected string
	}{
		{"GET", "/users", "getUsers"},
		{"POST", "/users", "postUsers"},
		{"GET", "/users/{id}", "getUsers"},
		{"DELETE", "/posts/{id}/comments/{commentId}", "deletePostsComments"},
		{"PUT", "/api/v1/resources", "putApiV1Resources"},
		{"PATCH", "/settings", "patchSettings"},
		{"GET", "/", "get"},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			result := generateOperationName(tt.method, tt.path)
			if result != tt.expected {
				t.Errorf("generateOperationName(%s, %s) = %s, want %s",
					tt.method, tt.path, result, tt.expected)
			}
		})
	}
}

func TestMapOpenAPITypeToGo(t *testing.T) {
	tests := []struct {
		name     string
		schema   *OpenAPISchema
		expected string
	}{
		{
			name:     "nil schema",
			schema:   nil,
			expected: "interface{}",
		},
		{
			name:     "string type",
			schema:   &OpenAPISchema{Type: "string"},
			expected: "string",
		},
		{
			name:     "string with date-time format",
			schema:   &OpenAPISchema{Type: "string", Format: "date-time"},
			expected: "time.Time",
		},
		{
			name:     "number type",
			schema:   &OpenAPISchema{Type: "number"},
			expected: "float64",
		},
		{
			name:     "number with float format",
			schema:   &OpenAPISchema{Type: "number", Format: "float"},
			expected: "float32",
		},
		{
			name:     "integer type",
			schema:   &OpenAPISchema{Type: "integer"},
			expected: "int64",
		},
		{
			name:     "integer with int32 format",
			schema:   &OpenAPISchema{Type: "integer", Format: "int32"},
			expected: "int32",
		},
		{
			name:     "boolean type",
			schema:   &OpenAPISchema{Type: "boolean"},
			expected: "bool",
		},
		{
			name: "array of strings",
			schema: &OpenAPISchema{
				Type:  "array",
				Items: &OpenAPISchema{Type: "string"},
			},
			expected: "[]string",
		},
		{
			name:     "array without items",
			schema:   &OpenAPISchema{Type: "array"},
			expected: "[]interface{}",
		},
		{
			name:     "object type",
			schema:   &OpenAPISchema{Type: "object"},
			expected: "map[string]interface{}",
		},
		{
			name:     "unknown type",
			schema:   &OpenAPISchema{Type: "unknownType"},
			expected: "interface{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapOpenAPITypeToGo(tt.schema)
			if result != tt.expected {
				t.Errorf("mapOpenAPITypeToGo() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestParseOpenAPISpecFromFile(t *testing.T) {
	// Create temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test-openapi.json")

	content := []byte(`{
		"openapi": "3.0.0",
		"info": {
			"title": "File Test API",
			"version": "1.0.0"
		},
		"paths": {
			"/test": {
				"get": {
					"summary": "Test endpoint",
					"responses": {
						"200": {
							"description": "Success"
						}
					}
				}
			}
		}
	}`)

	err := os.WriteFile(testFile, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	spec, err := ParseOpenAPISpec(testFile)
	if err != nil {
		t.Fatalf("Failed to parse OpenAPI spec from file: %v", err)
	}

	if spec.Info.Title != "File Test API" {
		t.Errorf("Expected title 'File Test API', got '%s'", spec.Info.Title)
	}

	// Test with non-existent file
	_, err = ParseOpenAPISpec("/nonexistent/file.json")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestExtractRequestBodyParams(t *testing.T) {
	requestBody := &OpenAPIRequestBody{
		Content: map[string]OpenAPIMediaType{
			"application/json": {
				Schema: &OpenAPISchema{
					Type: "object",
					Properties: map[string]*OpenAPISchema{
						"name": {
							Type: "string",
						},
						"age": {
							Type: "integer",
						},
						"email": {
							Type: "string",
						},
					},
					Required: []string{"name", "email"},
				},
			},
		},
	}

	params := extractRequestBodyParams(requestBody)

	if len(params) != 3 {
		t.Fatalf("Expected 3 parameters, got %d", len(params))
	}

	// Check that required fields are marked correctly
	paramMap := make(map[string]Parameter)
	for _, param := range params {
		paramMap[param.Name] = param
	}

	if !paramMap["name"].Required {
		t.Error("Expected 'name' to be required")
	}

	if !paramMap["email"].Required {
		t.Error("Expected 'email' to be required")
	}

	if paramMap["age"].Required {
		t.Error("Expected 'age' to not be required")
	}
}

package validator

import (
	"strings"
	"testing"
)

func TestValidateSourceSize(t *testing.T) {
	v := NewValidator()
	v.SetMaxSourceSize(100)

	// Test valid size
	smallCode := "package main\nfunc main() {}"
	if err := v.Validate(smallCode); err != nil {
		t.Errorf("Small code should be valid: %v", err)
	}

	// Test oversized code
	largeCode := strings.Repeat("a", 101)
	err := v.Validate(largeCode)
	if err == nil {
		t.Error("Oversized code should fail validation")
	}
	if !strings.Contains(err.Error(), "too large") {
		t.Errorf("Error should mention size, got: %v", err)
	}
}

func TestValidateForbiddenImports(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name      string
		code      string
		shouldErr bool
	}{
		{
			name:      "valid import",
			code:      `package main\nimport "fmt"\nfunc main() {}`,
			shouldErr: false,
		},
		{
			name:      "forbidden os/exec",
			code:      `package main\nimport "os/exec"\nfunc main() {}`,
			shouldErr: true,
		},
		{
			name:      "forbidden syscall",
			code:      `package main\nimport "syscall"\nfunc main() {}`,
			shouldErr: true,
		},
		{
			name:      "forbidden unsafe",
			code:      `package main\nimport "unsafe"\nfunc main() {}`,
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.code)
			if tt.shouldErr && err == nil {
				t.Error("Expected validation error")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

func TestValidateEmptySource(t *testing.T) {
	v := NewValidator()

	err := v.Validate("")
	if err == nil {
		t.Error("Empty source should fail validation")
	}

	err = v.Validate("   \n  \n  ")
	if err == nil {
		t.Error("Whitespace-only source should fail validation")
	}
}

func TestAddForbiddenImport(t *testing.T) {
	v := NewValidator()

	code := `package main\nimport "custom/package"\nfunc main() {}`

	// Should pass initially
	if err := v.Validate(code); err != nil {
		t.Errorf("Should be valid initially: %v", err)
	}

	// Add to forbidden list
	v.AddForbiddenImport("custom/package")

	// Should fail now
	if err := v.Validate(code); err == nil {
		t.Error("Should fail after adding to forbidden list")
	}
}

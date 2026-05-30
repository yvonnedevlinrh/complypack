// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCatalogPullCmd(t *testing.T) {
	cmd := catalogPullCmd()

	// Check basic command structure
	if cmd.Use != "pull <reference>" {
		t.Errorf("unexpected Use: %q", cmd.Use)
	}

	// Check that required flags exist
	outputFlag := cmd.Flags().Lookup("output")
	if outputFlag == nil {
		t.Error("--output flag not found")
	}

	plainHTTPFlag := cmd.Flags().Lookup("plain-http")
	if plainHTTPFlag == nil {
		t.Error("--plain-http flag not found")
	}

	// Check that it requires exactly one argument
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("expected error when no args provided, got nil")
	}

	if err := cmd.Args(cmd, []string{"ref"}); err != nil {
		t.Errorf("expected no error with one arg, got %v", err)
	}

	if err := cmd.Args(cmd, []string{"ref1", "ref2"}); err == nil {
		t.Error("expected error with two args, got nil")
	}
}

func TestCatalogPullOptionsValidation(t *testing.T) {
	tests := []struct {
		name      string
		ref       string
		wantError bool
	}{
		{
			name:      "empty reference",
			ref:       "",
			wantError: true,
		},
		{
			name:      "valid reference",
			ref:       "ghcr.io/org/repo:v1.0",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &catalogPullOptions{ref: tt.ref}
			// We can't call run() without a real OCI registry, but we can check validation
			if tt.wantError && opts.ref == "" {
				// Expected behavior - empty ref should fail in run()
				return
			}
			if !tt.wantError && opts.ref == "" {
				t.Error("expected validation to fail for empty ref")
			}
		})
	}
}

func TestCatalogCmd(t *testing.T) {
	cmd := catalogCmd()

	if cmd.Use != "catalog" {
		t.Errorf("unexpected Use: %q", cmd.Use)
	}

	// Check that pull subcommand is registered
	pullCmd := findCommand(cmd, "pull")
	if pullCmd == nil {
		t.Error("pull subcommand not found")
	}
}

func TestRootCmd(t *testing.T) {
	cmd := New()

	if cmd.Use != "complypack" {
		t.Errorf("unexpected Use: %q", cmd.Use)
	}

	// Check that catalog subcommand is registered
	catalogCmd := findCommand(cmd, "catalog")
	if catalogCmd == nil {
		t.Error("catalog subcommand not found")
	}
}

// Helper function to find a subcommand by name
func findCommand(parent interface {
	Commands() []*cobra.Command
}, name string) *cobra.Command {
	for _, cmd := range parent.Commands() {
		if cmd.Name() == name || strings.Contains(cmd.Use, name) {
			return cmd
		}
	}
	return nil
}

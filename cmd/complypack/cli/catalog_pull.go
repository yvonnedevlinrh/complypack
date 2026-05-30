// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/complytime/complypack/internal/registry"
	"github.com/gemaraproj/go-gemara/bundle"
	"github.com/spf13/cobra"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/memory"
)

type catalogPullOptions struct {
	ref       string
	output    string
	plainHTTP bool
}

// catalogPullCmd creates the `catalog pull` command.
func catalogPullCmd() *cobra.Command {
	opts := &catalogPullOptions{}

	cmd := &cobra.Command{
		Use:   "pull <reference>",
		Short: "Pull a Gemara control catalog from an OCI registry",
		Long: `Pull a Gemara control catalog from an OCI registry and output it to stdout or a file.

The reference should be a full OCI reference like:
  ghcr.io/org/controls:v1.0
  localhost:5000/catalogs/cis-benchmark@sha256:abc...

Authentication uses the Docker credential chain (same as docker login).`,
		Example: `  # Pull and output to stdout
  complypack catalog pull ghcr.io/org/controls:v1.0

  # Save to a file
  complypack catalog pull ghcr.io/org/controls:v1.0 --output controls.yaml

  # Pull from a local registry
  complypack catalog pull http://localhost:5000/controls:latest --plain-http`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.ref = args[0]
			return opts.run(cmd.Context())
		},
	}

	cmd.Flags().StringVarP(&opts.output, "output", "o", "", "Write catalog to file instead of stdout")
	cmd.Flags().BoolVar(&opts.plainHTTP, "plain-http", false, "Use HTTP instead of HTTPS")

	return cmd
}

func (o *catalogPullOptions) run(ctx context.Context) error {
	if o.ref == "" {
		return fmt.Errorf("reference cannot be empty")
	}

	// Get Docker credentials
	credFunc, err := registry.NewCredentialFunc()
	if err != nil {
		return fmt.Errorf("failed to load Docker credentials: %w", err)
	}

	// Create remote repository
	repo, err := registry.NewRepository(o.ref, credFunc, o.plainHTTP)
	if err != nil {
		return err
	}

	// Extract tag from reference
	tag := registry.ParseTag(o.ref)

	// Copy from remote to memory store
	fmt.Fprintf(os.Stderr, "Pulling %s...\n", o.ref)
	store := memory.New()
	_, err = oras.Copy(ctx, repo, tag, store, tag, oras.DefaultCopyOptions)
	if err != nil {
		return fmt.Errorf("failed to pull from registry: %w", err)
	}

	// Unpack the Gemara bundle
	b, err := bundle.Unpack(ctx, store, tag)
	if err != nil {
		return fmt.Errorf("failed to unpack bundle: %w", err)
	}

	// Report imports if present
	if len(b.Imports) > 0 {
		fmt.Fprintf(os.Stderr, "Bundle contains %d import(s)\n", len(b.Imports))
	}

	// Get the primary catalog file (first in Files array)
	if len(b.Files) == 0 {
		return fmt.Errorf("bundle contains no files")
	}
	catalogData := b.Files[0].Data

	// Write catalog to output
	writer := os.Stdout
	if o.output != "" {
		f, err := os.Create(o.output)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer f.Close()
		writer = f
		fmt.Fprintf(os.Stderr, "Writing catalog to %s\n", o.output)
	}

	if _, err := writer.Write(catalogData); err != nil {
		return fmt.Errorf("failed to write catalog: %w", err)
	}

	if o.output != "" {
		fmt.Fprintf(os.Stderr, "Successfully pulled catalog\n")
	}

	return nil
}

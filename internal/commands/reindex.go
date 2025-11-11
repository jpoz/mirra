package commands

import (
	"flag"
	"fmt"
	"log/slog"

	"github.com/llmite-ai/mirra/internal/recorder"
)

func Reindex(args []string) error {
	fs := flag.NewFlagSet("reindex", flag.ExitOnError)
	recordingsPath := fs.String("recordings", "./recordings", "Path to recordings directory")

	if err := fs.Parse(args); err != nil {
		return err
	}

	fmt.Printf("Rebuilding index for recordings in %s...\n", *recordingsPath)

	// Create a new index
	idx := recorder.NewIndex(*recordingsPath)

	// Rebuild from all JSONL files
	if err := idx.Rebuild(); err != nil {
		return fmt.Errorf("failed to rebuild index: %w", err)
	}

	// Save the index to disk
	if err := idx.Save(); err != nil {
		return fmt.Errorf("failed to save index: %w", err)
	}

	fmt.Printf("✓ Index rebuilt successfully!\n")
	fmt.Printf("  Total recordings indexed: %d\n", idx.Size())

	// Log to slog as well
	slog.Info("Index rebuilt successfully", "recordings", idx.Size())

	return nil
}

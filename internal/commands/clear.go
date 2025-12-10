package commands

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

func Clear(args []string) error {
	fs := flag.NewFlagSet("clear", flag.ExitOnError)
	recordingsPath := fs.String("recordings", "./recordings", "Path to recordings directory")
	force := fs.Bool("force", false, "Skip confirmation prompt")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Check if recordings directory exists
	if _, err := os.Stat(*recordingsPath); os.IsNotExist(err) {
		fmt.Printf("Recordings directory does not exist: %s\n", *recordingsPath)
		return nil
	}

	// Confirmation prompt unless --force is used
	if !*force {
		fmt.Printf("⚠️  This will delete all recordings in %s\n", *recordingsPath)
		fmt.Print("Are you sure you want to continue? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Cancelled")
			return nil
		}
	}

	fmt.Printf("Clearing recordings in %s...\n", *recordingsPath)

	// Remove all .jsonl files
	jsonlFiles, err := filepath.Glob(filepath.Join(*recordingsPath, "*.jsonl"))
	if err != nil {
		return fmt.Errorf("failed to list recording files: %w", err)
	}

	removedCount := 0
	for _, file := range jsonlFiles {
		if err := os.Remove(file); err != nil {
			slog.Warn("failed to remove recording file", "file", file, "error", err)
		} else {
			removedCount++
		}
	}

	// Remove index.json
	indexPath := filepath.Join(*recordingsPath, "index.json")
	if _, err := os.Stat(indexPath); err == nil {
		if err := os.Remove(indexPath); err != nil {
			slog.Warn("failed to remove index", "error", err)
		} else {
			fmt.Printf("✓ Removed index\n")
		}
	}

	// Remove groups directory and its contents
	groupsPath := filepath.Join(*recordingsPath, "groups")
	if _, err := os.Stat(groupsPath); err == nil {
		if err := os.RemoveAll(groupsPath); err != nil {
			slog.Warn("failed to remove groups directory", "error", err)
		} else {
			fmt.Printf("✓ Removed groups data\n")
		}
	}

	fmt.Printf("✓ Cleared successfully!\n")
	fmt.Printf("  Removed %d recording files\n", removedCount)

	slog.Info("Recordings cleared", "files_removed", removedCount)

	return nil
}

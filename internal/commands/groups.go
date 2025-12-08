package commands

import (
	"flag"
	"fmt"
	"time"

	"github.com/jpoz/mirra/internal/grouping"
)

// Groups handles the "mirra groups" command
func Groups(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("subcommand required: sessions")
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "sessions":
		return ListSessions(subArgs)
	default:
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

// ListSessions handles the "mirra groups sessions" command
func ListSessions(args []string) error {
	fs := flag.NewFlagSet("groups sessions", flag.ExitOnError)
	limit := fs.Int("limit", 20, "Number of groups to display")
	recordingsPath := fs.String("recordings", "./recordings", "Path to recordings directory")
	provider := fs.String("provider", "", "Filter by provider (claude|openai|gemini)")
	fromDate := fs.String("from", "", "Filter from date (YYYY-MM-DD)")
	toDate := fs.String("to", "", "Filter to date (YYYY-MM-DD)")
	showErrors := fs.Bool("errors", false, "Show only groups with errors")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// Initialize grouping manager
	manager := grouping.NewManager(*recordingsPath, true)

	// Parse date filters
	opts := &grouping.ListGroupsOptions{
		Page:     1,
		Limit:    *limit,
		Provider: *provider,
	}

	if *fromDate != "" {
		if from, err := time.Parse("2006-01-02", *fromDate); err == nil {
			opts.FromDate = &from
		} else {
			return fmt.Errorf("invalid from date: %s", *fromDate)
		}
	}

	if *toDate != "" {
		if to, err := time.Parse("2006-01-02", *toDate); err == nil {
			opts.ToDate = &to
		} else {
			return fmt.Errorf("invalid to date: %s", *toDate)
		}
	}

	if *showErrors {
		hasErrors := true
		opts.HasErrors = &hasErrors
	}

	// Get session groups
	groups, total := manager.ListSessionGroups(opts)

	if len(groups) == 0 {
		fmt.Println("No session groups found.")
		return nil
	}

	// Display results
	fmt.Printf("Found %d session groups (showing %d):\n\n", total, len(groups))

	for i, group := range groups {
		groupID := group.TraceID
		if groupID == "" {
			groupID = group.SessionID
		}

		fmt.Printf("%d. Session: %s\n", i+1, groupID)
		if group.SessionID != "" && group.TraceID != "" {
			fmt.Printf("   Session ID: %s\n", group.SessionID)
		}
		fmt.Printf("   Recordings: %d\n", group.RequestCount)
		fmt.Printf("   Providers: %v\n", group.Providers)
		fmt.Printf("   First: %s\n", group.FirstTimestamp.Format("2006-01-02 15:04:05"))
		fmt.Printf("   Last:  %s\n", group.LastTimestamp.Format("2006-01-02 15:04:05"))
		if group.HasErrors {
			fmt.Printf("   ⚠️  Has Errors\n")
		}
		fmt.Println()
	}

	if total > len(groups) {
		fmt.Printf("Showing %d of %d groups. Use --limit to see more.\n", len(groups), total)
	}

	return nil
}

// GetSession handles the "mirra groups session <trace-id>" command
func GetSession(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("trace ID required")
	}

	fs := flag.NewFlagSet("groups session", flag.ExitOnError)
	recordingsPath := fs.String("recordings", "./recordings", "Path to recordings directory")

	if err := fs.Parse(args[1:]); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	traceID := args[0]

	// Initialize grouping manager
	manager := grouping.NewManager(*recordingsPath, true)

	// Get session group
	group, err := manager.GetSessionGroup(traceID)
	if err != nil {
		return fmt.Errorf("session group not found: %w", err)
	}

	// Display session details
	fmt.Printf("Session: %s\n", group.TraceID)
	if group.SessionID != "" {
		fmt.Printf("Session ID: %s\n", group.SessionID)
	}
	fmt.Printf("Recordings: %d\n", group.RequestCount)
	fmt.Printf("Providers: %v\n", group.Providers)
	fmt.Printf("First: %s\n", group.FirstTimestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("Last:  %s\n", group.LastTimestamp.Format("2006-01-02 15:04:05"))
	if group.HasErrors {
		fmt.Printf("Has Errors: Yes\n")
	}
	fmt.Println("\nRecordings:")

	for i, recID := range group.RecordingIDs {
		fmt.Printf("  %d. %s\n", i+1, recID)
	}

	return nil
}

package main

import (
	"fmt"
	"os"

	"github.com/millwright-software/agent-desk/internal/session"
	"github.com/millwright-software/agent-desk/internal/statedb"
)

// handleSnooze prints the snooze picker tally: how often each suggestion
// row, or a typed time, has been chosen. The evidence for reshaping the
// ladder in session.SnoozeSuggestions.
func handleSnooze(profile string, args []string) {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help" || args[0] == "help") {
		fmt.Println("Usage: agent-desk snooze")
		fmt.Println()
		fmt.Println("Show how often each snooze suggestion (or a typed time) has been picked.")
		return
	}
	storage, err := session.NewStorageWithProfile(profile)
	if err != nil {
		fmt.Printf("Error: failed to initialize storage: %v\n", err)
		os.Exit(1)
	}
	db := storage.GetDB()
	if db == nil {
		fmt.Println("Error: no state database")
		os.Exit(1)
	}
	tally, err := db.SnoozeTally()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	rows := statedb.SortedTally(tally)
	if len(rows) == 0 {
		fmt.Println("No snoozes picked yet.")
		return
	}
	total := 0
	for _, r := range rows {
		total += r.Count
	}
	fmt.Printf("%-20s %5s\n", "pick", "count")
	for _, r := range rows {
		fmt.Printf("%-20s %5d\n", r.Label, r.Count)
	}
	fmt.Printf("%-20s %5d\n", "total", total)
}

// warmdesk-import reads a project from Jira, Trello, OpenProject, or Ryver
// and creates it in WarmDesk.
//
// Usage:
//
//	warmdesk-import [--config FILE] [--dry-run]
//	warmdesk-import [--config FILE] --dump-task PATH [--dump-task-ref REF]   # Ryver only, debugging
//	warmdesk-import [--config FILE] --dump-chat PATH                        # Ryver only, debugging
//
// Required fields can be supplied in the config file, as environment variables,
// or interactively when the program prompts for them.
//
// Environment variable overrides:
//
//	WARMDESK_URL, WARMDESK_USERNAME, WARMDESK_PASSWORD, WARMDESK_PROJECT
//	PLATFORM_API_TOKEN, PLATFORM_API_KEY
package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/tonk/warmdesk/migrate"
)

func main() {
	configFile := flag.String("config", "warmdesk-migrate.yaml", "path to migration config file")
	dryRun := flag.Bool("dry-run", false, "print what would be imported without writing to WarmDesk")
	dumpTask := flag.String("dump-task", "", "Ryver only: write a task's raw JSON (plus its comments) to this path and exit, for debugging")
	dumpTaskRef := flag.String("dump-task-ref", "", "Ryver only: with --dump-task, target this exact task by its short ref (e.g. CON-17) instead of auto-picking one")
	dumpChat := flag.String("dump-chat", "", "Ryver only: write the first page (up to 20) of raw chat history to this path and exit, for debugging")
	flag.Parse()

	cfg, err := migrate.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// Prompt for any required fields still missing
	cfg.Platform.Name = promptPlatform(cfg.Platform.Name)

	if *dumpTask != "" {
		if strings.ToLower(cfg.Platform.Name) != "ryver" {
			log.Fatalf("--dump-task is only supported for platform.name: ryver")
		}
		if err := migrate.DumpFirstTask(cfg.Platform, *dumpTaskRef, *dumpTask); err != nil {
			log.Fatalf("dump task: %v", err)
		}
		fmt.Printf("✓ wrote task JSON to %s\n", *dumpTask)
		return
	}

	if *dumpChat != "" {
		if strings.ToLower(cfg.Platform.Name) != "ryver" {
			log.Fatalf("--dump-chat is only supported for platform.name: ryver")
		}
		if err := migrate.DumpChatHistory(cfg.Platform, *dumpChat); err != nil {
			log.Fatalf("dump chat: %v", err)
		}
		fmt.Printf("✓ wrote chat history JSON to %s\n", *dumpChat)
		return
	}

	fmt.Printf("WarmDesk import\n")
	fmt.Printf("  source  : %s\n", strings.ToLower(cfg.Platform.Name))
	fmt.Printf("  target  : %s (project will be created)\n", cfg.WarmDesk.URL)

	// Read project from source platform
	fmt.Printf("\nReading from %s...\n", strings.Title(strings.ToLower(cfg.Platform.Name)))
	var project *migrate.Project
	switch strings.ToLower(cfg.Platform.Name) {
	case "jira":
		project, err = migrate.ImportFromJira(cfg.Platform, cfg.ColumnMap)
	case "trello":
		project, err = migrate.ImportFromTrello(cfg.Platform, cfg.ColumnMap)
	case "openproject":
		project, err = migrate.ImportFromOpenProject(cfg.Platform, cfg.ColumnMap)
	case "ryver":
		project, err = migrate.ImportFromRyver(cfg.Platform, cfg.ColumnMap, cfg.Include.CardsEnabled(), cfg.Include.ChatEnabled())
	default:
		log.Fatalf("unknown platform %q — must be jira, trello, openproject, or ryver", cfg.Platform.Name)
	}
	if err != nil {
		log.Fatalf("read from %s: %v", cfg.Platform.Name, err)
	}

	// Summary
	totalCards := 0
	for _, col := range project.Columns {
		totalCards += len(col.Cards)
	}
	fmt.Printf("\nProject: %s\n", project.Name)
	fmt.Printf("  %d column(s), %d card(s), %d topic(s), %d chat message(s)\n",
		len(project.Columns), totalCards, len(project.Topics), len(project.Messages))
	for _, col := range project.Columns {
		fmt.Printf("  %-20s  %d cards\n", col.Name, len(col.Cards))
	}

	if *dryRun {
		fmt.Println("\n[dry-run] no changes made to WarmDesk")
		return
	}

	// Authenticate with WarmDesk
	fmt.Printf("\nConnecting to WarmDesk...\n")
	token, err := migrate.Login(cfg.WarmDesk.URL, cfg.WarmDesk.Username, cfg.WarmDesk.Password)
	if err != nil {
		log.Fatalf("login: %v", err)
	}

	customerID, err := migrate.ResolveCustomerID(cfg.WarmDesk.URL, token, cfg.WarmDesk.Customer)
	if err != nil {
		log.Fatalf("resolve customer: %v", err)
	}

	userMap, err := migrate.ResolveUserMap(cfg.WarmDesk.URL, token, cfg.UserMap)
	if err != nil {
		log.Fatalf("resolve user_map: %v", err)
	}

	// Write to WarmDesk
	fmt.Printf("Creating project in WarmDesk...\n")
	// For import, ColumnMap is used in reverse: reverse map was already applied
	// during ReadFrom*, so we pass nil here to preserve the column names as-is.
	if err := migrate.WriteProject(cfg.WarmDesk.URL, token, project, nil, customerID, cfg.WarmDesk.KeyPrefix, userMap); err != nil {
		log.Fatalf("write project: %v", err)
	}

	fmt.Printf("\n✓ import complete\n")
}

func promptPlatform(current string) string {
	if current != "" {
		return current
	}
	fmt.Printf("Platform (jira|trello|openproject|ryver): ")
	var s string
	fmt.Scanln(&s)
	return strings.TrimSpace(s)
}

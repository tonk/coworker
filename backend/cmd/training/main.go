// training-seed populates the database with one set of resources per trainee.
//
// Usage (run from the backend/ directory):
//
//	go build -o training ./cmd/training
//	go run ./cmd/training COUNT PASSWORD_BASE
//	go run ./cmd/training --config /path/to/warmdesk.yaml 5 Training
//	go run ./cmd/training --reset
//
// COUNT is the number of trainees; guru00 is always the trainer slot.
// e.g. "5" creates guru00 (trainer) + guru01…guru05 (5 trainees).
// PASSWORD_BASE is prepended to the two-digit suffix, e.g. "Training" → "Training00".
//
// For each index XX the following is created:
//   - User       guru XX              (guru00, guru01, …)
//   - Customer   Ansible Laboratory XX
//   - Contract   Training XX          (under that customer)
//   - Project    EDA XX               (linked to that customer + contract, guru XX as owner)
//   - Columns    Backlog / In Progress / Done
//   - Starred    customer and project are starred for the user
//
// --reset removes every record that belongs to any guru XX user, without
// touching anything else in the database.
package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/tonk/warmdesk/config"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ─── configuration ───────────────────────────────────────────────────────────

const (
	usernamePrefix = "guru"                  // guru00, guru01, …
	emailDomain    = "@ansiblelab.nl"        // Standard email domain
	customerPrefix = "Round Table Knights "  // + XX  →  "Round Table Knights 00"
	contractPrefix = "Holy Grail Quests "    // + XX  →  "Holy Grail Quests 00"
	projectPrefix  = "Grail Finding "        // + XX  →  "Grail Finding 00"
	projectSlugPfx = "gf-"                   // + XX  →  "gf-00"
	projectKeyPfx  = "GF"
	projectColor   = "#6366f1"
)

type Character struct {
	FirstName string
	LastName  string
}

// `characters` is a package-level var since Go constants can only hold primitive values.
// Declaring it here makes it effectively fixed and accessible across the whole package.
var characters = []Character{
	{FirstName: "Bedivere", LastName: "The Wise"},
	{FirstName: "Brother", LastName: "Maynard"},
	{FirstName: "Dennis", LastName: "The Peasant"},
	{FirstName: "Frank", LastName: "The Famous Historian"},
	{FirstName: "French", LastName: "Taunter"},
	{FirstName: "Galahad", LastName: "The Pure"},
	{FirstName: "Lancelot", LastName: "The Brave"},
	{FirstName: "Prince", LastName: "Herbert"},
	{FirstName: "Robin", LastName: "The Not-So-Brave"},
	{FirstName: "Roger", LastName: "The Shrubber"},
	{FirstName: "Sir", LastName: "Bedevere"},
	{FirstName: "Sir", LastName: "Ector"},
	{FirstName: "Sir", LastName: "Gawain"},
	{FirstName: "The Black", LastName: "Knight"},
	{FirstName: "The Bridge", LastName: "Keeper"},
	{FirstName: "The French", LastName: "Taunter"},
	{FirstName: "The Green", LastName: "Knight"},
	{FirstName: "Tim", LastName: "The Enchanter"},
}
// Define the trainer, as they are King
var guru00 = Character{
	FirstName: "King",
	LastName: "Arthur",
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func must(err error) {
	if err != nil {
		log.Fatalf("training-seed: %v", err)
	}
}

func hashPassword(plain string) string {
	h, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	must(err)
	return string(h)
}

func getName(chars []Character, index int) Character {
	// Return the hash, based on the index. If the index is
	// higher then the number of hashes available, a roll-over
	// is performed.
	return chars[(index-1)%len(chars)]
}

// ─── main ─────────────────────────────────────────────────────────────────────

func main() {
	configPath := flag.String("config", "", "path to warmdesk.yaml (optional)")
	reset := flag.Bool("reset", false, "remove all guru** training data from the database")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] COUNT PASSWORD_BASE\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  COUNT          number of trainees; guru00 is always the trainer\n")
		fmt.Fprintf(os.Stderr, "                 e.g. 5 creates guru00 (trainer) + guru01…guru05 (5 trainees)\n")
		fmt.Fprintf(os.Stderr, "  PASSWORD_BASE  base password with two-digit suffix appended\n")
		fmt.Fprintf(os.Stderr, "                 e.g. Training → Training00, Training01, …\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	cfg := config.Load(*configPath)
	cfg.DBLog = "silent"
	must(database.Init(cfg))
	db := database.DB

	if *reset {
		fmt.Println("Removing all guru** training data…")
		removeTrainingData(db)
		return
	}

	args := flag.Args()
	if len(args) < 2 {
		log.Fatal("usage: go run ./cmd/training [--config PATH] [--reset] COUNT PASSWORD_BASE\n" +
			"  COUNT        number of trainees; guru00 is always the trainer (e.g. 5 creates guru00…guru05)\n" +
			"  PASSWORD_BASE  base password; two-digit suffix appended (e.g. Training → Training00)\n" +
			"  e.g.: go run ./cmd/training 5 Training")
	}

	// Copy the package-level slice so shuffling doesn't mutate the original
	shuffled := make([]Character, len(characters))
	copy(shuffled, characters)

	// And randomize them
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	n, err := strconv.Atoi(args[0])
	if err != nil || n < 0 {
		log.Fatalf("COUNT must be a non-negative integer, got %q", args[0])
	}
	passwordBase := args[1]

	fmt.Printf("Creating %d training slot(s): guru00 (trainer) + guru01…guru%02d (%d trainees)   password base: %q\n\n", n+1, n, n, passwordBase)

	created := 0
	skipped := 0
	var trainerID uint // guru00's user ID, set during slot 0

	for i := 0; i <= n; i++ {
		suffix := fmt.Sprintf("%02d", i)

		username := usernamePrefix + suffix
		email := username + emailDomain
		password := passwordBase + suffix
		customerName := customerPrefix + suffix
		contractName := contractPrefix + suffix
		projectName := projectPrefix + suffix
		projectSlug := projectSlugPfx + suffix

		// Idempotency: skip full creation if this slot already exists, but still
		// ensure CustomerAccess rows are present for both the user and the trainer.
		var existing models.User
		if db.Where("username = ?", username).First(&existing).Error == nil {
			if i == 0 {
				trainerID = existing.ID
			}
			if existing.AvatarURL == "" {
				db.Model(&existing).Update("avatar_url",
					fmt.Sprintf("https://api.dicebear.com/9.x/avataaars/svg?seed=%s", username))
			}
			var existingCustomer models.Customer
			if db.Where("name = ?", customerName).First(&existingCustomer).Error == nil {
				if existingCustomer.LogoURL == "" {
					db.Model(&existingCustomer).Update("logo_url",
						fmt.Sprintf("https://api.dicebear.com/9.x/shapes/svg?seed=%s&backgroundColor=dbeafe,e9d5ff,dcfce7", customerName))
				}
				ensureAccess(db, existing.ID, existingCustomer.ID, "member", username, i)
				if trainerID != 0 && trainerID != existing.ID {
					ensureAccess(db, trainerID, existingCustomer.ID, "admin", usernamePrefix+"00", i)
				}
			}
			fmt.Printf("  [%02d] %s already exists — skipping\n", i, username)
			skipped++
			continue
		}

		// 1. User
		thisName := guru00
		if i > 0 {
			thisName = getName(characters, i)
		}
		user := models.User{
			Email:              email,
			Username:           username,
			PasswordHash:       hashPassword(password),
			GlobalRole:         "user",
			FirstName:          thisName.FirstName,
			LastName:           thisName.LastName + " - " + suffix,
			DisplayName:        thisName.FirstName + " " + thisName.LastName + " - " + suffix,
			AvatarURL:          fmt.Sprintf("https://api.dicebear.com/9.x/avataaars/svg?seed=%s", thisName.FirstName),
			IsActive:           true,
			EmailNotifications: false,
		}
		must(db.Create(&user).Error)

		if i == 0 {
			trainerID = user.ID
		}

		// 2. Customer
		customer := models.Customer{
			Name:    customerName,
			LogoURL: fmt.Sprintf("https://api.dicebear.com/9.x/shapes/svg?seed=%s&backgroundColor=dbeafe,e9d5ff,dcfce7", customerName),
		}
		must(db.Create(&customer).Error)

		// 3. Contract
		contract := models.Contract{
			CustomerID: customer.ID,
			Name:       contractName,
		}
		must(db.Create(&contract).Error)

		// 4. Project (linked to customer + contract)
		// KeyPrefix must be unique: use "EDA" + slot suffix → "EDA00", "EDA01", …
		project := models.Project{
			Name:        projectName,
			Slug:        projectSlug,
			KeyPrefix:   projectKeyPfx + suffix,
			Color:       projectColor,
			CreatedByID: user.ID,
			CustomerID:  &customer.ID,
			ContractID:  &contract.ID,
		}
		must(db.Create(&project).Error)

		// 5. Default columns
		for j, colName := range []string{"Backlog", "In Progress", "Done"} {
			must(db.Create(&models.Column{
				ProjectID: project.ID,
				Name:      colName,
				Position:  float64((j + 1) * 1000),
			}).Error)
		}

		// 6. Project membership (owner)
		must(db.Create(&models.ProjectMember{
			ProjectID: project.ID,
			UserID:    user.ID,
			Role:      "owner",
		}).Error)

		// 7. Star the project and customer for the user
		must(db.Create(&models.StarredProject{
			UserID:    user.ID,
			ProjectID: project.ID,
		}).Error)
		must(db.Create(&models.CustomerFavorite{
			UserID:     user.ID,
			CustomerID: customer.ID,
		}).Error)

		// 8. Customer access.
		// Every user (including the trainer) gets an explicit row so they can
		// see their own customer under the strict visibility model.
		// Trainees get "member"; the trainer gets "admin" so they can manage all.
		must(db.Create(&models.CustomerAccess{
			UserID:     user.ID,
			CustomerID: customer.ID,
			Role:       "member",
		}).Error)
		// For trainee slots: also give the trainer access to this customer.
		if i > 0 && trainerID != 0 {
			must(db.Create(&models.CustomerAccess{
				UserID:     trainerID,
				CustomerID: customer.ID,
				Role:       "admin",
			}).Error)
		}

		fmt.Printf("  [%02d] %-10s  pw=%-14s  customer=%q\n",
			i, username, password, customerName)
		created++

		// 9. Create group "Shrubbery Bringing <XX>", add the user, and assign to the project.
		g := &models.UserGroup{Name: fmt.Sprintf("Shrubbery Bringing %s", suffix)}
		must(db.Create(g).Error)
		must(db.Create(&models.GroupMember{GroupID: g.ID, UserID: user.ID}).Error)
		must(db.Create(&models.GroupProjectAccess{
			GroupID:   g.ID,
			ProjectID: project.ID,
			Role:      "member",
		}).Error)
	}

	fmt.Println()
	fmt.Printf("Done: %d created, %d skipped.\n", created, skipped)
}

// ensureAccess creates a CustomerAccess row only if one does not already exist.
func ensureAccess(db *gorm.DB, userID, customerID uint, role, username string, slot int) {
	var count int64
	db.Model(&models.CustomerAccess{}).
		Where("user_id = ? AND customer_id = ?", userID, customerID).
		Count(&count)
	if count == 0 {
		must(db.Create(&models.CustomerAccess{
			UserID:     userID,
			CustomerID: customerID,
			Role:       role,
		}).Error)
		fmt.Printf("  [%02d] %s — added missing customer access\n", slot, username)
	}
}

// ─── reset ────────────────────────────────────────────────────────────────────

func removeTrainingData(db *gorm.DB) {
	// Collect IDs of all guru** users (usernamePrefix + exactly 2 chars).
	var userIDs []uint
	db.Model(&models.User{}).
		Where("username LIKE ? AND LENGTH(username) = ?", usernamePrefix+"__", len(usernamePrefix)+2).
		Pluck("id", &userIDs)

	// Projects matching projectPrefix + XX.
	var projectIDs []uint
	db.Unscoped().Model(&models.Project{}).
		Where("name LIKE ? AND LENGTH(name) = ?", projectPrefix+"__", len(projectPrefix)+2).
		Pluck("id", &projectIDs)

	if len(projectIDs) > 0 {
		// Cards and their dependants.
		var cardIDs []uint
		db.Unscoped().Model(&models.Card{}).
			Where("project_id IN ?", projectIDs).
			Pluck("id", &cardIDs)
		if len(cardIDs) > 0 {
			db.Unscoped().Where("card_id IN ?", cardIDs).Delete(&models.CardComment{})
			db.Where("card_id IN ?", cardIDs).Delete(&models.CardChecklistItem{})
			db.Where("card_id IN ?", cardIDs).Delete(&models.CardTag{})
			db.Exec("DELETE FROM card_labels WHERE card_id IN ?", cardIDs)
			db.Exec("DELETE FROM card_assignees WHERE card_id IN ?", cardIDs)
			db.Where("source_card_id IN ? OR target_card_id IN ?", cardIDs, cardIDs).
				Delete(&models.CardReference{})
		}
		db.Unscoped().Where("project_id IN ?", projectIDs).Delete(&models.Card{})
		db.Unscoped().Where("project_id IN ?", projectIDs).Delete(&models.Column{})
		db.Unscoped().Where("project_id IN ?", projectIDs).Delete(&models.Label{})
		db.Where("project_id IN ?", projectIDs).Delete(&models.StarredProject{})
		db.Where("project_id IN ?", projectIDs).Delete(&models.ProjectMember{})

		var topicIDs []uint
		db.Model(&models.Topic{}).
			Where("project_id IN ?", projectIDs).
			Pluck("id", &topicIDs)
		if len(topicIDs) > 0 {
			db.Unscoped().Where("topic_id IN ?", topicIDs).Delete(&models.TopicReply{})
		}
		db.Unscoped().Where("project_id IN ?", projectIDs).Delete(&models.Topic{})
		db.Unscoped().Where("id IN ?", projectIDs).Delete(&models.Project{})
		fmt.Printf("  Removed %d EDA project(s)\n", len(projectIDs))
	}

	// Customers + contracts matching customerPrefix + XX.
	var custIDs []uint
	db.Model(&models.Customer{}).
		Where("name LIKE ? AND LENGTH(name) = ?", customerPrefix+"__", len(customerPrefix)+2).
		Pluck("id", &custIDs)
	if len(custIDs) > 0 {
		db.Where("customer_id IN ?", custIDs).Delete(&models.CustomerFavorite{})
		db.Where("customer_id IN ?", custIDs).Delete(&models.CustomerAccess{})
		db.Where("customer_id IN ?", custIDs).Delete(&models.Contract{})
		db.Where("id IN ?", custIDs).Delete(&models.Customer{})
		fmt.Printf("  Removed %d customer(s) and contract(s)\n", len(custIDs))
	}

	// Groups matching "Holy Grail <XX>".
	var groupIDs []uint
	db.Model(&models.UserGroup{}).
		Where("name LIKE ? AND LENGTH(name) = ?", "Holy Grail "+"__", len("Holy Grail ")+2).
		Pluck("id", &groupIDs)
	if len(groupIDs) > 0 {
		db.Where("group_id IN ?", groupIDs).Delete(&models.GroupMember{})
		db.Where("group_id IN ?", groupIDs).Delete(&models.GroupProjectAccess{})
		db.Where("group_id IN ?", groupIDs).Delete(&models.GroupCustomerAccess{})
		db.Unscoped().Where("id IN ?", groupIDs).Delete(&models.UserGroup{})
		fmt.Printf("  Removed %d training group(s)\n", len(groupIDs))
	}

	// Users.
	if len(userIDs) > 0 {
		db.Where("user_id IN ?", userIDs).Delete(&models.CustomerAccess{})
		db.Unscoped().Where("id IN ?", userIDs).Delete(&models.User{})
		fmt.Printf("  Removed %d guru user(s)\n", len(userIDs))
	}

	if len(projectIDs) == 0 && len(custIDs) == 0 && len(userIDs) == 0 {
		fmt.Println("  Nothing to remove.")
	} else {
		fmt.Println("Done.")
	}
}

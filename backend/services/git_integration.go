package services

import (
	"regexp"
	"strconv"

	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/models"
)

// cardRefRe matches card references like PRJ-42 or WAR2-7 (1-10 uppercase letters/digits, dash, digits).
var cardRefRe = regexp.MustCompile(`\b([A-Z][A-Z0-9]{0,9})-(\d+)\b`)

// LinkCardsFromText scans text for card references, looks up the matching cards,
// and upserts a CardLink for each found card.  The template argument provides all
// non-card-specific fields (Platform, LinkType, Title, URL, Reference, Author,
// Status, RepoName); CardID is filled in per matched card.
// Returns the list of upserted links.
func LinkCardsFromText(text string, template models.CardLink) []models.CardLink {
	matches := cardRefRe.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return nil
	}

	seen := map[string]bool{}
	var result []models.CardLink

	for _, m := range matches {
		ref := m[0]
		if seen[ref] {
			continue
		}
		seen[ref] = true

		prefix := m[1]
		cardNum, err := strconv.Atoi(m[2])
		if err != nil {
			continue
		}

		// Find the project with this key_prefix.
		var project models.Project
		if err := database.DB.Where("key_prefix = ?", prefix).First(&project).Error; err != nil {
			continue
		}

		// Find the card within that project.
		var card models.Card
		if err := database.DB.Where("project_id = ? AND card_number = ?", project.ID, cardNum).First(&card).Error; err != nil {
			continue
		}

		// Upsert: find existing link by (card_id, platform, link_type, reference).
		link := template
		link.CardID = card.ID

		var existing models.CardLink
		err = database.DB.Where(models.CardLink{
			CardID:    card.ID,
			Platform:  link.Platform,
			LinkType:  link.LinkType,
			Reference: link.Reference,
		}).First(&existing).Error

		if err == nil {
			// Update mutable fields.
			database.DB.Model(&existing).Updates(map[string]interface{}{
				"title":     link.Title,
				"url":       link.URL,
				"author":    link.Author,
				"status":    link.Status,
				"repo_name": link.RepoName,
			})
			link.ID = existing.ID
		} else {
			database.DB.Create(&link)
		}

		result = append(result, link)
	}
	return result
}

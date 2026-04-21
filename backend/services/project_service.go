package services

import (
	"errors"
	"unicode"

	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/models"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrForbidden     = errors.New("forbidden")
	ErrAlreadyExists = errors.New("already exists")
)

var roleRank = map[string]int{
	"viewer": 1,
	"member": 2,
	"admin":  3,
	"owner":  4,
}

// GetProjectBySlug returns the project with the given slug.
func GetProjectBySlug(slug string) (*models.Project, error) {
	var project models.Project
	if err := database.DB.Where("slug = ? AND deleted_at IS NULL", slug).First(&project).Error; err != nil {
		return nil, ErrNotFound
	}
	return &project, nil
}

// GetMemberRole returns the role of a user in a project, or "" if not a member.
func GetMemberRole(projectID, userID uint) string {
	var member models.ProjectMember
	if err := database.DB.Where("project_id = ? AND user_id = ?", projectID, userID).First(&member).Error; err != nil {
		return ""
	}
	return member.Role
}

// GetGroupProjectRole returns the highest role any of the user's groups grant
// on the given project, or "" if no group access exists.
func GetGroupProjectRole(projectID, userID uint) string {
	type row struct{ Role string }
	var rows []row
	database.DB.Raw(`
		SELECT gpa.role FROM group_project_accesses gpa
		JOIN group_members gm ON gm.group_id = gpa.group_id
		WHERE gpa.project_id = ? AND gm.user_id = ?`,
		projectID, userID).Scan(&rows)
	best := 0
	bestRole := ""
	for _, r := range rows {
		if roleRank[r.Role] > best {
			best = roleRank[r.Role]
			bestRole = r.Role
		}
	}
	return bestRole
}

// RequireProjectRole checks the user has at least minRole in the project.
// Global admins pass unconditionally. Global viewers get implicit viewer-level
// access to every project (read-only), but are blocked on write operations.
// Group-based access is considered after direct membership.
func RequireProjectRole(projectID, userID uint, globalRole, minRole string) error {
	if globalRole == "admin" {
		return nil
	}
	if globalRole == "viewer" {
		if roleRank[minRole] <= roleRank["viewer"] {
			return nil
		}
		return ErrForbidden
	}
	if roleRank[GetMemberRole(projectID, userID)] >= roleRank[minRole] {
		return nil
	}
	if roleRank[GetGroupProjectRole(projectID, userID)] >= roleRank[minRole] {
		return nil
	}
	return ErrForbidden
}

// GenerateSlug creates a URL-safe slug from a name and ensures uniqueness.
func GenerateSlug(name string) string {
	slug := slugify(name)
	base := slug
	counter := 2
	for {
		var count int64
		database.DB.Model(&models.Project{}).Where("slug = ?", slug).Count(&count)
		if count == 0 {
			return slug
		}
		slug = base + "-" + itoa(counter)
		counter++
	}
}

// GenerateKeyPrefix creates a unique uppercase abbreviation from a project name.
// It derives a 3-character base (e.g. "WarmDesk" → "WAR", "My Project" → "MYP")
// and appends a numeric suffix if that base is already taken ("WAR2", "WAR3", …).
func GenerateKeyPrefix(name string) string {
	base := keyPrefixBase(name)
	candidate := base
	n := 2
	for {
		var count int64
		database.DB.Model(&models.Project{}).Where("key_prefix = ?", candidate).Count(&count)
		if count == 0 {
			return candidate
		}
		candidate = base + itoa(n)
		n++
	}
}

// keyPrefixBase derives the 3-character abbreviation without uniqueness checking.
func keyPrefixBase(name string) string {
	// Split into words by non-alphanumeric characters
	var words [][]rune
	var current []rune
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current = append(current, unicode.ToUpper(r))
		} else if len(current) > 0 {
			words = append(words, current)
			current = nil
		}
	}
	if len(current) > 0 {
		words = append(words, current)
	}

	var result []rune
	// Take first letter of each word (up to 3)
	for _, w := range words {
		if len(result) >= 3 {
			break
		}
		result = append(result, w[0])
	}
	// If still < 3, take more chars from the first word
	if len(result) < 3 && len(words) > 0 {
		for i := 1; i < len(words[0]) && len(result) < 3; i++ {
			result = append(result, words[0][i])
		}
	}
	// Pad with X if name too short
	for len(result) < 3 {
		result = append(result, 'X')
	}
	return string(result[:3])
}

func slugify(s string) string {
	result := make([]byte, 0, len(s))
	prev := '-'
	for _, r := range s {
		var b byte
		switch {
		case r >= 'a' && r <= 'z':
			b = byte(r)
		case r >= 'A' && r <= 'Z':
			b = byte(r + 32)
		case r >= '0' && r <= '9':
			b = byte(r)
		default:
			b = '-'
		}
		if b == '-' && prev == '-' {
			continue
		}
		result = append(result, b)
		prev = rune(b)
	}
	// trim trailing dash
	for len(result) > 0 && result[len(result)-1] == '-' {
		result = result[:len(result)-1]
	}
	if len(result) == 0 {
		return "project"
	}
	return string(result)
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	buf := make([]byte, 0, 10)
	for i > 0 {
		buf = append([]byte{byte('0' + i%10)}, buf...)
		i /= 10
	}
	return string(buf)
}

package migrate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

// ─── HTTP helper ─────────────────────────────────────────────────────────────

// do performs an HTTP request and returns the response body, status code, and
// any transport-level error. A non-2xx status is NOT treated as an error here;
// callers inspect the status code themselves.
func cwDo(method, url string, headers map[string]string, body interface{}) ([]byte, int, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, 0, fmt.Errorf("new request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("http %s %s: %w", method, url, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response body: %w", err)
	}
	return data, resp.StatusCode, nil
}

// ─── Auth ────────────────────────────────────────────────────────────────────

// Login authenticates with WarmDesk and returns the JWT access token.
func Login(baseURL, username, password string) (string, error) {
	type loginReq struct {
		Login    string `json:"login"` // email or username
		Password string `json:"password"`
	}
	type loginResp struct {
		AccessToken string `json:"access_token"`
	}

	data, status, err := cwDo("POST", baseURL+"/api/v1/auth/login", nil, loginReq{
		Login:    username,
		Password: password,
	})
	if err != nil {
		return "", err
	}
	if status != http.StatusOK {
		return "", fmt.Errorf("login failed (HTTP %d): %s", status, string(data))
	}

	var resp loginResp
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", fmt.Errorf("parse login response: %w", err)
	}
	if resp.AccessToken == "" {
		return "", fmt.Errorf("empty access token in login response")
	}
	return resp.AccessToken, nil
}

// ─── Read project ─────────────────────────────────────────────────────────────

// apiProject matches the WarmDesk JSON project response.
type apiProject struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Slug        string `json:"slug"`
}

type apiColumn struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type apiCard struct {
	ID               uint        `json:"id"`
	CardNumber       int         `json:"card_number"`
	Title            string      `json:"title"`
	Description      string      `json:"description"`
	Priority         string      `json:"priority"`
	DueDate          *time.Time  `json:"due_date"`
	Closed           bool        `json:"closed"`
	TimeSpentMinutes int         `json:"time_spent_minutes"`
	Labels           []apiLabel  `json:"labels"`
	Tags             []apiTag    `json:"tags"`
	Assignees        []apiUser   `json:"assignees"`
	Assignee         *apiUser    `json:"assignee"`
	Attachments      []apiAttach `json:"attachments"`
}

type apiLabel struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type apiTag struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type apiUser struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
}

type apiAttach struct {
	ID       uint   `json:"id"`
	Filename string `json:"original_filename"`
	MimeType string `json:"mime_type"`
}

type apiComment struct {
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	User      apiUser   `json:"user"`
}

type apiCheckItem struct {
	Body        string `json:"body"`
	IsCompleted bool   `json:"is_completed"`
}

type apiTopic struct {
	ID         uint    `json:"id"`
	Title      string  `json:"title"`
	Body       string  `json:"body"`
	User       apiUser `json:"user"`
	ReplyCount int     `json:"reply_count"`
}

type apiTopicDetail struct {
	ID      uint            `json:"id"`
	Title   string          `json:"title"`
	Body    string          `json:"body"`
	User    apiUser         `json:"user"`
	Replies []apiTopicReply `json:"replies"`
}

type apiTopicReply struct {
	Body string  `json:"body"`
	User apiUser `json:"user"`
}

// ReadProject fetches the full project from WarmDesk into the canonical types.
func ReadProject(baseURL, token, slug string) (*Project, error) {
	hdrs := map[string]string{"Authorization": "Bearer " + token}

	// ── project meta ─────────────────────────────────────────────────────────
	data, status, err := cwDo("GET", baseURL+"/api/v1/projects/"+slug, hdrs, nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("get project (HTTP %d): %s", status, string(data))
	}
	var proj apiProject
	if err := json.Unmarshal(data, &proj); err != nil {
		return nil, fmt.Errorf("parse project: %w", err)
	}

	result := &Project{
		Name:        proj.Name,
		Description: proj.Description,
	}

	// ── columns ───────────────────────────────────────────────────────────────
	data, status, err = cwDo("GET", baseURL+"/api/v1/projects/"+slug+"/columns", hdrs, nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("get columns (HTTP %d): %s", status, string(data))
	}
	var apiColumns []apiColumn
	if err := json.Unmarshal(data, &apiColumns); err != nil {
		return nil, fmt.Errorf("parse columns: %w", err)
	}

	for _, col := range apiColumns {
		cwCol := Column{Name: col.Name}

		// ── cards in column ───────────────────────────────────────────────────
		cardURL := fmt.Sprintf("%s/api/v1/projects/%s/columns/%d/cards", baseURL, slug, col.ID)
		data, status, err = cwDo("GET", cardURL, hdrs, nil)
		if err != nil {
			return nil, err
		}
		if status != http.StatusOK {
			return nil, fmt.Errorf("get cards for column %d (HTTP %d)", col.ID, status)
		}
		var cards []apiCard
		if err := json.Unmarshal(data, &cards); err != nil {
			return nil, fmt.Errorf("parse cards: %w", err)
		}

		for _, c := range cards {
			card, err := fetchFullCard(baseURL, slug, token, c)
			if err != nil {
				return nil, err
			}
			cwCol.Cards = append(cwCol.Cards, card)
		}

		result.Columns = append(result.Columns, cwCol)
	}

	// ── topics ────────────────────────────────────────────────────────────────
	data, status, err = cwDo("GET", baseURL+"/api/v1/projects/"+slug+"/topics", hdrs, nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("get topics (HTTP %d): %s", status, string(data))
	}
	var topics []apiTopic
	if err := json.Unmarshal(data, &topics); err != nil {
		return nil, fmt.Errorf("parse topics: %w", err)
	}

	for _, t := range topics {
		topic, err := fetchTopicDetail(baseURL, slug, token, t.ID)
		if err != nil {
			return nil, err
		}
		result.Topics = append(result.Topics, *topic)
	}

	return result, nil
}

// fetchFullCard loads comments, checklist and formats card from the API.
func fetchFullCard(baseURL, slug, token string, c apiCard) (Card, error) {
	hdrs := map[string]string{"Authorization": "Bearer " + token}

	card := Card{
		Title:       c.Title,
		Description: c.Description,
		Priority:    c.Priority,
		Closed:      c.Closed,
		TimeMinutes: c.TimeSpentMinutes,
	}
	if c.CardNumber > 0 {
		card.Ref = fmt.Sprintf("%d", c.CardNumber)
	}
	if c.DueDate != nil {
		card.DueDate = c.DueDate.Format("2006-01-02")
	}

	// Labels
	for _, l := range c.Labels {
		card.Labels = append(card.Labels, Label{Name: l.Name, Color: l.Color})
	}

	// Tags
	for _, t := range c.Tags {
		card.Tags = append(card.Tags, t.Name)
	}

	// Assignees (multiple)
	seen := map[uint]bool{}
	for _, a := range c.Assignees {
		if !seen[a.ID] {
			seen[a.ID] = true
			name := a.DisplayName
			if name == "" {
				name = a.Username
			}
			card.Assignees = append(card.Assignees, name)
		}
	}
	// Legacy single assignee fallback
	if c.Assignee != nil && !seen[c.Assignee.ID] {
		name := c.Assignee.DisplayName
		if name == "" {
			name = c.Assignee.Username
		}
		card.Assignees = append(card.Assignees, name)
	}

	// Attachments
	for _, a := range c.Attachments {
		card.Attachments = append(card.Attachments, Attachment{
			Filename: a.Filename,
			URL:      fmt.Sprintf("%s/api/v1/attachments/%d", baseURL, a.ID),
			MimeType: a.MimeType,
		})
	}

	// Comments
	commentURL := fmt.Sprintf("%s/api/v1/projects/%s/cards/%d/comments", baseURL, slug, c.ID)
	data, status, err := cwDo("GET", commentURL, hdrs, nil)
	if err != nil {
		return card, err
	}
	if status == http.StatusOK {
		var comments []apiComment
		if err := json.Unmarshal(data, &comments); err == nil {
			for _, cm := range comments {
				name := cm.User.DisplayName
				if name == "" {
					name = cm.User.Username
				}
				card.Comments = append(card.Comments, Comment{
					Author:    name,
					Body:      cm.Body,
					CreatedAt: cm.CreatedAt,
				})
			}
		}
	}

	// Checklist
	checkURL := fmt.Sprintf("%s/api/v1/projects/%s/cards/%d/checklist", baseURL, slug, c.ID)
	data, status, err = cwDo("GET", checkURL, hdrs, nil)
	if err != nil {
		return card, err
	}
	if status == http.StatusOK {
		var items []apiCheckItem
		if err := json.Unmarshal(data, &items); err == nil {
			for _, item := range items {
				card.Checklist = append(card.Checklist, CheckItem{
					Text: item.Body,
					Done: item.IsCompleted,
				})
			}
		}
	}

	return card, nil
}

func fetchTopicDetail(baseURL, slug, token string, topicID uint) (*Topic, error) {
	hdrs := map[string]string{"Authorization": "Bearer " + token}
	url := fmt.Sprintf("%s/api/v1/projects/%s/topics/%d", baseURL, slug, topicID)
	data, status, err := cwDo("GET", url, hdrs, nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("get topic %d (HTTP %d)", topicID, status)
	}

	var t apiTopicDetail
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, fmt.Errorf("parse topic detail: %w", err)
	}

	author := t.User.DisplayName
	if author == "" {
		author = t.User.Username
	}

	topic := &Topic{
		Title:  t.Title,
		Body:   t.Body,
		Author: author,
	}
	for _, r := range t.Replies {
		rAuthor := r.User.DisplayName
		if rAuthor == "" {
			rAuthor = r.User.Username
		}
		topic.Replies = append(topic.Replies, TopicReply{
			Author: rAuthor,
			Body:   r.Body,
		})
	}
	return topic, nil
}

// ─── Write project ────────────────────────────────────────────────────────────

// ResolveCustomerID looks up a WarmDesk customer by exact name (case
// insensitive) and returns its ID. Every project must belong to a customer,
// so this is called before WriteProject.
func ResolveCustomerID(baseURL, token, name string) (uint, error) {
	hdrs := map[string]string{"Authorization": "Bearer " + token}
	data, status, err := cwDo("GET", baseURL+"/api/v1/customers", hdrs, nil)
	if err != nil {
		return 0, err
	}
	if status != http.StatusOK {
		return 0, fmt.Errorf("list customers (HTTP %d): %s", status, string(data))
	}

	var customers []struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(data, &customers); err != nil {
		return 0, fmt.Errorf("parse customers: %w", err)
	}
	for _, cust := range customers {
		if strings.EqualFold(cust.Name, name) {
			return cust.ID, nil
		}
	}
	return 0, fmt.Errorf("customer %q not found (or not visible to this account)", name)
}

// ResolvedUser is a WarmDesk account resolved from an external assignee name.
type ResolvedUser struct {
	ID       uint
	Username string
}

// ResolveUserMap turns a config-supplied {external name: WarmDesk username}
// map into {external name: ResolvedUser}, looking up each username via the
// admin users list. An external name whose configured WarmDesk username
// doesn't exist is dropped with a warning rather than failing the import —
// its cards are just left unassigned.
func ResolveUserMap(baseURL, token string, userMap map[string]string) (map[string]ResolvedUser, error) {
	resolved := make(map[string]ResolvedUser, len(userMap))
	if len(userMap) == 0 {
		return resolved, nil
	}

	hdrs := map[string]string{"Authorization": "Bearer " + token}
	data, status, err := cwDo("GET", baseURL+"/api/v1/admin/users", hdrs, nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("list users (HTTP %d): %s", status, string(data))
	}

	var users []struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
	}
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, fmt.Errorf("parse users: %w", err)
	}
	byUsername := make(map[string]ResolvedUser, len(users))
	for _, u := range users {
		byUsername[strings.ToLower(u.Username)] = ResolvedUser{ID: u.ID, Username: u.Username}
	}

	for externalName, username := range userMap {
		ru, ok := byUsername[strings.ToLower(username)]
		if !ok {
			fmt.Printf("  ⚠ user_map: WarmDesk username %q (mapped from %q) not found — those cards will be left unassigned\n", username, externalName)
			continue
		}
		resolved[externalName] = ru
	}
	return resolved, nil
}

// WriteProject creates a project and all its content in WarmDesk using the
// REST API. customerID is required — WarmDesk rejects project creation
// without one. explicitPrefix overrides the auto-derived card prefix (pass ""
// to keep the default); WarmDesk enforces globally unique prefixes, so two
// source project names that happen to share their first 3 letters need one.
// userMap resolves external assignee names (e.g. a Ryver display name) to a
// WarmDesk user ID, as produced by ResolveUserMap.
func WriteProject(baseURL, token string, p *Project, columnMap map[string]string, customerID uint, explicitPrefix string, userMap map[string]ResolvedUser) error {
	hdrs := map[string]string{"Authorization": "Bearer " + token}

	// ── create project ────────────────────────────────────────────────────────
	slug := slugify(p.Name)
	type createProjReq struct {
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
		Color       string `json:"color"`
		KeyPrefix   string `json:"key_prefix"`
		CustomerID  uint   `json:"customer_id"`
	}
	prefix := strings.ToUpper(slug)
	if len(prefix) > 3 {
		prefix = prefix[:3]
	}
	if explicitPrefix != "" {
		prefix = strings.ToUpper(explicitPrefix)
	}
	data, status, err := cwDo("POST", baseURL+"/api/v1/projects", hdrs, createProjReq{
		Name:        p.Name,
		Slug:        slug,
		Description: p.Description,
		Color:       "#6366f1",
		KeyPrefix:   prefix,
		CustomerID:  customerID,
	})
	if err != nil {
		return err
	}
	if status != http.StatusCreated && status != http.StatusOK {
		return fmt.Errorf("create project (HTTP %d): %s", status, string(data))
	}
	var createdProj apiProject
	if err := json.Unmarshal(data, &createdProj); err != nil {
		return fmt.Errorf("parse created project: %w", err)
	}
	fmt.Printf("  → created project %q (slug=%s)\n", p.Name, createdProj.Slug)

	// ── collect unique labels ─────────────────────────────────────────────────
	labelIDMap := map[string]uint{} // label name → id in WarmDesk
	uniqueLabels := map[string]Label{}
	for _, col := range p.Columns {
		for _, card := range col.Cards {
			for _, l := range card.Labels {
				uniqueLabels[l.Name] = l
			}
		}
	}
	// CreateProject auto-populates every new project with default labels
	// (from the "default_labels" system setting), and CreateLabel does no
	// dedup of its own — match against existing labels case-insensitively
	// first, or a source tag like "bug" ends up duplicating an existing
	// "Bug" label instead of reusing it.
	existingLabels := map[string]apiLabel{} // lowercase name → label
	if data, status, err := cwDo("GET", fmt.Sprintf("%s/api/v1/projects/%s/labels", baseURL, createdProj.Slug), hdrs, nil); err == nil && status == http.StatusOK {
		var labels []apiLabel
		if json.Unmarshal(data, &labels) == nil {
			for _, lbl := range labels {
				existingLabels[strings.ToLower(lbl.Name)] = lbl
			}
		}
	}

	for _, l := range uniqueLabels {
		if existing, ok := existingLabels[strings.ToLower(l.Name)]; ok {
			labelIDMap[l.Name] = existing.ID
			continue
		}

		type createLabelReq struct {
			Name  string `json:"name"`
			Color string `json:"color"`
		}
		color := l.Color
		if color == "" {
			color = "#6366f1"
		}
		data, status, err := cwDo("POST",
			fmt.Sprintf("%s/api/v1/projects/%s/labels", baseURL, createdProj.Slug),
			hdrs,
			createLabelReq{Name: l.Name, Color: color},
		)
		if err != nil {
			return err
		}
		if status == http.StatusCreated || status == http.StatusOK {
			var lbl apiLabel
			if err := json.Unmarshal(data, &lbl); err == nil {
				labelIDMap[l.Name] = lbl.ID
				// Record it so a second source label that only differs by
				// case (e.g. "Bug" and "bug" on different cards) reuses the
				// same label instead of creating another duplicate.
				existingLabels[strings.ToLower(lbl.Name)] = lbl
			}
		}
	}

	// ── columns and cards ─────────────────────────────────────────────────────
	//
	// CreateProject auto-populates every new project with default columns
	// (from the "default_columns" system setting). Reuse any of those that
	// happen to match a migrated column name instead of creating a duplicate,
	// and remove whichever defaults end up unused once migration is done.
	columnsURL := fmt.Sprintf("%s/api/v1/projects/%s/columns", baseURL, createdProj.Slug)
	existingCols := map[string]apiColumn{}
	if data, status, err := cwDo("GET", columnsURL, hdrs, nil); err == nil && status == http.StatusOK {
		var cols []apiColumn
		if json.Unmarshal(data, &cols) == nil {
			for _, c := range cols {
				existingCols[c.Name] = c
			}
		}
	}
	usedColumnNames := map[string]bool{}
	// Tracks which resolved users have already been added as a project
	// member, so a user with many assigned cards only triggers one POST.
	addedMembers := map[uint]bool{}
	// Tracks temporary API keys minted so imported comments can be
	// attributed to their real author; revoked once the import is done.
	tempKeys := map[uint]tempAPIKey{}
	defer func() {
		for userID, tk := range tempKeys {
			revokeTempAPIKey(baseURL, token, userID, tk.ID)
		}
	}()

	for _, col := range p.Columns {
		colName := MapColumn(col.Name, columnMap)
		usedColumnNames[colName] = true

		createdCol, ok := existingCols[colName]
		if !ok {
			type createColReq struct {
				Name string `json:"name"`
			}
			data, status, err := cwDo("POST", columnsURL, hdrs, createColReq{Name: colName})
			if err != nil {
				return err
			}
			if status != http.StatusCreated && status != http.StatusOK {
				return fmt.Errorf("create column %q (HTTP %d): %s", colName, status, string(data))
			}
			if err := json.Unmarshal(data, &createdCol); err != nil {
				return fmt.Errorf("parse created column: %w", err)
			}
			fmt.Printf("  → column %q\n", colName)
		} else {
			fmt.Printf("  → column %q (reusing default)\n", colName)
		}

		for _, card := range col.Cards {
			if err := writeCard(baseURL, token, createdProj.Slug, createdCol.ID, card, labelIDMap, userMap, addedMembers, tempKeys); err != nil {
				return fmt.Errorf("write card %q: %w", card.Title, err)
			}
		}
	}

	// Remove default columns that weren't used by any migrated column.
	for name, col := range existingCols {
		if usedColumnNames[name] {
			continue
		}
		delURL := fmt.Sprintf("%s/%d", columnsURL, col.ID)
		if data, status, err := cwDo("DELETE", delURL, hdrs, nil); err != nil || status != http.StatusOK {
			fmt.Printf("  ⚠ could not remove unused default column %q (HTTP %d): %s\n", name, status, string(data))
			continue
		}
		fmt.Printf("  → removed unused default column %q\n", name)
	}

	// ── topics ────────────────────────────────────────────────────────────────
	for _, topic := range p.Topics {
		if err := writeTopic(baseURL, token, createdProj.Slug, topic, userMap, addedMembers, tempKeys); err != nil {
			return fmt.Errorf("write topic %q: %w", topic.Title, err)
		}
	}

	return nil
}

// uploadAttachment downloads data from a source URL and uploads it to
// WarmDesk as an attachment owned by the given entity (owner_type "card",
// "card_comment", etc.). The source URL is fetched with a bare GET — Ryver's
// file URLs are pre-signed and reject a Bearer/Authorization header (see
// pyryver's File.download_data(), which deliberately omits it).
func uploadAttachment(baseURL string, authHdrs map[string]string, ownerType string, ownerID uint, filename, sourceURL string) error {
	srcResp, err := http.Get(sourceURL)
	if err != nil {
		return fmt.Errorf("download %q: %w", sourceURL, err)
	}
	defer srcResp.Body.Close()
	if srcResp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %q: HTTP %d", sourceURL, srcResp.StatusCode)
	}
	data, err := io.ReadAll(srcResp.Body)
	if err != nil {
		return fmt.Errorf("read %q: %w", sourceURL, err)
	}

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	if err := w.WriteField("owner_type", ownerType); err != nil {
		return err
	}
	if err := w.WriteField("owner_id", fmt.Sprintf("%d", ownerID)); err != nil {
		return err
	}
	fw, err := w.CreateFormFile("file", filename)
	if err != nil {
		return err
	}
	if _, err := fw.Write(data); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", baseURL+"/api/v1/attachments", &body)
	if err != nil {
		return err
	}
	for k, v := range authHdrs {
		req.Header.Set(k, v)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		respData, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload attachment (HTTP %d): %s", resp.StatusCode, string(respData))
	}
	return nil
}

// ensureProjectMember adds u as a project member (role "member") the first
// time it's asked to for a given user, via the safe additive /members
// endpoint. A 409 (already a member) is treated as success.
func ensureProjectMember(baseURL, token, slug string, u ResolvedUser, addedMembers map[uint]bool) {
	if addedMembers[u.ID] {
		return
	}
	addedMembers[u.ID] = true
	hdrs := map[string]string{"Authorization": "Bearer " + token}
	type addMemberReq struct {
		Login string `json:"login"`
		Role  string `json:"role"`
	}
	data, status, err := cwDo("POST",
		fmt.Sprintf("%s/api/v1/projects/%s/members", baseURL, slug),
		hdrs,
		addMemberReq{Login: u.Username, Role: "member"},
	)
	if err != nil || (status != http.StatusCreated && status != http.StatusOK && status != http.StatusConflict) {
		fmt.Printf("  ⚠ could not add %q as a project member (HTTP %d): %s\n", u.Username, status, string(data))
	}
}

// tempAPIKey is a short-lived, unscoped API key minted on behalf of another
// user purely so a migration-created comment can be attributed to its real
// author. CreateComment always attributes to whoever's credentials made the
// request — there's no "author override" field — so impersonating them via
// their own (temporary) key is the only way to get correct authorship.
type tempAPIKey struct {
	ID  uint
	Key string
}

// mintTempAPIKey creates a temporary personal API key for userID, via the
// same admin capability exposed in Admin → Users → API Keys.
func mintTempAPIKey(baseURL, token string, userID uint) (tempAPIKey, error) {
	hdrs := map[string]string{"Authorization": "Bearer " + token}
	type createKeyReq struct {
		Name string `json:"name"`
	}
	data, status, err := cwDo("POST",
		fmt.Sprintf("%s/api/v1/admin/users/%d/api-keys", baseURL, userID),
		hdrs,
		createKeyReq{Name: "ryver-migration (temporary)"},
	)
	if err != nil {
		return tempAPIKey{}, err
	}
	if status != http.StatusCreated {
		return tempAPIKey{}, fmt.Errorf("HTTP %d: %s", status, string(data))
	}
	var resp struct {
		ID  uint   `json:"id"`
		Key string `json:"key"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return tempAPIKey{}, fmt.Errorf("parse created api key: %w", err)
	}
	return tempAPIKey{ID: resp.ID, Key: resp.Key}, nil
}

// revokeTempAPIKey deletes a key minted by mintTempAPIKey. Best-effort —
// an import that already wrote everything shouldn't fail on cleanup.
func revokeTempAPIKey(baseURL, token string, userID, keyID uint) {
	hdrs := map[string]string{"Authorization": "Bearer " + token}
	cwDo("DELETE", //nolint:errcheck
		fmt.Sprintf("%s/api/v1/admin/users/%d/api-keys/%d", baseURL, userID, keyID),
		hdrs, nil,
	)
}

// authorHeaders returns the HTTP headers to use so a comment/reply/topic
// gets created as its real author (via a temporary API key for the
// user_map-resolved WarmDesk account), and whether that succeeded. Falls
// back to the importer's own headers (defaultHdrs) when the author isn't
// resolvable or a key couldn't be minted — callers should then fall back to
// the old text-prefix attribution.
func authorHeaders(baseURL, token, slug, author string, userMap map[string]ResolvedUser, addedMembers map[uint]bool, tempKeys map[uint]tempAPIKey, defaultHdrs map[string]string) (map[string]string, bool) {
	u, ok := userMap[author]
	if !ok {
		return defaultHdrs, false
	}
	ensureProjectMember(baseURL, token, slug, u, addedMembers)
	tk, minted := tempKeys[u.ID]
	if !minted {
		var err error
		tk, err = mintTempAPIKey(baseURL, token, u.ID)
		if err != nil {
			fmt.Printf("  ⚠ could not create a temporary API key for %q — attributed to the importer instead: %v\n", u.Username, err)
			return defaultHdrs, false
		}
		tempKeys[u.ID] = tk
	}
	if tk.Key == "" {
		return defaultHdrs, false
	}
	return map[string]string{"X-API-Key": tk.Key}, true
}

func writeCard(baseURL, token, slug string, columnID uint, card Card, labelIDMap map[string]uint, userMap map[string]ResolvedUser, addedMembers map[uint]bool, tempKeys map[uint]tempAPIKey) error {
	hdrs := map[string]string{"Authorization": "Bearer " + token}

	type createCardReq struct {
		Title       string  `json:"title"`
		Description string  `json:"description"`
		Priority    string  `json:"priority"`
		StartDate   *string `json:"start_date,omitempty"`
		DueDate     *string `json:"due_date,omitempty"`
		AssigneeID  *uint   `json:"assignee_id,omitempty"`
	}
	req := createCardReq{
		Title:       card.Title,
		Description: card.Description,
		Priority:    card.Priority,
	}
	if card.StartDate != "" {
		req.StartDate = &card.StartDate
	}
	if card.DueDate != "" {
		req.DueDate = &card.DueDate
	}
	// The primary "Assignee" shown on a card is a separate field from the
	// (multiple) assignees list below — set it to the source item's
	// creator, so a card always shows *some* assignee even when the source
	// had no explicit assignees, as long as the creator is a mapped user.
	if creator, ok := userMap[card.Creator]; ok {
		ensureProjectMember(baseURL, token, slug, creator, addedMembers)
		req.AssigneeID = &creator.ID
	}

	data, status, err := cwDo("POST",
		fmt.Sprintf("%s/api/v1/projects/%s/columns/%d/cards", baseURL, slug, columnID),
		hdrs, req,
	)
	if err != nil {
		return err
	}
	if status != http.StatusCreated && status != http.StatusOK {
		return fmt.Errorf("create card (HTTP %d): %s", status, string(data))
	}
	var created apiCard
	if err := json.Unmarshal(data, &created); err != nil {
		return fmt.Errorf("parse created card: %w", err)
	}
	fmt.Printf("    → card %q\n", card.Title)

	// Close the card if the source marked it done — CreateCard has no
	// "closed" field, so this has to be a separate update.
	if card.Closed {
		type updateCardReq struct {
			Closed *bool `json:"closed"`
		}
		closed := true
		cwDo("PUT", //nolint:errcheck
			fmt.Sprintf("%s/api/v1/projects/%s/cards/%d", baseURL, slug, created.ID),
			hdrs,
			updateCardReq{Closed: &closed},
		)
	}

	// Assign users, resolved via the configured user_map. A user must be a
	// project member before they can meaningfully see a card assigned to
	// them, so add them as one (role "member") the first time they come up.
	for _, name := range card.Assignees {
		u, ok := userMap[name]
		if !ok {
			continue
		}
		ensureProjectMember(baseURL, token, slug, u, addedMembers)
		url := fmt.Sprintf("%s/api/v1/projects/%s/cards/%d/assignees/%d", baseURL, slug, created.ID, u.ID)
		cwDo("POST", url, hdrs, nil) //nolint:errcheck
	}

	// Assign labels
	for _, l := range card.Labels {
		if labelID, ok := labelIDMap[l.Name]; ok {
			url := fmt.Sprintf("%s/api/v1/projects/%s/cards/%d/labels/%d", baseURL, slug, created.ID, labelID)
			cwDo("POST", url, hdrs, nil) //nolint:errcheck
		}
	}

	// Tags — WarmDesk has no separate project-level tag list; AddCardTag
	// creates the tag (scoped to this card) on first use, so attaching it
	// here is the entire "add to the project" step.
	for _, tag := range card.Tags {
		type addTagReq struct {
			Name string `json:"name"`
		}
		cwDo("POST", //nolint:errcheck
			fmt.Sprintf("%s/api/v1/projects/%s/cards/%d/tags", baseURL, slug, created.ID),
			hdrs,
			addTagReq{Name: tag},
		)
	}

	// Checklist
	for _, item := range card.Checklist {
		type checkReq struct {
			Body        string `json:"body"`
			IsCompleted bool   `json:"is_completed"`
		}
		cwDo("POST", //nolint:errcheck
			fmt.Sprintf("%s/api/v1/projects/%s/cards/%d/checklist", baseURL, slug, created.ID),
			hdrs,
			checkReq{Body: item.Text, IsCompleted: item.Done},
		)
	}

	// Attachments — downloaded from the source platform and re-uploaded
	// into WarmDesk's own storage; a source URL alone isn't enough since
	// it may require the source platform's own auth, expire, or simply not
	// be reachable by anyone who later opens the card in WarmDesk.
	for _, att := range card.Attachments {
		if err := uploadAttachment(baseURL, hdrs, "card", created.ID, att.Filename, att.URL); err != nil {
			fmt.Printf("  ⚠ could not import attachment %q: %v\n", att.Filename, err)
		}
	}

	// Comments — attributed to their real author when resolvable via
	// user_map (by posting through a temporary key minted for that user),
	// falling back to a text prefix under the importer's own name otherwise.
	for _, cm := range card.Comments {
		commentHdrs, attributed := authorHeaders(baseURL, token, slug, cm.Author, userMap, addedMembers, tempKeys, hdrs)
		body := cm.Body
		if !attributed && cm.Author != "" {
			body = fmt.Sprintf("*[%s]* %s", cm.Author, cm.Body)
		}
		// A comment that's just an attached image has no text at all —
		// CreateComment accepts an empty body for exactly this case.

		type commentReq struct {
			Body string `json:"body"`
		}
		cData, cStatus, cErr := cwDo("POST",
			fmt.Sprintf("%s/api/v1/projects/%s/cards/%d/comments", baseURL, slug, created.ID),
			commentHdrs,
			commentReq{Body: body},
		)
		if cErr != nil || (cStatus != http.StatusCreated && cStatus != http.StatusOK) {
			fmt.Printf("  ⚠ could not create comment (HTTP %d): %s\n", cStatus, string(cData))
			continue
		}
		if len(cm.Attachments) == 0 {
			continue
		}
		var createdComment struct {
			ID uint `json:"id"`
		}
		if err := json.Unmarshal(cData, &createdComment); err != nil {
			continue
		}
		for _, att := range cm.Attachments {
			if err := uploadAttachment(baseURL, commentHdrs, "card_comment", createdComment.ID, att.Filename, att.URL); err != nil {
				fmt.Printf("  ⚠ could not import comment attachment %q: %v\n", att.Filename, err)
			}
		}
	}

	return nil
}

func writeTopic(baseURL, token, slug string, topic Topic, userMap map[string]ResolvedUser, addedMembers map[uint]bool, tempKeys map[uint]tempAPIKey) error {
	hdrs := map[string]string{"Authorization": "Bearer " + token}

	type createTopicReq struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	topicHdrs, attributed := authorHeaders(baseURL, token, slug, topic.Author, userMap, addedMembers, tempKeys, hdrs)
	body := topic.Body
	if !attributed && topic.Author != "" {
		body = fmt.Sprintf("*[%s]* %s", topic.Author, topic.Body)
	}
	data, status, err := cwDo("POST",
		fmt.Sprintf("%s/api/v1/projects/%s/topics", baseURL, slug),
		topicHdrs,
		createTopicReq{Title: topic.Title, Body: body},
	)
	if err != nil {
		return err
	}
	if status != http.StatusCreated && status != http.StatusOK {
		return fmt.Errorf("create topic (HTTP %d): %s", status, string(data))
	}

	var created struct {
		ID uint `json:"id"`
	}
	if err := json.Unmarshal(data, &created); err != nil {
		return fmt.Errorf("parse created topic: %w", err)
	}
	fmt.Printf("  → topic %q\n", topic.Title)

	// Replies — same real-author attribution as topic creation and card comments.
	for _, reply := range topic.Replies {
		replyHdrs, rAttributed := authorHeaders(baseURL, token, slug, reply.Author, userMap, addedMembers, tempKeys, hdrs)
		replyBody := reply.Body
		if !rAttributed && reply.Author != "" {
			replyBody = fmt.Sprintf("*[%s]* %s", reply.Author, reply.Body)
		}
		type replyReq struct {
			Body string `json:"body"`
		}
		cwDo("POST", //nolint:errcheck
			fmt.Sprintf("%s/api/v1/projects/%s/topics/%d/replies", baseURL, slug, created.ID),
			replyHdrs,
			replyReq{Body: replyBody},
		)
	}

	return nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// slugify converts a project name to a URL-safe slug.
func slugify(name string) string {
	s := strings.ToLower(name)
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else if r == ' ' || r == '-' || r == '_' {
			b.WriteByte('-')
		}
	}
	// Trim leading/trailing dashes
	result := strings.Trim(b.String(), "-")
	// Collapse consecutive dashes
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}
	return result
}

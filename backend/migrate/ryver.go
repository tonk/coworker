package migrate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// ─── Ryver HTTP helper ────────────────────────────────────────────────────────
//
// Ryver exposes an OData-based REST API at https://{org}.ryver.com/api/1/odata.svc
// Authentication uses a Bearer token obtained from the Ryver admin console.
//
// Entity set names in Ryver's OData service are lowercase (workrooms, posts,
// postComments, tasks, taskBoards, taskComments) — verified against Ryver's
// developer docs (https://api.ryver.com/ryvrest_api_examples.html) and the
// pyryver reference client. Bound action names (Chat.PostMessage(),
// Post.PostCreateTopic(), TaskBoard.Create(), ...) keep their PascalCase form.
//
// Tasks are not queried directly by workroom — a team/forum only has a task
// board if one has been set up (GET workrooms(id)/board, 404 if none), and
// tasks are listed from that board (GET taskBoards(id)/tasks).

func ryverDo(method, rawURL, token string, body interface{}) ([]byte, int, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, rawURL, bodyReader)
	if err != nil {
		return nil, 0, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("http %s %s: %w", method, rawURL, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response: %w", err)
	}
	return data, resp.StatusCode, nil
}

func ryverBase(org string) string {
	return fmt.Sprintf("https://%s.ryver.com/api/1/odata.svc", org)
}

// ryverPaginate fetches every page of an OData collection query (Ryver caps
// each page at 50 results regardless of a larger $top) and returns the raw
// "results" entries so callers can unmarshal them into the shape they need.
func ryverPaginate(token, url string) ([]json.RawMessage, error) {
	const pageSize = 50
	var all []json.RawMessage
	skip := 0
	for {
		sep := "&"
		if !strings.Contains(url, "?") {
			sep = "?"
		}
		pageURL := fmt.Sprintf("%s%s$skip=%d&$top=%d", url, sep, skip, pageSize)
		data, status, err := ryverDo("GET", pageURL, token, nil)
		if err != nil {
			return nil, err
		}
		if status != http.StatusOK {
			return nil, fmt.Errorf("HTTP %d: %s", status, string(data))
		}

		var page struct {
			D struct {
				Results []json.RawMessage `json:"results"`
			} `json:"d"`
		}
		if err := json.Unmarshal(data, &page); err != nil {
			return nil, fmt.Errorf("parse page: %w", err)
		}
		all = append(all, page.D.Results...)
		if len(page.D.Results) < pageSize {
			break
		}
		skip += len(page.D.Results)
	}
	return all, nil
}

// ryverGetTaskBoardID looks up the task board attached to a team/forum.
// Ryver returns 404 when tasks have never been set up for that team.
func ryverGetTaskBoardID(base, token string, teamID int) (int, bool, error) {
	url := fmt.Sprintf("%s/workrooms(%d)/board", base, teamID)
	data, status, err := ryverDo("GET", url, token, nil)
	if err != nil {
		return 0, false, err
	}
	if status == http.StatusNotFound {
		return 0, false, nil
	}
	if status != http.StatusOK {
		return 0, false, fmt.Errorf("HTTP %d: %s", status, string(data))
	}

	var resp struct {
		D struct {
			Results struct {
				ID int `json:"id"`
			} `json:"results"`
		} `json:"d"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return 0, false, fmt.Errorf("parse task board: %w", err)
	}
	return resp.D.Results.ID, true, nil
}

// ryverCreateTaskBoard sets up a task board for a team/forum that doesn't
// have one yet, as a "board" type (categories, i.e. WarmDesk-style columns)
// rather than a flat "list". Unlike regular entity creates, this bound action
// returns the created object directly, without a "d.results" wrapper.
func ryverCreateTaskBoard(base, token string, teamID int) (int, error) {
	url := fmt.Sprintf("%s/workrooms(%d)/TaskBoard.Create()", base, teamID)
	body := map[string]interface{}{
		"board": map[string]interface{}{
			"type": "board",
		},
	}
	data, status, err := ryverDo("POST", url, token, body)
	if err != nil {
		return 0, err
	}
	if status != http.StatusOK && status != http.StatusCreated {
		return 0, fmt.Errorf("HTTP %d: %s", status, string(data))
	}

	var resp struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return 0, fmt.Errorf("parse created task board: %w", err)
	}
	return resp.ID, nil
}

// ryverListCategories returns the categories (columns) already set up on a
// task board.
func ryverListCategories(token, base string, boardID int) ([]struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}, error) {
	url := fmt.Sprintf("%s/taskBoards(%d)/categories", base, boardID)
	raw, err := ryverPaginate(token, url)
	if err != nil {
		return nil, err
	}
	cats := make([]struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}, 0, len(raw))
	for _, r := range raw {
		var c struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}
		if err := json.Unmarshal(r, &c); err != nil {
			continue
		}
		cats = append(cats, c)
	}
	return cats, nil
}

// ryverCreateCategory creates a new category (column) on a task board.
func ryverCreateCategory(base, token string, boardID int, name string) (int, error) {
	url := fmt.Sprintf("%s/taskCategories", base)
	body := map[string]interface{}{
		"board": map[string]interface{}{"id": boardID},
		"name":  name,
	}
	data, status, err := ryverDo("POST", url, token, body)
	if err != nil {
		return 0, err
	}
	if status != http.StatusOK && status != http.StatusCreated {
		return 0, fmt.Errorf("HTTP %d: %s", status, string(data))
	}

	var resp struct {
		D struct {
			Results struct {
				ID int `json:"id"`
			} `json:"results"`
		} `json:"d"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return 0, fmt.Errorf("parse created category: %w", err)
	}
	return resp.D.Results.ID, nil
}

// ─── Export to Ryver ──────────────────────────────────────────────────────────

// ExportToRyver exports a canonical project to Ryver as tasks in a team's task
// board, plus forum topics.
//
// Ryver represents Kanban-style columns as task board "categories" (not
// tags). The WarmDesk column name (translated via columnMap) is resolved to
// an existing category or created as a new one, and each task is filed under
// its category.
func ExportToRyver(cfg PlatformConfig, p *Project, columnMap map[string]string) error {
	base := ryverBase(cfg.Org)
	token := cfg.APIToken

	// Find the team by name
	teamID, err := ryverFindTeam(base, token, cfg.Team)
	if err != nil {
		return fmt.Errorf("find team %q: %w", cfg.Team, err)
	}
	fmt.Printf("  → team id: %d\n", teamID)

	boardID, hasBoard, err := ryverGetTaskBoardID(base, token, teamID)
	if err != nil {
		return fmt.Errorf("get task board: %w", err)
	}
	if !hasBoard {
		boardID, err = ryverCreateTaskBoard(base, token, teamID)
		if err != nil {
			return fmt.Errorf("create task board: %w", err)
		}
		fmt.Printf("  → created task board id: %d\n", boardID)
	}

	existingCats, err := ryverListCategories(token, base, boardID)
	if err != nil {
		return fmt.Errorf("list categories: %w", err)
	}
	categoryIDs := map[string]int{}
	for _, c := range existingCats {
		categoryIDs[c.Name] = c.ID
	}

	for _, col := range p.Columns {
		categoryName := MapColumn(col.Name, columnMap)
		categoryID, ok := categoryIDs[categoryName]
		if !ok {
			categoryID, err = ryverCreateCategory(base, token, boardID, categoryName)
			if err != nil {
				return fmt.Errorf("create category %q: %w", categoryName, err)
			}
			categoryIDs[categoryName] = categoryID
			fmt.Printf("  → created category %q (id=%d)\n", categoryName, categoryID)
		}

		for _, card := range col.Cards {
			fmt.Printf("  → exporting card %s: %s\n", card.Ref, card.Title)

			// Build body: description + checklist
			body := card.Description
			if len(card.Checklist) > 0 {
				body += "\n\n**Checklist:**\n"
				for _, item := range card.Checklist {
					check := "[ ]"
					if item.Done {
						check = "[x]"
					}
					body += fmt.Sprintf("- %s %s\n", check, item.Text)
				}
			}
			if card.TimeMinutes > 0 {
				h := card.TimeMinutes / 60
				m := card.TimeMinutes % 60
				body += fmt.Sprintf("\n⏱ Time spent: %d:%02d\n", h, m)
			}

			// Build tags: card tags + labels (columns are categories, not tags)
			tags := append([]string{}, card.Tags...)
			for _, l := range card.Labels {
				tags = append(tags, l.Name)
			}

			taskBody := map[string]interface{}{
				"board": map[string]interface{}{
					"id": boardID,
				},
				"category": map[string]interface{}{
					"id": categoryID,
				},
				"subject": card.Title,
				"body":    body,
				"tags":    tags,
			}
			if card.DueDate != "" {
				taskBody["dueDate"] = card.DueDate + "T00:00:00Z"
			}
			if card.Closed {
				// There is no "isComplete" flag on the Task entity — a task
				// is marked done by setting its completeDate.
				taskBody["completeDate"] = time.Now().UTC().Format(time.RFC3339)
			}

			taskURL := fmt.Sprintf("%s/tasks", base)
			data, status, err := ryverDo("POST", taskURL, token, taskBody)
			if err != nil {
				return fmt.Errorf("create task %q: %w", card.Title, err)
			}
			if status != http.StatusOK && status != http.StatusCreated {
				fmt.Printf("    ⚠ could not export %q (HTTP %d): %s\n", card.Title, status, string(data))
				continue
			}

			// Extract created entity id for comments
			var created struct {
				D struct {
					Results struct {
						ID int `json:"id"`
					} `json:"results"`
				} `json:"d"`
			}
			var taskID int
			if err := json.Unmarshal(data, &created); err == nil {
				taskID = created.D.Results.ID
			}

			// Post comments
			if taskID > 0 {
				for _, cm := range card.Comments {
					commentBody := cm.Body
					if cm.Author != "" {
						commentBody = fmt.Sprintf("*[%s]* %s", cm.Author, cm.Body)
					}
					commentURL := fmt.Sprintf("%s/taskComments?$format=json", base)
					ryverDo("POST", commentURL, token, map[string]interface{}{ //nolint:errcheck
						"comment": commentBody,
						"task":    map[string]interface{}{"id": taskID},
					})
				}
			}
		}
	}

	// Export topics as forum posts
	for _, topic := range p.Topics {
		body := topic.Body
		if topic.Author != "" {
			body = fmt.Sprintf("*[%s]* %s", topic.Author, topic.Body)
		}
		topicURL := fmt.Sprintf("%s/workrooms(%d)/Post.PostCreateTopic()", base, teamID)
		data, status, err := ryverDo("POST", topicURL, token, map[string]interface{}{
			"subject": topic.Title,
			"body":    body,
		})
		if err != nil || (status != http.StatusOK && status != http.StatusCreated) {
			fmt.Printf("  ⚠ could not export topic %q\n", topic.Title)
			continue
		}
		fmt.Printf("  → topic %q\n", topic.Title)

		// Bound actions like Post.PostCreateTopic() may return either the
		// standard "d.results" wrapper or the created object directly —
		// try both rather than assuming one shape.
		var wrapped struct {
			D struct {
				Results struct {
					ID int `json:"id"`
				} `json:"results"`
			} `json:"d"`
		}
		var direct struct {
			ID int `json:"id"`
		}
		topicID := 0
		if err := json.Unmarshal(data, &wrapped); err == nil && wrapped.D.Results.ID > 0 {
			topicID = wrapped.D.Results.ID
		} else if err := json.Unmarshal(data, &direct); err == nil {
			topicID = direct.ID
		}

		if topicID > 0 {
			for _, reply := range topic.Replies {
				replyBody := reply.Body
				if reply.Author != "" {
					replyBody = fmt.Sprintf("*[%s]* %s", reply.Author, reply.Body)
				}
				replyURL := fmt.Sprintf("%s/postComments?$format=json", base)
				ryverDo("POST", replyURL, token, map[string]interface{}{ //nolint:errcheck
					"comment": replyBody,
					"post":    map[string]interface{}{"id": topicID},
				})
			}
		}
	}

	fmt.Printf("✓ export to Ryver complete\n")
	return nil
}

// ─── Import from Ryver ────────────────────────────────────────────────────────

type ryverTaskJSON struct {
	ID      int    `json:"id"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
	// A task is done iff completeDate is set — there is no separate
	// "isComplete" field on the Task entity.
	CompleteDate string   `json:"completeDate"`
	CreateDate   string   `json:"createDate"`
	DueDate      string   `json:"dueDate"`
	Tags         []string `json:"tags"`
	Category     *struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"category"`
	CreateUser struct {
		Username    string `json:"username"`
		DisplayName string `json:"displayName"`
	} `json:"createUser"`
	// Assignees is a to-many navigation property, expanded inline as a
	// {"results": [...]} collection like other list responses.
	Assignees struct {
		Results []struct {
			Username    string `json:"username"`
			DisplayName string `json:"displayName"`
		} `json:"results"`
	} `json:"assignees"`
}

// ryverAttachmentJSON is a File entity as returned by tasks(id)/attachments —
// its download URL and MIME type sit directly on it; the nested "storage"
// object duplicates the same info and isn't needed.
type ryverAttachmentJSON struct {
	RecordType string `json:"recordType"` // "file" | "link"
	FileName   string `json:"fileName"`
	Title      string `json:"title"`
	URL        string `json:"url"`
	Type       string `json:"type"` // MIME type
}

// ryverFetchAttachments fetches the attachments of a Ryver entity (e.g.
// entityPath "tasks" or "taskComments") as their own paginated collection —
// see ImportFromRyver's task-attachments comment for why an inline $expand
// on the parent query isn't reliable for this. "link" attachments (external
// bookmarks, not uploaded files) are skipped since there's nothing to
// download and re-host.
func ryverFetchAttachments(token, base, entityPath string, entityID int) []Attachment {
	url := fmt.Sprintf("%s/%s(%d)/attachments", base, entityPath, entityID)
	raw, err := ryverPaginate(token, url)
	if err != nil {
		return nil
	}
	var attachments []Attachment
	for _, araw := range raw {
		var a ryverAttachmentJSON
		if err := json.Unmarshal(araw, &a); err != nil {
			continue
		}
		if a.RecordType != "file" {
			continue
		}
		filename := a.FileName
		if filename == "" {
			filename = a.Title
		}
		attachments = append(attachments, Attachment{Filename: filename, URL: a.URL, MimeType: a.Type})
	}
	return attachments
}

type ryverTopicJSON struct {
	ID         int    `json:"id"`
	Subject    string `json:"subject"`
	Body       string `json:"body"`
	CreateUser struct {
		Username    string `json:"username"`
		DisplayName string `json:"displayName"`
	} `json:"createUser"`
}

type ryverCommentJSON struct {
	ID int `json:"id"`
	// The reply/comment text is stored under "comment", not "body".
	Comment    string `json:"comment"`
	CreateUser struct {
		Username    string `json:"username"`
		DisplayName string `json:"displayName"`
	} `json:"createUser"`
}

// ryverChatMessageJSON is a raw chat message as returned by
// workrooms(id)/Chat.History(). Unlike other entities, its author is only a
// bare user id — there is no inline createUser object to read a name from.
type ryverChatMessageJSON struct {
	When    string `json:"when"`
	Subtype string `json:"subtype"` // empty for a regular chat message
	Body    string `json:"body"`
	From    struct {
		ID int `json:"id"`
	} `json:"from"`
	// At most one attached file or link; Ryver represents both the same way.
	Extras *struct {
		File *struct {
			FileName string `json:"fileName"`
			URL      string `json:"url"`
			Type     string `json:"type"`
		} `json:"file"`
	} `json:"extras"`
}

// ryverBuildUserMap fetches every user in the Ryver organization once and
// returns a userID -> display name (falling back to username) lookup, used
// to resolve chat message authors.
func ryverBuildUserMap(base, token string) (map[int]string, error) {
	raw, err := ryverPaginate(token, base+"/users?$select=id,username,displayName")
	if err != nil {
		return nil, err
	}
	byID := make(map[int]string, len(raw))
	for _, r := range raw {
		var u struct {
			ID          int    `json:"id"`
			Username    string `json:"username"`
			DisplayName string `json:"displayName"`
		}
		if err := json.Unmarshal(r, &u); err != nil {
			continue
		}
		name := u.DisplayName
		if name == "" {
			name = u.Username
		}
		byID[u.ID] = name
	}
	return byID, nil
}

// ImportFromRyver reads tasks and topics from a Ryver team and returns the
// canonical project representation.
func ImportFromRyver(cfg PlatformConfig, columnMap map[string]string, includeCards, includeChat bool) (*Project, error) {
	base := ryverBase(cfg.Org)
	token := cfg.APIToken
	reverseMap := ReverseColumnMap(columnMap)

	// Find team
	teamID, err := ryverFindTeam(base, token, cfg.Team)
	if err != nil {
		return nil, fmt.Errorf("find team %q: %w", cfg.Team, err)
	}

	proj := &Project{Name: cfg.Team}
	colIndex := map[string]*Column{}

	if includeCards {
		// Tasks live on the team's task board, if one has been set up.
		boardID, hasBoard, err := ryverGetTaskBoardID(base, token, teamID)
		if err != nil {
			return nil, fmt.Errorf("get task board: %w", err)
		}
		if hasBoard {
			tasksURL := fmt.Sprintf("%s/taskBoards(%d)/tasks?$expand=createUser,category,assignees&$filter=(archived+eq+false+and+parent+eq+null)", base, boardID)
			rawTasks, err := ryverPaginate(token, tasksURL)
			if err != nil {
				return nil, fmt.Errorf("get tasks: %w", err)
			}

			for _, raw := range rawTasks {
				var t ryverTaskJSON
				if err := json.Unmarshal(raw, &t); err != nil {
					continue
				}

				// The task's category is the Ryver equivalent of a WarmDesk column.
				colName := "Tasks"
				if t.Category != nil && t.Category.Name != "" {
					colName = MapColumnReverse(t.Category.Name, reverseMap)
				}

				// Ryver's #-style tags map to WarmDesk's colored Labels, not its
				// free-form per-card Tags — Labels is the closer conceptual fit
				// for a category-of-work marker like "feature" or "bug".
				labels := make([]Label, 0, len(t.Tags))
				for _, tag := range t.Tags {
					tag = strings.TrimSpace(tag)
					if tag != "" {
						labels = append(labels, Label{Name: tag})
					}
				}

				col, ok := colIndex[colName]
				if !ok {
					proj.Columns = append(proj.Columns, Column{Name: colName})
					col = &proj.Columns[len(proj.Columns)-1]
					colIndex[colName] = col
				}

				dueDate := ""
				if len(t.DueDate) >= 10 {
					dueDate = t.DueDate[:10]
				}
				startDate := ""
				if len(t.CreateDate) >= 10 {
					startDate = t.CreateDate[:10]
				}

				// A task is closed if it has been marked complete, or if it sits
				// in a category literally named "Completed" (some boards use a
				// category instead of the completion checkbox to track this).
				closed := t.CompleteDate != ""
				if t.Category != nil && strings.EqualFold(t.Category.Name, "completed") {
					closed = true
				}

				var comments []Comment
				commentsURL := fmt.Sprintf("%s/tasks(%d)/comments?$format=json&$expand=createUser", base, t.ID)
				rawComments, err := ryverPaginate(token, commentsURL)
				if err == nil {
					for _, craw := range rawComments {
						var c ryverCommentJSON
						if err := json.Unmarshal(craw, &c); err != nil {
							continue
						}
						author := c.CreateUser.DisplayName
						if author == "" {
							author = c.CreateUser.Username
						}
						comments = append(comments, Comment{
							Author:      author,
							Body:        c.Comment,
							Attachments: ryverFetchAttachments(token, base, "taskComments", c.ID),
						})
					}
				}

				var assignees []string
				for _, a := range t.Assignees.Results {
					name := a.DisplayName
					if name == "" {
						name = a.Username
					}
					if name != "" {
						assignees = append(assignees, name)
					}
				}

				creator := t.CreateUser.DisplayName
				if creator == "" {
					creator = t.CreateUser.Username
				}

				// Fetched as its own paginated collection, not via inline
				// $expand on the tasks query — an expanded nav collection is
				// subject to its own (much smaller) server-side default page
				// size, which silently truncated tasks with more than one
				// attachment down to just the first.
				attachments := ryverFetchAttachments(token, base, "tasks", t.ID)

				col.Cards = append(col.Cards, Card{
					Title:       t.Subject,
					Description: t.Body,
					StartDate:   startDate,
					DueDate:     dueDate,
					Closed:      closed,
					Creator:     creator,
					Labels:      labels,
					Comments:    comments,
					Assignees:   assignees,
					Attachments: attachments,
				})
			}
		}

		// Get forum topics
		topicsURL := fmt.Sprintf("%s/workrooms(%d)/Post.Stream(archived=false)?$format=json&$expand=createUser", base, teamID)
		rawTopics, err := ryverPaginate(token, topicsURL)
		if err != nil {
			return nil, fmt.Errorf("get topics: %w", err)
		}

		for _, raw := range rawTopics {
			var t ryverTopicJSON
			if err := json.Unmarshal(raw, &t); err != nil {
				continue
			}
			author := t.CreateUser.DisplayName
			if author == "" {
				author = t.CreateUser.Username
			}
			topic := Topic{
				Title:  t.Subject,
				Body:   t.Body,
				Author: author,
			}

			// Get replies
			repliesURL := fmt.Sprintf("%s/posts(%d)/comments?$format=json&$expand=createUser", base, t.ID)
			rawReplies, err := ryverPaginate(token, repliesURL)
			if err == nil {
				for _, rraw := range rawReplies {
					var r ryverCommentJSON
					if err := json.Unmarshal(rraw, &r); err != nil {
						continue
					}
					rAuthor := r.CreateUser.DisplayName
					if rAuthor == "" {
						rAuthor = r.CreateUser.Username
					}
					topic.Replies = append(topic.Replies, TopicReply{
						Author: rAuthor,
						Body:   r.Comment,
					})
				}
			}
			proj.Topics = append(proj.Topics, topic)
		}
	} // includeCards

	if includeChat {
		// Team/group chat history. Unlike tasks/topics/comments, a raw chat
		// message only carries its author as a bare user ID ("from":{"id":N}),
		// not an inline createUser object — so resolve authors via one
		// org-wide user lookup instead of a request per message.
		usersByID, err := ryverBuildUserMap(base, token)
		if err != nil {
			return nil, fmt.Errorf("list users: %w", err)
		}
		messagesURL := fmt.Sprintf("%s/workrooms(%d)/Chat.History()?$format=json", base, teamID)
		rawMessages, err := ryverPaginate(token, messagesURL)
		if err != nil {
			return nil, fmt.Errorf("get chat history: %w", err)
		}
		for _, mraw := range rawMessages {
			var m ryverChatMessageJSON
			if err := json.Unmarshal(mraw, &m); err != nil {
				continue
			}
			// Skip system-generated "X created a task/topic" announcements —
			// SUBTYPE_CHAT_MESSAGE is represented as an absent subtype, not a
			// literal value, per Ryver's own API.
			if m.Subtype != "" {
				continue
			}
			// Chat message timestamps are a different format from task/topic
			// dates ("2022-06-01T08:01:27.588852Z" — fractional seconds, literal
			// "Z" for UTC) rather than e.g. "2021-09-22T15:07:53+0000".
			when, err := time.Parse(time.RFC3339Nano, m.When)
			if err != nil {
				continue
			}
			author := usersByID[m.From.ID]

			var attachment *Attachment
			if m.Extras != nil && m.Extras.File != nil && m.Extras.File.URL != "" {
				attachment = &Attachment{
					Filename: m.Extras.File.FileName,
					URL:      m.Extras.File.URL,
					MimeType: m.Extras.File.Type,
				}
			}

			proj.Messages = append(proj.Messages, ChatMessage{
				Author:     author,
				Body:       m.Body,
				CreatedAt:  when,
				Attachment: attachment,
			})
		}
	} // includeChat

	return proj, nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func ryverFindTeam(base, token, teamName string) (int, error) {
	url := fmt.Sprintf("%s/workrooms?$filter=name+eq+'%s'&$select=id,name", base, teamName)
	data, status, err := ryverDo("GET", url, token, nil)
	if err != nil {
		return 0, err
	}
	if status != http.StatusOK {
		return 0, fmt.Errorf("HTTP %d: %s", status, string(data))
	}

	var resp struct {
		D struct {
			Results []struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			} `json:"results"`
		} `json:"d"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return 0, fmt.Errorf("parse teams: %w", err)
	}
	if len(resp.D.Results) == 0 {
		return 0, fmt.Errorf("team %q not found", teamName)
	}
	return resp.D.Results[0].ID, nil
}

// DumpFirstTask fetches a task from a Ryver team's board — expanding
// createUser, category, assignees, and attachments (with their nested
// storage) — plus its comments and each comment's own attachments (fetched
// the same dedicated way ImportFromRyver does), and writes the combined raw
// JSON to path. Purely a debugging aid for inspecting what fields Ryver
// actually returns.
//
// If ref is non-empty, it must be the task's Ryver short reference (e.g.
// "CON-17", as shown in the Ryver UI) and that exact task is dumped.
// Otherwise it scans for a task that actually has an attachment
// (attachmentsCount > 0) and dumps that one if found, falling back to the
// first task otherwise — an empty attachments list wouldn't show what an
// attachment's shape is.
// DumpChatHistory fetches the first page (up to 20) of a Ryver team's raw,
// unfiltered chat history — no subtype filtering, no date parsing — and
// writes it to path. Purely a debugging aid for inspecting what a message
// actually looks like when ImportFromRyver reports zero messages
// unexpectedly (e.g. a subtype value or "when" timestamp format that
// doesn't match what was assumed).
func DumpChatHistory(cfg PlatformConfig, path string) error {
	base := ryverBase(cfg.Org)
	token := cfg.APIToken

	teamID, err := ryverFindTeam(base, token, cfg.Team)
	if err != nil {
		return fmt.Errorf("find team %q: %w", cfg.Team, err)
	}

	url := fmt.Sprintf("%s/workrooms(%d)/Chat.History()?$format=json&$top=20&$skip=0", base, teamID)
	data, status, err := ryverDo("GET", url, token, nil)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", status, string(data))
	}

	var pretty bytes.Buffer
	if err := json.Indent(&pretty, data, "", "  "); err != nil {
		return fmt.Errorf("indent json: %w", err)
	}
	return os.WriteFile(path, pretty.Bytes(), 0o644)
}

func DumpFirstTask(cfg PlatformConfig, ref, path string) error {
	base := ryverBase(cfg.Org)
	token := cfg.APIToken

	teamID, err := ryverFindTeam(base, token, cfg.Team)
	if err != nil {
		return fmt.Errorf("find team %q: %w", cfg.Team, err)
	}
	boardID, hasBoard, err := ryverGetTaskBoardID(base, token, teamID)
	if err != nil {
		return fmt.Errorf("get task board: %w", err)
	}
	if !hasBoard {
		return fmt.Errorf("team %q has no task board set up", cfg.Team)
	}

	var chosen json.RawMessage
	if ref != "" {
		tasksURL := fmt.Sprintf("%s/taskBoards(%d)/tasks?$filter=short+eq+'%s'&$expand=createUser,category,assignees,attachments,attachments/storage", base, boardID, ref)
		rawTasks, err := ryverPaginate(token, tasksURL)
		if err != nil {
			return fmt.Errorf("get task %q: %w", ref, err)
		}
		if len(rawTasks) == 0 {
			return fmt.Errorf("no task with short ref %q found on team %q's board", ref, cfg.Team)
		}
		chosen = rawTasks[0]
	} else {
		tasksURL := fmt.Sprintf("%s/taskBoards(%d)/tasks?$expand=createUser,category,assignees,attachments,attachments/storage", base, boardID)
		rawTasks, err := ryverPaginate(token, tasksURL)
		if err != nil {
			return fmt.Errorf("get tasks: %w", err)
		}
		if len(rawTasks) == 0 {
			return fmt.Errorf("team %q's task board has no tasks", cfg.Team)
		}

		chosen = rawTasks[0]
		for _, raw := range rawTasks {
			var t struct {
				AttachmentsCount int `json:"attachmentsCount"`
			}
			if json.Unmarshal(raw, &t) == nil && t.AttachmentsCount > 0 {
				chosen = raw
				break
			}
		}
	}

	var chosenID struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(chosen, &chosenID); err != nil {
		return fmt.Errorf("parse chosen task id: %w", err)
	}
	taskID := chosenID.ID

	// Comments, plus each comment's own attachments — fetched the exact
	// same dedicated way ImportFromRyver does, so this dump reflects
	// reality rather than an inline $expand that may behave differently.
	commentsURL := fmt.Sprintf("%s/tasks(%d)/comments?$format=json&$expand=createUser", base, taskID)
	rawComments, err := ryverPaginate(token, commentsURL)
	if err != nil {
		return fmt.Errorf("get comments: %w", err)
	}
	type dumpedComment struct {
		Comment     json.RawMessage `json:"comment"`
		Attachments []Attachment    `json:"resolved_attachments"`
	}
	dumpedComments := make([]dumpedComment, 0, len(rawComments))
	for _, craw := range rawComments {
		var c struct {
			ID int `json:"id"`
		}
		if err := json.Unmarshal(craw, &c); err != nil {
			continue
		}
		dumpedComments = append(dumpedComments, dumpedComment{
			Comment:     craw,
			Attachments: ryverFetchAttachments(token, base, "taskComments", c.ID),
		})
	}

	out := struct {
		Task     json.RawMessage `json:"task"`
		Comments []dumpedComment `json:"comments"`
	}{Task: chosen, Comments: dumpedComments}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal dump: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

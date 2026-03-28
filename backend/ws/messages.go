package ws

// Message types
const (
	// Client -> Server
	TypeChatSend    = "chat.send"
	TypeChatEdit    = "chat.edit"
	TypeChatDelete  = "chat.delete"
	TypePing        = "ping"

	// Server -> Client: chat
	TypeChatMessageCreated = "chat.message.created"
	TypeChatMessageUpdated = "chat.message.updated"
	TypeChatMessageDeleted = "chat.message.deleted"

	// Server -> Client: board
	TypeBoardCardCreated      = "board.card.created"
	TypeBoardCardUpdated      = "board.card.updated"
	TypeBoardCardMoved        = "board.card.moved"
	TypeBoardCardDeleted      = "board.card.deleted"
	TypeBoardColumnCreated    = "board.column.created"
	TypeBoardColumnUpdated    = "board.column.updated"
	TypeBoardColumnDeleted    = "board.column.deleted"
	TypeBoardColumnsReordered = "board.columns.reordered"
	TypeBoardCommentCreated   = "board.card.comment.created"
	TypeBoardCommentUpdated   = "board.card.comment.updated"
	TypeBoardCommentDeleted   = "board.card.comment.deleted"

	// Server -> Client: presence
	TypePresenceJoined = "presence.joined"
	TypePresenceLeft   = "presence.left"
	TypePresenceList   = "presence.list"

	// Server -> Client: reactions
	TypeChatReactionUpdated = "chat.reaction.updated"
	TypeDMReactionUpdated   = "dm.reaction.updated"

	// Server -> Client: DM message updates
	TypeDMMessageUpdated = "dm.message.updated"
	TypeDMMessageDeleted = "dm.message.deleted"

	// Server -> Client: topics
	TypeTopicCreated      = "topic.created"
	TypeTopicUpdated      = "topic.updated"
	TypeTopicDeleted      = "topic.deleted"
	TypeTopicReplyCreated = "topic.reply.created"
	TypeTopicReplyUpdated = "topic.reply.updated"
	TypeTopicReplyDeleted = "topic.reply.deleted"

	// System
	TypePong  = "pong"
	TypeError = "error"
)

type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
	ID      string      `json:"id,omitempty"`
}

type ChatSendPayload struct {
	Body string `json:"body"`
}

type ChatEditPayload struct {
	MessageID uint   `json:"message_id"`
	Body      string `json:"body"`
}

type ChatDeletePayload struct {
	MessageID uint `json:"message_id"`
}

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	ID      string `json:"id,omitempty"`
}

type PresenceUser struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
}

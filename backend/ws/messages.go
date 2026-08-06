package ws

// Message types
const (
	// Client -> Server
	TypeChatSend    = "chat.send"
	TypeChatEdit    = "chat.edit"
	TypeChatDelete  = "chat.delete"
	TypeChatTyping  = "chat.typing"
	TypePing        = "ping"

	// Server -> Client: typing indicator
	TypeChatUserTyping = "chat.user.typing"

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

	TypeChecklistItemCreated = "board.card.checklist.created"
	TypeChecklistItemUpdated = "board.card.checklist.updated"
	TypeChecklistItemDeleted = "board.card.checklist.deleted"
	TypeChecklistReordered   = "board.card.checklist.reordered"

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

	// Server -> Client: personal notifications
	TypeMentionNotification = "mention.notification"

	// Server -> Client: git card links
	TypeCardLinkCreated = "card.link.created"

	// Server -> Client: sprints
	TypeSprintCreated     = "sprint.created"
	TypeSprintUpdated     = "sprint.updated"
	TypeSprintDeleted     = "sprint.deleted"
	TypeSprintStarted     = "sprint.started"
	TypeSprintCompleted   = "sprint.completed"
	TypeSprintCardAdded   = "sprint.card.added"
	TypeSprintCardRemoved = "sprint.card.removed"

	// Call signaling (relayed directly between users via BroadcastToUser)
	TypeCallOffer       = "call.offer"
	TypeCallAnswer      = "call.answer"
	TypeCallICE         = "call.ice"
	TypeCallHangup      = "call.hangup"
	TypeCallReject      = "call.reject"
	TypeCallRing        = "call.ring"
	TypeCallUnavailable = "call.unavailable"
	TypeCallFailed      = "call.failed"
	TypeCallMute        = "call.mute"
	TypeCallScreenShare = "call.screen_share"
	TypeCallGroupInvite = "call.group_invite"

	// Server -> Client: tickets / helpdesk
	TypeTicketCreated  = "ticket.created"
	TypeTicketUpdated  = "ticket.updated"
	TypeTicketDeleted  = "ticket.deleted"
	TypeTicketMsgAdded = "ticket.message.added"

	// System
	TypePong  = "pong"
	TypeError = "error"
)

type Message struct {
	Type    string      `json:"type"`
	Payload any `json:"payload"`
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

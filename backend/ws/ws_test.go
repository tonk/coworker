package ws

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoryPubSub_IsLocal(t *testing.T) {
	m := &memoryPubSub{}
	assert.True(t, m.IsLocal())
}

func TestMemoryPubSub_Publish(t *testing.T) {
	m := &memoryPubSub{}
	assert.NoError(t, m.Publish("ch", []byte("data")))
}

func TestMemoryPubSub_Subscribe(t *testing.T) {
	m := &memoryPubSub{}
	called := false
	cancel := m.Subscribe("ch", func(b []byte) { called = true })
	assert.NotNil(t, cancel)
	cancel()
	assert.False(t, called)
}

func TestInitPubSub(t *testing.T) {
	m := &memoryPubSub{}
	InitPubSub(m)
	assert.Same(t, m, globalPubSub)
}

func TestStartPubSubListener_noopWhenLocal(t *testing.T) {
	m := &memoryPubSub{}
	InitPubSub(m)
	// Should not panic or hang
	StartPubSubListener()
}

func TestMessageTypeConstants(t *testing.T) {
	assert.Equal(t, "chat.send", TypeChatSend)
	assert.Equal(t, "chat.edit", TypeChatEdit)
	assert.Equal(t, "chat.delete", TypeChatDelete)
	assert.Equal(t, "chat.typing", TypeChatTyping)
	assert.Equal(t, "ping", TypePing)
	assert.Equal(t, "pong", TypePong)
	assert.Equal(t, "error", TypeError)

	assert.Equal(t, "chat.user.typing", TypeChatUserTyping)
	assert.Equal(t, "chat.message.created", TypeChatMessageCreated)
	assert.Equal(t, "chat.message.updated", TypeChatMessageUpdated)
	assert.Equal(t, "chat.message.deleted", TypeChatMessageDeleted)

	assert.Equal(t, "board.card.created", TypeBoardCardCreated)
	assert.Equal(t, "board.card.updated", TypeBoardCardUpdated)
	assert.Equal(t, "board.card.moved", TypeBoardCardMoved)
	assert.Equal(t, "board.card.deleted", TypeBoardCardDeleted)
	assert.Equal(t, "board.column.created", TypeBoardColumnCreated)
	assert.Equal(t, "board.columns.reordered", TypeBoardColumnsReordered)
	assert.Equal(t, "board.card.comment.created", TypeBoardCommentCreated)

	assert.Equal(t, "presence.joined", TypePresenceJoined)
	assert.Equal(t, "presence.left", TypePresenceLeft)
	assert.Equal(t, "presence.list", TypePresenceList)

	assert.Equal(t, "mention.notification", TypeMentionNotification)
	assert.Equal(t, "card.link.created", TypeCardLinkCreated)

	assert.Equal(t, "sprint.created", TypeSprintCreated)
	assert.Equal(t, "sprint.updated", TypeSprintUpdated)
	assert.Equal(t, "sprint.deleted", TypeSprintDeleted)

	assert.Equal(t, "call.offer", TypeCallOffer)
	assert.Equal(t, "call.answer", TypeCallAnswer)
	assert.Equal(t, "call.ice", TypeCallICE)
	assert.Equal(t, "call.hangup", TypeCallHangup)
	assert.Equal(t, "call.reject", TypeCallReject)
	assert.Equal(t, "call.ring", TypeCallRing)
	assert.Equal(t, "call.unavailable", TypeCallUnavailable)
	assert.Equal(t, "call.group_invite", TypeCallGroupInvite)

	assert.Equal(t, "chat.reaction.updated", TypeChatReactionUpdated)
	assert.Equal(t, "dm.reaction.updated", TypeDMReactionUpdated)
	assert.Equal(t, "dm.message.updated", TypeDMMessageUpdated)
	assert.Equal(t, "dm.message.deleted", TypeDMMessageDeleted)

	assert.Equal(t, "topic.created", TypeTopicCreated)
	assert.Equal(t, "topic.updated", TypeTopicUpdated)
	assert.Equal(t, "topic.deleted", TypeTopicDeleted)
	assert.Equal(t, "topic.reply.created", TypeTopicReplyCreated)
	assert.Equal(t, "topic.reply.updated", TypeTopicReplyUpdated)
	assert.Equal(t, "topic.reply.deleted", TypeTopicReplyDeleted)
}

func TestBroadcastChannel(t *testing.T) {
	assert.Equal(t, "warmdesk:broadcast", broadcastChannel)
}

func TestMessageStruct(t *testing.T) {
	m := Message{Type: "test", Payload: "data", ID: "123"}
	assert.Equal(t, "test", m.Type)
	assert.Equal(t, "data", m.Payload)
	assert.Equal(t, "123", m.ID)
}

func TestChatSendPayload(t *testing.T) {
	p := ChatSendPayload{Body: "hello"}
	assert.Equal(t, "hello", p.Body)
}

func TestChatDeletePayload(t *testing.T) {
	p := ChatDeletePayload{MessageID: 42}
	assert.Equal(t, uint(42), p.MessageID)
}

func TestErrorPayload(t *testing.T) {
	p := ErrorPayload{Code: "ERR", Message: "something went wrong", ID: "req1"}
	assert.Equal(t, "ERR", p.Code)
	assert.Equal(t, "something went wrong", p.Message)
	assert.Equal(t, "req1", p.ID)
}

func TestPresenceUser(t *testing.T) {
	u := PresenceUser{ID: 1, Username: "alice", DisplayName: "Alice", AvatarURL: "https://example.com/avatar"}
	assert.Equal(t, uint(1), u.ID)
	assert.Equal(t, "alice", u.Username)
	assert.Equal(t, "Alice", u.DisplayName)
	assert.Equal(t, "https://example.com/avatar", u.AvatarURL)
}

package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvelopeAddress_plain(t *testing.T) {
	assert.Equal(t, "alice@example.com", envelopeAddress("alice@example.com"))
}

func TestEnvelopeAddress_displayName(t *testing.T) {
	assert.Equal(t, "alice@example.com", envelopeAddress("Alice <alice@example.com>"))
	assert.Equal(t, "bob@test.org", envelopeAddress("\"Bob\" <bob@test.org>"))
}

func TestEnvelopeAddress_invalid(t *testing.T) {
	assert.Equal(t, "not-an-email", envelopeAddress("not-an-email"))
	assert.Equal(t, "", envelopeAddress(""))
}

func TestExtractMentions_none(t *testing.T) {
	assert.Nil(t, ExtractMentions("Hello world"))
	assert.Nil(t, ExtractMentions(""))
}

func TestExtractMentions_single(t *testing.T) {
	assert.Equal(t, []string{"alice"}, ExtractMentions("Hey @alice"))
}

func TestExtractMentions_multiple(t *testing.T) {
	assert.Equal(t, []string{"alice", "bob"}, ExtractMentions("Hey @alice and @bob"))
}

func TestExtractMentions_duplicates(t *testing.T) {
	assert.Equal(t, []string{"alice"}, ExtractMentions("@alice @alice"))
}

func TestExtractMentions_underscore(t *testing.T) {
	assert.Equal(t, []string{"test_user"}, ExtractMentions("Hello @test_user"))
}

func TestExtractMentions_nonUsernameRef(t *testing.T) {
	assert.Nil(t, ExtractMentions("not an @ mention"))
	assert.Equal(t, []string{"a", "c"}, ExtractMentions("@a b @c"))
}

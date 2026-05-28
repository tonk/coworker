package services

import (
	"strings"
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

func TestFoldBody_shortLinesUnchanged(t *testing.T) {
	body := "short line\nanother short line"
	assert.Equal(t, body, foldBody(body))
}

func TestFoldBody_longLineFolded(t *testing.T) {
	long := strings.Repeat("a", 500) + " " + strings.Repeat("b", 500)
	result := foldBody(long)
	// The result should contain a \r\n at the fold point
	assert.Contains(t, result, "\r\n")
	// Each line (after split by \n) should be <= maxSMTPLineLen
	for _, line := range strings.Split(result, "\n") {
		t.Logf("line length: %d: %s", len(line), line[:min(len(line), 50)])
		assert.LessOrEqual(t, len(line), maxSMTPLineLen)
	}
}

func TestFoldBody_multipleLongLines(t *testing.T) {
	line1 := strings.Repeat("x", 600) + " " + strings.Repeat("y", 600)
	line2 := strings.Repeat("m", 500) + " " + strings.Repeat("n", 500)
	body := line1 + "\n" + line2
	result := foldBody(body)
	lines := strings.Split(result, "\n")
	assert.Len(t, lines, 4) // both lines folded, plus the in-between
	for _, line := range lines {
		assert.LessOrEqual(t, len(line), maxSMTPLineLen, "each line must be within limit")
	}
}

func TestFoldBody_noSpacesFoldedAtExactPosition(t *testing.T) {
	long := strings.Repeat("a", maxSMTPLineLen+50)
	result := foldBody(long)
	// No spaces, so it folds at exact maxSMTPLineLen
	assert.Contains(t, result, "\r\n")
	prefix := result[:strings.Index(result, "\r\n")]
	assert.Len(t, prefix, maxSMTPLineLen)
}

func TestFoldBody_atBoundaryUnchanged(t *testing.T) {
	body := strings.Repeat("a", maxSMTPLineLen)
	assert.Equal(t, body, foldBody(body))
}

func TestFoldBody_justOverBoundaryFolded(t *testing.T) {
	body := strings.Repeat("a", maxSMTPLineLen+1)
	result := foldBody(body)
	assert.Contains(t, result, "\r\n")
}

func TestFoldBody_empty(t *testing.T) {
	assert.Equal(t, "", foldBody(""))
}

func TestFoldBody_preservesShortLines(t *testing.T) {
	body := "hello\nworld\nfoo\nbar"
	assert.Equal(t, body, foldBody(body))
}

func TestFoldHeader_shortUnchanged(t *testing.T) {
	subject := "Normal subject line"
	assert.Equal(t, subject, foldHeader(subject))
}

func TestFoldHeader_longFolded(t *testing.T) {
	subject := strings.Repeat("x", 600) + " " + strings.Repeat("y", 600)
	result := foldHeader(subject)
	for _, line := range strings.Split(result, "\n") {
		assert.LessOrEqual(t, len(line), maxSMTPLineLen)
	}
}

func TestFoldHeader_empty(t *testing.T) {
	assert.Equal(t, "", foldHeader(""))
}

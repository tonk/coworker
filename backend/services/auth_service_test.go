package services

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestAuthService(t *testing.T) *AuthService {
	t.Helper()
	return NewAuthService("test-secret-key-that-is-long-enough-32")
}

func TestHashAndCheckPassword(t *testing.T) {
	s := newTestAuthService(t)

	hash, err := s.HashPassword("correct-password")
	require.NoError(t, err)
	assert.NotEmpty(t, hash)

	assert.True(t, s.CheckPassword(hash, "correct-password"))
	assert.False(t, s.CheckPassword(hash, "wrong-password"))
	assert.False(t, s.CheckPassword("invalid-hash", "password"))
}

func TestIssueAccessToken(t *testing.T) {
	s := newTestAuthService(t)
	token, err := s.IssueAccessToken(1, "alice", "admin")
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := s.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, uint(1), claims.UserID)
	assert.Equal(t, "alice", claims.Username)
	assert.Equal(t, "admin", claims.GlobalRole)
	assert.Empty(t, claims.Purpose)
	assert.False(t, claims.MFAPending)
}

func TestIssueRefreshToken(t *testing.T) {
	s := newTestAuthService(t)
	token, err := s.IssueRefreshToken(2, "bob", "member")
	require.NoError(t, err)

	claims, err := s.ValidateRefreshToken(token)
	require.NoError(t, err)
	assert.Equal(t, uint(2), claims.UserID)
	assert.Equal(t, "bob", claims.Username)
	assert.Equal(t, "member", claims.GlobalRole)
	assert.Equal(t, "refresh", claims.Purpose)

	_, err = s.ValidateToken(token)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestIssueMFAToken(t *testing.T) {
	s := newTestAuthService(t)
	token, err := s.IssueMFAToken(3, "charlie")
	require.NoError(t, err)

	claims, err := s.ValidateMFAToken(token)
	require.NoError(t, err)
	assert.Equal(t, uint(3), claims.UserID)
	assert.True(t, claims.MFAPending)
	assert.Equal(t, "mfa", claims.Purpose)

	_, err = s.ValidateToken(token)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestValidateRefreshToken_rejectsAccessToken(t *testing.T) {
	s := newTestAuthService(t)
	accessToken, err := s.IssueAccessToken(1, "alice", "admin")
	require.NoError(t, err)

	_, err = s.ValidateRefreshToken(accessToken)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestValidateMFAToken_rejectsNonMFAToken(t *testing.T) {
	s := newTestAuthService(t)
	accessToken, err := s.IssueAccessToken(1, "alice", "admin")
	require.NoError(t, err)

	_, err = s.ValidateMFAToken(accessToken)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestIssueWSTicket(t *testing.T) {
	s := newTestAuthService(t)
	token, err := s.IssueWSTicket(4, "dave", "viewer")
	require.NoError(t, err)

	claims, err := s.ValidateWSTicket(token)
	require.NoError(t, err)
	assert.Equal(t, uint(4), claims.UserID)
	assert.Equal(t, "ws", claims.Purpose)

	_, err = s.ValidateToken(token)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestIssueMediaTicket(t *testing.T) {
	s := newTestAuthService(t)
	token, err := s.IssueMediaTicket(5, "eve", "member")
	require.NoError(t, err)

	claims, err := s.ValidateMediaTicket(token)
	require.NoError(t, err)
	assert.Equal(t, uint(5), claims.UserID)
	assert.Equal(t, "media", claims.Purpose)

	_, err = s.ValidateToken(token)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestIssuePasskeyChallenge(t *testing.T) {
	s := newTestAuthService(t)
	token, err := s.IssuePasskeyChallenge(6, "passkey_reg", `{"challenge":"abc"}`)
	require.NoError(t, err)

	claims, err := s.ValidatePasskeyChallenge(token, "passkey_reg")
	require.NoError(t, err)
	assert.Equal(t, uint(6), claims.UserID)
	assert.Equal(t, "passkey_reg", claims.Purpose)
	assert.Equal(t, `{"challenge":"abc"}`, claims.PasskeySession)
}

func TestValidatePasskeyChallenge_wrongPurpose(t *testing.T) {
	s := newTestAuthService(t)
	token, err := s.IssuePasskeyChallenge(0, "passkey_auth", `{}`)
	require.NoError(t, err)

	_, err = s.ValidatePasskeyChallenge(token, "passkey_reg")
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestValidateToken_expired(t *testing.T) {
	s := &AuthService{secret: []byte("test-secret-key-that-is-long-enough-32")}
	claims := Claims{
		UserID:   1,
		Username: "alice",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(s.secret)
	require.NoError(t, err)

	_, err = s.ValidateToken(tokenStr)
	assert.ErrorIs(t, err, ErrExpiredToken)
}

func TestValidateToken_invalidSignature(t *testing.T) {
	s := newTestAuthService(t)
	other := NewAuthService("different-secret-not-long-enough-32!")

	tokenStr, err := other.IssueAccessToken(1, "alice", "admin")
	require.NoError(t, err)

	_, err = s.ValidateToken(tokenStr)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestGenerateAndVerifyTOTP(t *testing.T) {
	s := newTestAuthService(t)
	secret, uri, err := s.GenerateTOTPSecret("alice", "WarmDesk")
	require.NoError(t, err)
	assert.NotEmpty(t, secret)
	assert.Contains(t, uri, "otpauth://totp/")
	assert.Contains(t, uri, "issuer=WarmDesk")
	assert.Contains(t, uri, "alice")

	code, err := totp.GenerateCode(secret, time.Now())
	require.NoError(t, err)
	assert.Len(t, code, 6)
	assert.True(t, s.VerifyTOTP(secret, code))
	assert.False(t, s.VerifyTOTP(secret, "000000"))
}

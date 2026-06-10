package services

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

var (
	ErrInvalidToken  = errors.New("invalid token")
	ErrExpiredToken  = errors.New("token expired")
	ErrInvalidCreds  = errors.New("invalid credentials")
)

type Claims struct {
	UserID     uint   `json:"user_id"`
	Username   string `json:"username"`
	GlobalRole string `json:"global_role"`
	MFAPending bool   `json:"mfa_pending,omitempty"`
	// Purpose restricts a token to a specific subsystem ("ws" | "media" | "passkey_reg" | "passkey_auth").
	// Empty = normal access token.
	Purpose        string `json:"purpose,omitempty"`
	PasskeySession string `json:"passkey_session,omitempty"`
	jwt.RegisteredClaims
}

type AuthService struct {
	secret []byte
}

func NewAuthService(secret string) *AuthService {
	return &AuthService{secret: []byte(secret)}
}

func (s *AuthService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	return string(hash), err
}

func (s *AuthService) CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func (s *AuthService) IssueAccessToken(userID uint, username, role string) (string, error) {
	claims := Claims{
		UserID:     userID,
		Username:   username,
		GlobalRole: role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *AuthService) IssueRefreshToken(userID uint, username, role string) (string, error) {
	claims := Claims{
		UserID:     userID,
		Username:   username,
		GlobalRole: role,
		Purpose:    "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// IssueMFAToken returns a short-lived (5 min) JWT used only for the TOTP verification step.
func (s *AuthService) IssueMFAToken(userID uint, username string) (string, error) {
	claims := Claims{
		UserID:     userID,
		Username:   username,
		MFAPending: true,
		Purpose:    "mfa",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// GenerateTOTPSecret creates a new TOTP secret. Returns the base32 secret and the
// otpauth:// URI suitable for encoding as a QR code.
func (s *AuthService) GenerateTOTPSecret(username, issuer string) (secret, uri string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: username,
	})
	if err != nil {
		return "", "", err
	}
	return key.Secret(), key.URL(), nil
}

// VerifyTOTP validates a 6-digit TOTP code against the given base32 secret.
func (s *AuthService) VerifyTOTP(secret, code string) bool {
	return totp.Validate(code, secret)
}

// IssueWSTicket returns a 30-second JWT accepted only by the WebSocket upgrade endpoints.
func (s *AuthService) IssueWSTicket(userID uint, username, role string) (string, error) {
	claims := Claims{
		UserID:     userID,
		Username:   username,
		GlobalRole: role,
		Purpose:    "ws",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// IssueMediaTicket returns a 5-minute JWT accepted only by the attachment download endpoint.
func (s *AuthService) IssueMediaTicket(userID uint, username, role string) (string, error) {
	claims := Claims{
		UserID:     userID,
		Username:   username,
		GlobalRole: role,
		Purpose:    "media",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *AuthService) parseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// ParseAccessClaimsLenient parses an access token without validating expiry.
// Used by the Logout handler to identify the user even when the token has just expired.
func (s *AuthService) ParseAccessClaimsLenient(tokenStr string) (*Claims, error) {
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, err := parser.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || claims.Purpose != "" {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// ValidateToken validates a normal access token; rejects purpose-limited tickets.
func (s *AuthService) ValidateToken(tokenStr string) (*Claims, error) {
	claims, err := s.parseToken(tokenStr)
	if err != nil {
		return nil, err
	}
	if claims.Purpose != "" {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// ValidateRefreshToken validates a refresh token; rejects access tokens or other purpose-limited tokens.
func (s *AuthService) ValidateRefreshToken(tokenStr string) (*Claims, error) {
	claims, err := s.parseToken(tokenStr)
	if err != nil {
		return nil, err
	}
	if claims.Purpose != "refresh" {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// ValidateMFAToken validates a short-lived MFA challenge token; rejects any other token type.
func (s *AuthService) ValidateMFAToken(tokenStr string) (*Claims, error) {
	claims, err := s.parseToken(tokenStr)
	if err != nil {
		return nil, err
	}
	if claims.Purpose != "mfa" || !claims.MFAPending {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// ValidateWSTicket validates a short-lived WebSocket ticket.
func (s *AuthService) ValidateWSTicket(tokenStr string) (*Claims, error) {
	claims, err := s.parseToken(tokenStr)
	if err != nil {
		return nil, err
	}
	if claims.Purpose != "ws" {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// IssuePasskeyChallenge issues a 5-minute JWT carrying the serialized WebAuthn SessionData.
// purpose must be "passkey_reg" (registration) or "passkey_auth" (authentication).
// userID may be 0 for passkey_auth tokens when the user is not yet known.
func (s *AuthService) IssuePasskeyChallenge(userID uint, purpose, sessionJSON string) (string, error) {
	claims := Claims{
		UserID:         userID,
		Purpose:        purpose,
		PasskeySession: sessionJSON,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// ValidatePasskeyChallenge validates a passkey challenge token and checks the purpose.
func (s *AuthService) ValidatePasskeyChallenge(tokenStr, purpose string) (*Claims, error) {
	claims, err := s.parseToken(tokenStr)
	if err != nil {
		return nil, err
	}
	if claims.Purpose != purpose {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// ValidateMediaTicket validates a short-lived media download ticket.
func (s *AuthService) ValidateMediaTicket(tokenStr string) (*Claims, error) {
	claims, err := s.parseToken(tokenStr)
	if err != nil {
		return nil, err
	}
	if claims.Purpose != "media" {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

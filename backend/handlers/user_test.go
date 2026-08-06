package handlers

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// bindTestReq mirrors the validated fields of AdminCreateUser's request struct.
type bindTestReq struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required,min=3,max=50"`
}

// bind runs the given JSON body through gin's actual ShouldBindJSON path (the
// same one AdminCreateUser uses), so the "binding" struct tag is honored
// exactly as it is in production.
func bind(t *testing.T, body string) error {
	t.Helper()
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	var req bindTestReq
	return c.ShouldBindJSON(&req)
}

func TestBindErrorMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("invalid email", func(t *testing.T) {
		err := bind(t, `{"email":"not-an-email","username":"alice"}`)
		require.Error(t, err)
		assert.Equal(t, "invalid email address", bindErrorMessage(err))
	})

	t.Run("required field", func(t *testing.T) {
		err := bind(t, `{"email":"alice@example.com","username":""}`)
		require.Error(t, err)
		assert.Equal(t, "username is required", bindErrorMessage(err))
	})

	t.Run("too short", func(t *testing.T) {
		err := bind(t, `{"email":"alice@example.com","username":"ab"}`)
		require.Error(t, err)
		assert.Equal(t, "username must be at least 3 characters", bindErrorMessage(err))
	})

	t.Run("non-validation error falls back to generic message", func(t *testing.T) {
		assert.Equal(t, "invalid request", bindErrorMessage(errors.New("unexpected EOF")))
	})
}

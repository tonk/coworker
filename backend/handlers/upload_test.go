package handlers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonk/warmdesk/config"
)

func TestDeleteOldUpload(t *testing.T) {
	dir := t.TempDir()
	prevCfg := attachmentCfg
	attachmentCfg = &config.Config{UploadDir: dir}
	defer func() { attachmentCfg = prevCfg }()

	writeFile := func(name string) string {
		path := filepath.Join(dir, name)
		require.NoError(t, os.WriteFile(path, []byte("x"), 0644))
		return path
	}

	t.Run("removes the old file when replaced by a different one", func(t *testing.T) {
		oldPath := writeFile("old1.png")
		deleteOldUpload("/uploads/old1.png", "/uploads/new1.png")
		_, err := os.Stat(oldPath)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("removes the old file when cleared", func(t *testing.T) {
		oldPath := writeFile("old2.png")
		deleteOldUpload("/uploads/old2.png", "")
		_, err := os.Stat(oldPath)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("no-op when the URL is unchanged", func(t *testing.T) {
		oldPath := writeFile("same.png")
		deleteOldUpload("/uploads/same.png", "/uploads/same.png")
		_, err := os.Stat(oldPath)
		assert.NoError(t, err)
	})

	t.Run("no-op when there was no previous URL", func(t *testing.T) {
		// Must not panic or attempt to delete anything for an empty old value.
		deleteOldUpload("", "/uploads/new.png")
	})

	t.Run("no-op for an external (non-local) old URL", func(t *testing.T) {
		deleteOldUpload("https://example.com/old.png", "/uploads/new.png")
	})

	t.Run("no-op when the old file no longer exists on disk", func(t *testing.T) {
		deleteOldUpload("/uploads/does-not-exist.png", "/uploads/new.png")
	})
}

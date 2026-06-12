package handlers

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/config"
)

var docsCfg *config.Config
var docsWebFS fs.FS

// InitDocs stores config and the web filesystem for guide PDF resolution.
func InitDocs(cfg *config.Config, webFS fs.FS) {
	docsCfg = cfg
	docsWebFS = webFS
}

// DownloadUserGuide serves the bundled user guide PDF to authenticated users.
func DownloadUserGuide(c *gin.Context) {
	serveGuidePDF(c, "user-guide.pdf")
}

// DownloadAdminGuide serves the bundled admin guide PDF to admins.
func DownloadAdminGuide(c *gin.Context) {
	serveGuidePDF(c, "admin-guide.pdf")
}

func serveGuidePDF(c *gin.Context, filename string) {
	data, err := readGuidePDF(filename)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "guide not available"})
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, "application/pdf", data)
}

func readGuidePDF(filename string) ([]byte, error) {
	rel := filepath.Join("docs", filename)

	if docsWebFS != nil {
		if data, err := fs.ReadFile(docsWebFS, rel); err == nil {
			return data, nil
		}
	}

	if docsCfg != nil && docsCfg.WebDir != "" {
		if data, err := os.ReadFile(filepath.Join(docsCfg.WebDir, rel)); err == nil {
			return data, nil
		}
	}

	for _, base := range []string{".", ".."} {
		if data, err := os.ReadFile(filepath.Join(base, rel)); err == nil {
			return data, nil
		}
	}

	return nil, os.ErrNotExist
}

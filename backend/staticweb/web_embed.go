//go:build embed

package staticweb

import (
	"embed"
	"io/fs"
)

//go:embed files
var embedFS embed.FS

// FS is the embedded frontend build, rooted at the files/ directory.
// Populated by the Makefile's embed-web step before compilation.
var FS, _ = fs.Sub(embedFS, "files")

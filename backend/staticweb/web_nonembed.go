//go:build !embed

package staticweb

import "io/fs"

// FS is nil in non-embed builds; set WEB_DIR env/config to serve the frontend from the filesystem.
var FS fs.FS

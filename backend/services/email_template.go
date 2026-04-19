package services

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// appInfoReader is registered from main.go so services can read live branding/version
// without importing the handlers package (which would create a cycle).
var appInfoReader func() (version, companyName, logoURL, instanceURL string)

// SetAppInfoReader registers the function used to read live app info for emails.
func SetAppInfoReader(fn func() (string, string, string, string)) {
	appInfoReader = fn
}

func getAppInfo() (version, companyName, logoURL, instanceURL string) {
	if appInfoReader != nil {
		return appInfoReader()
	}
	return "dev", "WarmDesk", "", ""
}

// GetAppInfo is the exported variant used by handler packages that need branding
// context (e.g. to build an email subject line) without duplicating the lookup.
func GetAppInfo() (version, companyName, logoURL, instanceURL string) {
	return getAppInfo()
}

// warmDeskLogoSVG is a minimal, Inkscape-metadata-free version of the WarmDesk icon.
const warmDeskLogoSVG = `<svg width="160" height="156.5" viewBox="0 0 160 156.5" xmlns="http://www.w3.org/2000/svg">` +
	`<rect fill="#1D9E75" x="24" y="104.5" width="11" height="52" rx="3"/>` +
	`<rect fill="#1D9E75" x="125" y="104.5" width="11" height="52" rx="3"/>` +
	`<rect fill="#5DCAA5" x="10" y="88.5" width="140" height="20" rx="5"/>` +
	`<rect fill="#9FE1CB" x="0" y="76.5" width="160" height="16" rx="4"/>` +
	`<rect fill="#1D9E75" x="54" y="40.5" width="52" height="40" rx="5"/>` +
	`<rect fill="#0F6E56" x="48" y="34.5" width="64" height="10" rx="4"/>` +
	`<path d="m106,50.5q20,0 20,10 0,10-20,10" fill="none" stroke="#9FE1CB" stroke-width="5" stroke-linecap="round"/>` +
	`<path d="m66,30.5q-4,-12 0,-24" fill="none" stroke="#9FE1CB" stroke-width="3" stroke-linecap="round"/>` +
	`<path d="m80,27.5q-4,-13 0,-26" fill="none" stroke="#9FE1CB" stroke-width="3" stroke-linecap="round"/>` +
	`<path d="m94,30.5q-4,-12 0,-24" fill="none" stroke="#9FE1CB" stroke-width="3" stroke-linecap="round"/>` +
	`</svg>`

var warmDeskLogoDataURI = "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(warmDeskLogoSVG))

// WrapHTML wraps bodyHTML in the standard WarmDesk email shell:
//   - blue header with company name/logo and a subtitle line
//   - white content area (bodyHTML)
//   - footer with the WarmDesk icon, version, and instance URL
func WrapHTML(headerSubtitle, bodyHTML string) string {
	ver, companyName, logoURL, instanceURL := getAppInfo()

	// Strip a leading "v" from the version tag (e.g. "v0.7.7" → "0.7.7").
	ver = strings.TrimPrefix(ver, "v")

	displayName := companyName
	if displayName == "" {
		displayName = "WarmDesk"
	}

	var companyLogoHTML string
	if logoURL != "" {
		companyLogoHTML = fmt.Sprintf(
			`<img src="%s" alt="%s" style="max-height:48px;max-width:160px;margin-bottom:10px;display:block;margin-left:auto;margin-right:auto">`,
			logoURL, displayName)
	}

	footerLabel := "WarmDesk"
	if companyName != "" && companyName != "WarmDesk" {
		footerLabel = companyName + " &nbsp;·&nbsp; WarmDesk"
	}

	var urlHTML string
	if instanceURL != "" {
		urlHTML = fmt.Sprintf(` &nbsp;·&nbsp; <a href="%s" style="color:#bbb;text-decoration:none">%s</a>`, instanceURL, instanceURL)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<meta name="color-scheme" content="light dark">
<meta name="supported-color-schemes" content="light dark">
<style>
@media (prefers-color-scheme: dark) {
  /* Button: use a lighter blue so white text stays readable */
  .wd-btn { background-color: #4d94e8 !important; border-color: #4d94e8 !important; color: #ffffff !important; }
  /* Card: soften the white to avoid harsh contrast */
  .wd-card { background-color: #1e1e2e !important; }
  .wd-body td { color: #d0d0d0 !important; }
  .wd-footer { background-color: #2a2a3a !important; border-top-color: #444 !important; }
}
</style>
</head>
<body style="margin:0;padding:0;background:#f0f2f5;font-family:Arial,Helvetica,sans-serif">
<table width="100%%" cellpadding="0" cellspacing="0" style="background:#f0f2f5;padding:32px 16px">
<tr><td align="center">
<table width="560" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:10px;overflow:hidden;box-shadow:0 2px 12px rgba(0,0,0,.10);max-width:560px">

  <!-- Header -->
  <tr>
    <td style="background:#1a5fb4;padding:28px 32px;text-align:center">
      %s
      <div style="color:#ffffff;font-size:20px;font-weight:bold;letter-spacing:.3px">%s</div>
      <div style="color:#a8c8f8;font-size:13px;margin-top:4px">%s</div>
    </td>
  </tr>

  <!-- Content -->
  %s

  <!-- Footer -->
  <tr>
    <td style="background:#f5f5f5;padding:14px 32px;text-align:center;border-top:1px solid #e8e8e8">
      <img src="%s" width="18" height="18" alt="WarmDesk" style="vertical-align:middle;margin-right:5px">
      <span style="font-size:12px;color:#999">%s &nbsp;·&nbsp; v%s%s</span>
    </td>
  </tr>

</table>
</td></tr>
</table>
</body>
</html>`,
		companyLogoHTML, displayName, headerSubtitle,
		bodyHTML,
		warmDeskLogoDataURI, footerLabel, ver, urlHTML,
	)
}

// WrapText adds a consistent plain-text header/footer around bodyText.
func WrapText(headerSubtitle, bodyText string) string {
	ver, companyName, _, instanceURL := getAppInfo()
	ver = strings.TrimPrefix(ver, "v")
	if companyName == "" {
		companyName = "WarmDesk"
	}
	sep := strings.Repeat("-", 48)
	var b strings.Builder
	fmt.Fprintf(&b, "%s — %s\n%s\n\n", companyName, headerSubtitle, sep)
	b.WriteString(bodyText)
	b.WriteString("\n" + sep + "\n")
	fmt.Fprintf(&b, "WarmDesk v%s", ver)
	if instanceURL != "" {
		fmt.Fprintf(&b, "  |  %s", instanceURL)
	}
	b.WriteString("\n")
	return b.String()
}

package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/testutil"
)

// A regression guard for the bug where saving any setting on the Branding &
// Billing admin tab failed with a generic "invalid request" as soon as the
// optional company_payment_terms/default_vat_rate fields were blank: they
// were typed as json.Number, which rejects "" outright and fails the
// unmarshal of the WHOLE request struct — so even fields unrelated to
// billing (like login_branding_enabled) never got saved.

func TestFlexNumberUnmarshalJSON(t *testing.T) {
	cases := []struct {
		name    string
		json    string
		want    flexNumber
		wantErr bool
	}{
		{"empty string", `""`, "", false},
		{"null", `null`, "", false},
		{"numeric string", `"30"`, "30", false},
		{"decimal string", `"21.5"`, "21.5", false},
		{"bare int", `30`, "30", false},
		{"bare float", `21.5`, "21.5", false},
		{"garbage string", `"abc"`, "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var n flexNumber
			err := json.Unmarshal([]byte(tc.json), &n)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, n)
		})
	}
}

func TestAdminUpdateSystemSettings_BlankBillingFieldsDontBlockOtherFields(t *testing.T) {
	db, cleanup := testutil.SetupTestDB()
	defer cleanup()
	database.DB = db

	// Mirrors exactly what saveBrandingSettings() sends: every billing field
	// included, none of them ever filled in on a fresh instance.
	payload := map[string]interface{}{
		"company_name":           "",
		"company_logo":           "",
		"company_logo_dark":      "",
		"login_branding_enabled": true,
		"company_address":        "",
		"company_city":           "",
		"company_postal_code":    "",
		"company_country":        "",
		"company_vat_number":     "",
		"company_coc_number":     "",
		"company_iban":           "",
		"company_bic":            "",
		"company_payment_terms":  "",
		"invoice_number_prefix":  "",
		"default_vat_rate":       "",
		"invoice_vat_exempt":     false,
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/admin/system", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c, rec := newAuthTestContext(req)

	AdminUpdateSystemSettings(c)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "true", loadAllSettings()[settingLoginBrandingEnabled])
}

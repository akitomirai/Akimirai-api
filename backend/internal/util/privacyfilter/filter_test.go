package privacyfilter

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRedactJSONBodyRedactsSelectedSensitiveValues(t *testing.T) {
	body := []byte(`{
		"messages": [
			{"role": "user", "content": "mail me at user@example.com from 192.168.1.1, phone 13800138000, id 110105199001011234, card 4532015112830366"}
		],
		"metadata": {
			"api_key": "sk-abcdefghijklmnopqrstuvwxyz123456",
			"authorization": "Bearer eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.sflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
		}
	}`)

	out, changed, err := RedactJSONBody(body, Config{
		Enabled: true,
		Types:   []string{TypeEmail, TypeIPAddress, TypePhone, TypeIDCard, TypeBankCard, TypeAPIKey, TypeToken},
	})
	if err != nil {
		t.Fatalf("RedactJSONBody returned error: %v", err)
	}
	if !changed {
		t.Fatal("expected JSON body to change")
	}

	redacted := string(out)
	for _, leaked := range []string{
		"user@example.com",
		"192.168.1.1",
		"13800138000",
		"110105199001011234",
		"4532015112830366",
		"sk-abcdefghijklmnopqrstuvwxyz123456",
		"eyJhbGciOiJIUzI1NiJ9",
	} {
		if strings.Contains(redacted, leaked) {
			t.Fatalf("expected %q to be redacted from %s", leaked, redacted)
		}
	}
	for _, marker := range []string{
		replacementEmail,
		replacementIP,
		replacementPhone,
		replacementIDCard,
		replacementBankCard,
		replacementAPIKey,
		replacementToken,
	} {
		if !strings.Contains(redacted, marker) {
			t.Fatalf("expected marker %q in %s", marker, redacted)
		}
	}
}

func TestRedactJSONBodyKeepsDisabledOrInvalidJSONUnchanged(t *testing.T) {
	body := []byte(`{"prompt":"user@example.com"}`)
	out, changed, err := RedactJSONBody(body, Config{Enabled: false, Types: []string{TypeEmail}})
	if err != nil {
		t.Fatalf("RedactJSONBody returned error: %v", err)
	}
	if changed || string(out) != string(body) {
		t.Fatalf("expected disabled filter to keep body unchanged, changed=%v out=%s", changed, out)
	}

	invalid := []byte(`{"prompt":`)
	out, changed, err = RedactJSONBody(invalid, Config{Enabled: true, Types: []string{TypeEmail}})
	if err != nil {
		t.Fatalf("RedactJSONBody returned error: %v", err)
	}
	if changed || string(out) != string(invalid) {
		t.Fatalf("expected invalid JSON to pass through unchanged, changed=%v out=%s", changed, out)
	}
}

func TestRedactJSONBodyPreservesNonStringValues(t *testing.T) {
	body := []byte(`{"n":123,"ok":true,"nested":["user@example.com",null]}`)
	out, changed, err := RedactJSONBody(body, Config{Enabled: true, Types: []string{TypeEmail}})
	if err != nil {
		t.Fatalf("RedactJSONBody returned error: %v", err)
	}
	if !changed {
		t.Fatal("expected body to change")
	}
	var parsed map[string]any
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatalf("redacted body should remain valid JSON: %v", err)
	}
	if parsed["n"].(float64) != 123 || parsed["ok"].(bool) != true {
		t.Fatalf("expected non-string fields to be preserved: %#v", parsed)
	}
}

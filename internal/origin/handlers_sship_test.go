package origin

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestSSHAndIP_Handlers(t *testing.T) {
	_, db, h := testServer(t)
	seedNS(t, db, "acme")

	// 1. IP Allowlist initially disabled
	res := do(h, "GET", "/v1/origin/namespaces/acme/inbound-ip-allowlist", "")
	if res.Code != http.StatusOK {
		t.Fatalf("get allowlist expected 200, got %d", res.Code)
	}
	var alResp struct {
		Enabled bool             `json:"enabled"`
		Entries []map[string]any `json:"entries"`
		ETag    string           `json:"etag"`
	}
	_ = json.Unmarshal(res.Body.Bytes(), &alResp)
	if alResp.Enabled || len(alResp.Entries) != 0 {
		t.Fatalf("expected disabled and empty allowlist, got %+v", alResp)
	}

	// 2. Enable IP Allowlist
	res = do(h, "PATCH", "/v1/origin/namespaces/acme/inbound-ip-allowlist", `{"enabled":true}`)
	if res.Code != http.StatusOK {
		t.Fatalf("enable allowlist expected 200, got %d", res.Code)
	}
	_ = json.Unmarshal(res.Body.Bytes(), &alResp)
	if !alResp.Enabled {
		t.Fatalf("expected enabled true")
	}

	// 3. Add entry
	res = do(h, "POST", "/v1/origin/namespaces/acme/inbound-ip-allowlist/entries", `{"cidr":"192.168.1.0/24","description":"office"}`)
	if res.Code != http.StatusOK {
		t.Fatalf("add entry expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var entryResp map[string]any
	_ = json.Unmarshal(res.Body.Bytes(), &entryResp)
	entryID, _ := entryResp["id"].(string)
	if entryID == "" || entryResp["cidr"] != "192.168.1.0/24" {
		t.Fatalf("unexpected entry response: %+v", entryResp)
	}

	// 4. Get entry
	res = do(h, "GET", "/v1/origin/namespaces/acme/inbound-ip-allowlist/entries/"+entryID, "")
	if res.Code != http.StatusOK {
		t.Fatalf("get entry expected 200, got %d", res.Code)
	}

	// 5. Update entry
	res = do(h, "PATCH", "/v1/origin/namespaces/acme/inbound-ip-allowlist/entries/"+entryID, `{"description":"headquarters"}`)
	if res.Code != http.StatusOK {
		t.Fatalf("update entry expected 200, got %d", res.Code)
	}
	_ = json.Unmarshal(res.Body.Bytes(), &entryResp)
	if entryResp["description"] != "headquarters" {
		t.Fatalf("expected updated description, got %v", entryResp["description"])
	}

	// 6. Replace entries
	replaceBody := `{"entries":[{"cidr":"10.0.0.0/16","description":"vpc"}],"allowEmpty":false}`
	res = do(h, "POST", "/v1/origin/namespaces/acme/inbound-ip-allowlist/entries:replace", replaceBody)
	if res.Code != http.StatusOK {
		t.Fatalf("replace entries expected 200, got %d: %s", res.Code, res.Body.String())
	}

	// 7. SSH CAs
	// Initially empty
	res = do(h, "GET", "/v1/origin/namespaces/acme/ssh-certificate-authorities", "")
	if res.Code != http.StatusOK {
		t.Fatalf("list ssh cas expected 200, got %d", res.Code)
	}
	var caListResp struct {
		CertificateAuthorities []map[string]any `json:"certificateAuthorities"`
	}
	_ = json.Unmarshal(res.Body.Bytes(), &caListResp)
	if len(caListResp.CertificateAuthorities) != 0 {
		t.Fatalf("expected empty cas, got %d", len(caListResp.CertificateAuthorities))
	}

	// Add SSH CA
	addCaBody := `{"name":"Corp CA","publicKey":"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIInJ8YxG668jK9+k5kU8vGfK9w2Z3mR8aX7n0p2y9z8Q corp-ca"}`
	res = do(h, "POST", "/v1/origin/namespaces/acme/ssh-certificate-authorities", addCaBody)
	if res.Code != http.StatusOK {
		t.Fatalf("add ssh ca expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var caResp map[string]any
	_ = json.Unmarshal(res.Body.Bytes(), &caResp)
	caID, _ := caResp["id"].(string)
	if caID == "" || caResp["keyType"] != "ssh-ed25519" || caResp["fingerprint"] == "" {
		t.Fatalf("unexpected ca response: %+v", caResp)
	}

	// Set requirement
	res = do(h, "POST", "/v1/origin/namespaces/acme/ssh-certificate-authorities:setRequirement", `{"requireCertificates":true}`)
	if res.Code != http.StatusOK {
		t.Fatalf("set requirement expected 200, got %d", res.Code)
	}

	// Delete SSH CA
	res = do(h, "DELETE", "/v1/origin/namespaces/acme/ssh-certificate-authorities/"+caID, "")
	if res.Code != http.StatusNoContent {
		t.Fatalf("delete ca expected 204, got %d", res.Code)
	}
}

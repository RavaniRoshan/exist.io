package origin

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestMirror_Handlers(t *testing.T) {
	srv, db, h := testServer(t)
	seedRepoWithBranches(t, srv, db, "acme", "rocket")

	// 1. TransitionRepoMirror
	transBody := `{"transition":"initial_to_inbound"}`
	res := do(h, "POST", "/v1/origin/repos/acme/rocket/mirror:transition", transBody)
	if res.Code != http.StatusOK {
		t.Fatalf("transition mirror expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var transResp struct {
		Repository map[string]any `json:"repository"`
		Job        map[string]any `json:"job"`
	}
	_ = json.Unmarshal(res.Body.Bytes(), &transResp)
	jobID, _ := transResp.Job["id"].(string)
	if jobID == "" || transResp.Job["status"] != "succeeded" {
		t.Fatalf("unexpected transition job: %+v", transResp.Job)
	}

	// 2. GetMirrorTransitionJob
	res = do(h, "GET", "/v1/origin/repos/acme/rocket/mirror/transition-jobs/"+jobID, "")
	if res.Code != http.StatusOK {
		t.Fatalf("get job expected 200, got %d", res.Code)
	}

	// 3. GetActiveMirrorTransitionJob
	res = do(h, "GET", "/v1/origin/repos/acme/rocket/mirror/transition-jobs:active", "")
	if res.Code != http.StatusOK {
		t.Fatalf("get active job expected 200, got %d", res.Code)
	}
	var activeResp struct {
		ActiveJob map[string]any `json:"activeJob"`
		LastJob   map[string]any `json:"lastJob"`
	}
	_ = json.Unmarshal(res.Body.Bytes(), &activeResp)
	if activeResp.LastJob == nil {
		t.Fatalf("expected lastJob to be present")
	}

	// 4. SyncMirror
	syncBody := `{"ref":"refs/heads/main","wait":true}`
	res = do(h, "POST", "/v1/origin/repos/acme/rocket:syncMirror", syncBody)
	if res.Code != http.StatusOK {
		t.Fatalf("sync mirror expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var syncResp struct {
		Synced bool `json:"synced"`
	}
	_ = json.Unmarshal(res.Body.Bytes(), &syncResp)
	if !syncResp.Synced {
		t.Fatalf("expected synced true")
	}

	// 5. DetachRepoMirror
	res = do(h, "DELETE", "/v1/origin/repos/acme/rocket/mirror", "")
	if res.Code != http.StatusNoContent {
		t.Fatalf("detach mirror expected 204, got %d", res.Code)
	}
}

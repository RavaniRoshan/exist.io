package app_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"exist.io/internal/app"
	"exist.io/internal/config"
	"exist.io/internal/store"
)

func TestNotificationsAPI(t *testing.T) {
	cfg := config.Config{
		DataDir: ":memory:",
	}
	application, err := app.Open(cfg)
	if err != nil {
		t.Fatalf("Open app: %v", err)
	}
	defer application.Close()

	ctx := context.Background()

	// Create user
	u := store.User{
		ID:           "usr-test-notif",
		Login:        "shiva",
		Email:        "shiva@exist.io",
		PasswordHash: "dummyhash",
	}
	_ = application.Store.CreateUserWithPassword(ctx, u)

	// Create in-app notification via application helper
	if err := application.CreateInAppNotification(ctx, u.ID, "pr.opened", "PR #1 Opened", "Feature branch opened", "/shiva/repo/pull/1"); err != nil {
		t.Fatalf("CreateInAppNotification: %v", err)
	}

	handler := application.Handler()

	// 1. GET unread count
	req := httptest.NewRequest("GET", "/api/v1/notifications/unread-count", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var countResp struct {
		Ok    bool `json:"ok"`
		Count int  `json:"count"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&countResp)
	if !countResp.Ok || countResp.Count != 1 {
		t.Fatalf("expected count 1, got %+v", countResp)
	}

	// 2. GET notifications list
	req = httptest.NewRequest("GET", "/api/v1/notifications", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var listResp struct {
		Ok            bool                      `json:"ok"`
		Notifications []store.InAppNotification `json:"notifications"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&listResp)
	if !listResp.Ok || len(listResp.Notifications) != 1 {
		t.Fatalf("expected 1 notification, got %+v", listResp)
	}
	notifID := listResp.Notifications[0].ID

	// 3. PATCH mark as read
	req = httptest.NewRequest("PATCH", "/api/v1/notifications/"+notifID+"/read", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	// Verify unread count is now 0
	req = httptest.NewRequest("GET", "/api/v1/notifications/unread-count", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	_ = json.NewDecoder(rec.Body).Decode(&countResp)
	if countResp.Count != 0 {
		t.Fatalf("expected count 0, got %d", countResp.Count)
	}

	// 4. POST mark-all-read
	req = httptest.NewRequest("POST", "/api/v1/notifications/mark-all-read", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

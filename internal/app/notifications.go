package app

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"exist.io/internal/auth"
	"exist.io/internal/store"
	"exist.io/pkg/idgen"
	"exist.io/pkg/scopes"
)

// resolveUserID determines the active user ID from the request context, session, or default admin.
func resolveUserID(r *http.Request, db *store.DB) string {
	if tok, ok := auth.ContextToken(r.Context()); ok && tok.Type() == scopes.TypeUser {
		return tok.Subject()
	}
	if cookie, err := r.Cookie("exist_session"); err == nil && cookie.Value != "" {
		hash := auth.HashCredential(cookie.Value)
		if _, user, err := db.GetSessionByHash(r.Context(), hash); err == nil && user.ID != "" {
			return user.ID
		}
	}
	if u, err := db.GetUserByLogin(r.Context(), "shiva"); err == nil {
		return u.ID
	}
	users, err := db.ListUsers(r.Context())
	if err == nil && len(users) > 0 {
		return users[0].ID
	}
	return "usr-default-admin"
}

// CreateInAppNotification creates a notification for a target user.
func (a *App) CreateInAppNotification(ctx context.Context, userID, eventType, title, body, url string) error {
	if a.Store == nil {
		return nil
	}
	n := store.InAppNotification{
		ID:        idgen.New("notif"),
		UserID:    userID,
		EventType: eventType,
		Title:     title,
		Body:      body,
		URL:       url,
		Read:      false,
	}
	return a.Store.CreateInAppNotification(ctx, n)
}

func (a *App) registerNotificationRoutes(mux *http.ServeMux, db *store.DB) {
	mux.HandleFunc("GET /api/v1/notifications", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		userID := resolveUserID(r, db)
		unreadOnly, _ := strconv.ParseBool(r.URL.Query().Get("unread"))

		list, err := db.ListInAppNotifications(r.Context(), userID, unreadOnly)
		if err != nil {
			http.Error(w, "Failed to list notifications: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if list == nil {
			list = []store.InAppNotification{}
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":            true,
			"notifications": list,
		})
	})

	mux.HandleFunc("PATCH /api/v1/notifications/{id}/read", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		id := r.PathValue("id")
		userID := resolveUserID(r, db)

		if err := db.MarkInAppNotificationRead(r.Context(), id, userID); err != nil {
			http.Error(w, "Failed to mark notification as read: "+err.Error(), http.StatusNotFound)
			return
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok": true,
		})
	})

	mux.HandleFunc("POST /api/v1/notifications/mark-all-read", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		userID := resolveUserID(r, db)
		if err := db.MarkAllInAppNotificationsRead(r.Context(), userID); err != nil {
			http.Error(w, "Failed to mark all notifications as read: "+err.Error(), http.StatusInternalServerError)
			return
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok": true,
		})
	})

	mux.HandleFunc("GET /api/v1/notifications/unread-count", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		userID := resolveUserID(r, db)
		count, err := db.CountUnreadInAppNotifications(r.Context(), userID)
		if err != nil {
			http.Error(w, "Failed to get unread count: "+err.Error(), http.StatusInternalServerError)
			return
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":    true,
			"count": count,
		})
	})
}

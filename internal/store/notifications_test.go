package store_test

import (
	"context"
	"testing"

	"exist.io/internal/store"
)

func TestInAppNotifications(t *testing.T) {
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Create user
	u := store.User{
		ID:           "usr-test-1",
		Login:        "alice",
		Email:        "alice@example.com",
		PasswordHash: "fakehash",
	}
	if err := db.CreateUserWithPassword(ctx, u); err != nil {
		t.Fatalf("create user: %v", err)
	}

	// Initial count should be 0
	count, err := db.CountUnreadInAppNotifications(ctx, u.ID)
	if err != nil {
		t.Fatalf("count unread: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 unread, got %d", count)
	}

	// Create 2 notifications
	n1 := store.InAppNotification{
		ID:        "notif-1",
		UserID:    u.ID,
		EventType: "pr.opened",
		Title:     "New Pull Request: Feature A",
		Body:      "alice opened PR #1",
		URL:       "/owner/repo/pull/1",
	}
	n2 := store.InAppNotification{
		ID:        "notif-2",
		UserID:    u.ID,
		EventType: "pr.reviewed",
		Title:     "PR #1 Approved",
		Body:      "bob approved your changes",
		URL:       "/owner/repo/pull/1",
	}
	if err := db.CreateInAppNotification(ctx, n1); err != nil {
		t.Fatalf("create notif 1: %v", err)
	}
	if err := db.CreateInAppNotification(ctx, n2); err != nil {
		t.Fatalf("create notif 2: %v", err)
	}

	// Count should be 2
	count, err = db.CountUnreadInAppNotifications(ctx, u.ID)
	if err != nil || count != 2 {
		t.Fatalf("expected 2 unread, got %d (err: %v)", count, err)
	}

	// List notifications
	list, err := db.ListInAppNotifications(ctx, u.ID, false)
	if err != nil || len(list) != 2 {
		t.Fatalf("expected 2 notifications, got %d (err: %v)", len(list), err)
	}

	// Mark n1 as read
	if err := db.MarkInAppNotificationRead(ctx, n1.ID, u.ID); err != nil {
		t.Fatalf("mark read: %v", err)
	}

	// Count should now be 1
	count, err = db.CountUnreadInAppNotifications(ctx, u.ID)
	if err != nil || count != 1 {
		t.Fatalf("expected 1 unread, got %d (err: %v)", count, err)
	}

	// List unread only
	unreadList, err := db.ListInAppNotifications(ctx, u.ID, true)
	if err != nil || len(unreadList) != 1 || unreadList[0].ID != n2.ID {
		t.Fatalf("expected only n2 unread, got: %+v", unreadList)
	}

	// Mark all read
	if err := db.MarkAllInAppNotificationsRead(ctx, u.ID); err != nil {
		t.Fatalf("mark all read: %v", err)
	}
	count, err = db.CountUnreadInAppNotifications(ctx, u.ID)
	if err != nil || count != 0 {
		t.Fatalf("expected 0 unread after mark all read, got %d", count)
	}
}

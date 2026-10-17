package notifications

// EventType represents the type of notification event.
type EventType string

const (
	EventPROpened      EventType = "pr.opened"
	EventPRReviewed    EventType = "pr.reviewed"
	EventPRMerged      EventType = "pr.merged"
	EventPRClosed      EventType = "pr.closed"
	EventIssueAssigned EventType = "issue.assigned"
)

// NotificationEvent represents a notification to be dispatched.
type NotificationEvent struct {
	Type       EventType
	RepoOwner  string
	RepoName   string
	ActorName  string
	Recipients []string          // email addresses
	Subject    string            // optional custom subject; defaults to generated subject if empty
	Payload    map[string]string // key-value pairs for template rendering
}

// Repo returns the full repository path in "owner/name" format.
func (e NotificationEvent) Repo() string {
	if e.RepoOwner != "" && e.RepoName != "" {
		return e.RepoOwner + "/" + e.RepoName
	}
	if e.RepoName != "" {
		return e.RepoName
	}
	return ""
}

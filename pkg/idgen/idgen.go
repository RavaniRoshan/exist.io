// Package idgen generates the prefixed entity IDs used across exist.io.
// Every ID is "<prefix>_<26-char ULID>" (e.g. "repo_01k..."); the ULID
// portion is produced with oklog/ulid's monotonic entropy reader so IDs
// created within the same millisecond sort strictly increasing.
package idgen

import (
	"crypto/rand"
	"io"
	"sync"
	"time"

	"github.com/oklog/ulid"
)

var (
	mu      sync.Mutex
	entropy io.Reader = ulid.Monotonic(rand.Reader, 0)
)

// New returns a unique monotonic entity ID formatted as "<prefix>_<26-character ULID>".
func New(prefix string) string {
	mu.Lock()
	id := ulid.MustNew(ulid.Timestamp(time.Now()), entropy)
	mu.Unlock()
	return prefix + "_" + id.String()
}

// NewApp generates a prefixed app ID ("app_...").
func NewApp() string             { return New("app") }
// NewInstallation generates a prefixed installation ID ("inst_...").
func NewInstallation() string    { return New("inst") }
// NewNS generates a prefixed namespace ID ("ns_...").
func NewNS() string              { return New("ns") }
func NewRepo() string            { return New("repo") }
func NewPR() string              { return New("pr") }
func NewCheckRun() string        { return New("cr") }
func NewCheckRunGroup() string   { return New("crg") }
func NewCheckRunAction() string  { return New("cra") }
func NewNSCA() string            { return New("nsca") }
func NewNSIP() string            { return New("nsip") }
func NewWebhookDelivery() string { return New("whd") }
func NewWebhookAttempt() string  { return New("wha") }
func NewEvent() string           { return New("evt") }
func NewGroup() string           { return New("grp") }
func NewUser() string            { return New("user") }
func NewStack() string           { return New("stk") }
func NewInstallToken() string    { return New("oit") }
func NewPRComment() string       { return New("prc") }
func NewPRReview() string        { return New("prr") }
func NewPRThread() string        { return New("prt") }
func NewLabel() string              { return New("lbl") }
func NewReaction() string           { return New("rxn") }
func NewRuleset() string            { return New("rs") }
func NewRulesetRule() string        { return New("rsr") }
func NewRulesetBypassActor() string { return New("rsba") }
func NewCheckRunAnnotation() string { return New("cra") }
func NewIssue() string              { return New("issue") }
func NewIssueComment() string       { return New("ic") }
func NewGrant() string              { return New("grant") }

package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/go-git/go-git/v6/plumbing/object"

	"exist.io/internal/engine"
	"exist.io/internal/gitstore"
	"exist.io/internal/store"
)

// AutoMergeEvent represents a trigger for auto-merge evaluation.
type AutoMergeEvent struct {
	Type          string // "check_completed" | "review_submitted"
	RepoID        string
	PullRequestID string
}

// AutoMergeEngine listens for check-run and review events and executes
// pending auto-merges when all merge requirements are satisfied.
type AutoMergeEngine struct {
	db      *store.DB
	git     *gitstore.Store
	eng     *engine.Engine
	eventCh chan AutoMergeEvent
	stopCh  chan struct{}
	wg      sync.WaitGroup

	locksMu sync.Mutex
	prLocks map[string]*sync.Mutex
}

// NewAutoMergeEngine creates a new auto-merge engine.
func NewAutoMergeEngine(db *store.DB, git *gitstore.Store, eng *engine.Engine) *AutoMergeEngine {
	return &AutoMergeEngine{
		db:      db,
		git:     git,
		eng:     eng,
		eventCh: make(chan AutoMergeEvent, 256),
		stopCh:  make(chan struct{}),
		prLocks: make(map[string]*sync.Mutex),
	}
}

// Start begins the auto-merge worker goroutine.
func (e *AutoMergeEngine) Start() {
	e.wg.Add(1)
	go e.run()
}

// Stop gracefully shuts down the auto-merge engine.
func (e *AutoMergeEngine) Stop() {
	close(e.stopCh)
	e.wg.Wait()
}

// Notify sends an event to the auto-merge engine for evaluation.
func (e *AutoMergeEngine) Notify(evt AutoMergeEvent) {
	select {
	case e.eventCh <- evt:
	default:
		// Channel full, skip (non-blocking)
	}
}

func (e *AutoMergeEngine) run() {
	defer e.wg.Done()
	for {
		select {
		case <-e.stopCh:
			return
		case evt := <-e.eventCh:
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := e.evaluate(ctx, evt); err != nil {
				log.Printf("automerge: evaluate %s/%s: %v", evt.RepoID, evt.PullRequestID, err)
			}
			cancel()
		}
	}
}

// getPRLock returns a mutex dedicated to the given PR ID to prevent concurrent merges.
func (e *AutoMergeEngine) getPRLock(prID string) *sync.Mutex {
	e.locksMu.Lock()
	defer e.locksMu.Unlock()
	if e.prLocks == nil {
		e.prLocks = make(map[string]*sync.Mutex)
	}
	lk, ok := e.prLocks[prID]
	if !ok {
		lk = &sync.Mutex{}
		e.prLocks[prID] = lk
	}
	return lk
}

// Evaluate performs immediate synchronous evaluation of an auto-merge event.
// Exposed for testing and direct invocation.
func (e *AutoMergeEngine) Evaluate(ctx context.Context, evt AutoMergeEvent) error {
	return e.evaluate(ctx, evt)
}

func (e *AutoMergeEngine) evaluate(ctx context.Context, evt AutoMergeEvent) error {
	if evt.PullRequestID == "" {
		return nil
	}

	prLock := e.getPRLock(evt.PullRequestID)
	prLock.Lock()
	defer prLock.Unlock()

	// 1. Get the auto-merge config for the PR. If nil, return (not enabled).
	if e.db == nil {
		return nil
	}
	am, err := e.db.GetAutoMerge(ctx, evt.PullRequestID)
	if err != nil {
		return fmt.Errorf("get auto-merge config: %w", err)
	}
	if am == nil {
		return nil
	}

	// 2. Load the PR. Check it's still open and not merged.
	pr, err := e.db.GetPullRequestByID(ctx, evt.PullRequestID)
	if err != nil {
		return fmt.Errorf("get pull request %s: %w", evt.PullRequestID, err)
	}
	if pr.Merged || pr.State != "open" || pr.Draft {
		return nil
	}

	// 3. Load all check runs for the PR. If any are not "success", return.
	checkRuns, err := e.db.ListCheckRuns(ctx, pr.RepoID, pr.HeadSHA)
	if err != nil {
		return fmt.Errorf("list check runs for commit %s: %w", pr.HeadSHA, err)
	}
	latestRuns := make(map[string]store.CheckRun)
	for _, cr := range checkRuns {
		key := cr.Key
		if key == "" {
			key = cr.Name
		}
		if key == "" {
			key = cr.ID
		}
		if _, ok := latestRuns[key]; !ok {
			latestRuns[key] = cr
		}
	}
	for _, cr := range latestRuns {
		if cr.Status != "completed" || cr.Conclusion != "success" {
			return nil
		}
	}

	// 4. Load all reviews. Count approvals. If there are any "changes_requested" reviews that aren't dismissed, return.
	reviews, err := e.db.ListReviews(ctx, pr.ID)
	if err != nil {
		return fmt.Errorf("list reviews for PR %s: %w", pr.ID, err)
	}
	approvedCount := 0
	for _, rev := range reviews {
		if rev.Dismissed {
			continue
		}
		if rev.State == "changes_requested" {
			return nil
		}
		if rev.State == "approved" || rev.State == "approve" {
			approvedCount++
		}
	}

	// 5. Check required review count from rulesets (use e.eng engine if available, or default to 0).
	requiredReviews := 0
	if e.eng != nil {
		requiredReviews = e.eng.RequiredReviewCount(ctx, pr.RepoID, pr.BaseRef)
	}
	if approvedCount < requiredReviews {
		return nil
	}

	repo, err := e.db.GetRepoByID(ctx, pr.RepoID)
	if err != nil {
		return fmt.Errorf("get repo %s: %w", pr.RepoID, err)
	}

	// If engine is available, also verify full ruleset mergeability
	if e.eng != nil {
		rulesets, err := e.db.ListRulesets(ctx, pr.RepoID)
		if err == nil && len(rulesets) > 0 {
			res := e.eng.EvaluateMerge(ctx, rulesets, repo.DefaultBranch, pr.BaseRef, "", "", approvedCount, checkRuns)
			if !res.Allowed {
				return nil
			}
		}
	}

	// 6. If all conditions pass, execute the merge:
	nsSlug := ""
	if ns, err := e.db.GetNamespace(ctx, repo.NamespaceID); err == nil && ns.Slug != "" {
		nsSlug = ns.Slug
	} else if strings.Contains(repo.FullName, "/") {
		nsSlug = strings.Split(repo.FullName, "/")[0]
	}
	repoName := repo.Name

	method := am.MergeMethod
	if method == "" {
		if repo.AllowMergeCommit {
			method = "merge"
		} else if repo.AllowSquashMerge {
			method = "squash"
		} else {
			method = "squash"
		}
	}

	authorSig := object.Signature{
		Name:  "Exist Engine",
		Email: "merge@exist.io",
		When:  time.Now().UTC(),
	}

	msg := fmt.Sprintf("Merge pull request #%d from %s\n\n%s", pr.Number, pr.HeadRef, pr.Title)
	if am.CommitTitle != "" {
		msg = am.CommitTitle
		if am.CommitMessage != "" {
			msg += "\n\n" + am.CommitMessage
		}
	}

	var mergeCommitSHA string
	if e.git != nil {
		mergeable, _, _, err := e.git.CheckMergeable(nsSlug, repoName, pr.BaseRef, pr.HeadRef)
		if err == nil && !mergeable {
			// Conflicts prevent auto-merge
			return nil
		}

		sha, err := e.git.Merge(nsSlug, repoName, pr.BaseRef, pr.HeadRef, method, authorSig, msg)
		if err != nil {
			if errors.Is(err, gitstore.ErrMergeConflict) {
				return nil
			}
			return fmt.Errorf("git merge: %w", err)
		}
		mergeCommitSHA = sha
	} else {
		mergeCommitSHA = "0000000000000000000000000000000000000000"
	}

	if err := e.db.SetPullRequestMergeResult(ctx, pr.ID, mergeCommitSHA); err != nil {
		return fmt.Errorf("set pull request merge result: %w", err)
	}

	if err := e.db.DeleteAutoMerge(ctx, pr.ID); err != nil {
		return fmt.Errorf("delete auto merge: %w", err)
	}

	return nil
}

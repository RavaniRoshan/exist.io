package origin

import (
	"context"
	"errors"
	"net/http"

	"exist.io/internal/auth"
	"exist.io/internal/store"
	"exist.io/pkg/scopes"
)

var (
	reqChecksRead = scopes.Requirements{
		TokenTypes: []string{scopes.TypeInstallation, scopes.TypeUser},
		Scopes:     []string{scopes.RepositoryChecksRead},
	}
	reqChecksWrite = scopes.Requirements{
		TokenTypes: []string{scopes.TypeInstallation, scopes.TypeUser},
		Scopes:     []string{scopes.RepositoryChecksWrite},
	}
)

func (s *Server) registerChecks(mux *http.ServeMux) {
	s.register("OriginService_PostCheckRun", reqChecksWrite)
	s.register("OriginService_BatchUpsertCheckRuns", reqChecksWrite)
	s.register("OriginService_GetCheckRun", reqChecksRead)
	s.register("OriginService_CreateCheckRunAnnotations", reqChecksWrite)
	s.register("OriginService_ListCheckRunAnnotations", reqChecksRead)
	s.register("OriginService_RerequestCheckRun", reqPRContentsWrite)
	s.register("OriginService_GetCheckSuite", reqChecksRead)
	s.register("OriginService_ListCheckRunsForSuite", reqChecksRead)
	s.register("OriginService_ListCheckRunsForCommit", reqChecksRead)
	s.register("OriginService_ListCheckSuitesForCommit", reqChecksRead)

	mux.Handle("POST /v1/origin/repos/{ownerSlug}/{repoName}/check-runs",
		s.guard("OriginService_PostCheckRun", s.postCheckRun))
	mux.Handle("POST /v1/origin/repos/{ownerSlug}/{repoName}/check-runs:batchUpsert",
		s.guard("OriginService_BatchUpsertCheckRuns", s.batchUpsertCheckRuns))
	mux.Handle("GET /v1/origin/repos/{ownerSlug}/{repoName}/check-runs/{checkRunId}",
		s.guard("OriginService_GetCheckRun", s.getCheckRun))
	mux.Handle("POST /v1/origin/repos/{ownerSlug}/{repoName}/check-runs/{checkRunId}/annotations",
		s.guard("OriginService_CreateCheckRunAnnotations", s.createCheckRunAnnotations))
	mux.Handle("GET /v1/origin/repos/{ownerSlug}/{repoName}/check-runs/{checkRunId}/annotations",
		s.guard("OriginService_ListCheckRunAnnotations", s.listCheckRunAnnotations))
	mux.Handle("POST /v1/origin/repos/{ownerSlug}/{repoName}/check-runs/{checkRunId}/rerequest",
		s.guard("OriginService_RerequestCheckRun", s.rerequestCheckRun))
	mux.Handle("GET /v1/origin/repos/{ownerSlug}/{repoName}/check-suites/{checkSuiteId}",
		s.guard("OriginService_GetCheckSuite", s.getCheckSuite))
	mux.Handle("GET /v1/origin/repos/{ownerSlug}/{repoName}/check-suites/{checkSuiteId}/check-runs",
		s.guard("OriginService_ListCheckRunsForSuite", s.listCheckRunsForSuite))
	mux.Handle("GET /v1/origin/repos/{ownerSlug}/{repoName}/commits/{sha}/check-runs",
		s.guard("OriginService_ListCheckRunsForCommit", s.listCheckRunsForCommit))
	mux.Handle("GET /v1/origin/repos/{ownerSlug}/{repoName}/commits/{sha}/check-suites",
		s.guard("OriginService_ListCheckSuitesForCommit", s.listCheckSuitesForCommit))
}

type checkSuiteInput struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	DetailsURL string `json:"detailsUrl,omitempty"`
	ExternalID string `json:"externalId,omitempty"`
}

type checkRunInput struct {
	Key               string `json:"key"`
	Name              string `json:"name"`
	Status            string `json:"status"`
	Conclusion        string `json:"conclusion,omitempty"`
	DetailsURL        string `json:"detailsUrl,omitempty"`
	ExternalID        string `json:"externalId,omitempty"`
	ExternalUpdatedAt string `json:"externalUpdatedAt,omitempty"`
	StartedAt         string `json:"startedAt,omitempty"`
	CompletedAt       string `json:"completedAt,omitempty"`
	DeadlineAt        string `json:"deadlineAt,omitempty"`
	IsRerequestable   bool   `json:"isRerequestable,omitempty"`
	Output            *struct {
		Title   string `json:"title,omitempty"`
		Summary string `json:"summary,omitempty"`
		Text    string `json:"text,omitempty"`
	} `json:"output,omitempty"`
}

type postCheckRunRequest struct {
	HeadSHA    string          `json:"headSha"`
	BaseSHA    string          `json:"baseSha,omitempty"`
	CheckSuite checkSuiteInput `json:"checkSuite"`
	CheckRun   checkRunInput   `json:"checkRun"`
}

type batchUpsertCheckRunsRequest struct {
	HeadSHA    string          `json:"headSha"`
	BaseSHA    string          `json:"baseSha,omitempty"`
	CheckSuite checkSuiteInput `json:"checkSuite"`
	CheckRuns  []checkRunInput `json:"checkRuns"`
}

type annotationInput struct {
	AnnotationLevel string `json:"annotationLevel"`
	Message         string `json:"message"`
	Title           string `json:"title,omitempty"`
	RawDetails      string `json:"rawDetails,omitempty"`
	Location        *struct {
		Path      string `json:"path"`
		StartLine int    `json:"startLine"`
		EndLine   int    `json:"endLine"`
		Columns   *struct {
			StartColumn int `json:"startColumn,omitempty"`
			EndColumn   int `json:"endColumn,omitempty"`
		} `json:"columns,omitempty"`
	} `json:"location,omitempty"`
}

type createAnnotationsRequest struct {
	Annotations []annotationInput `json:"annotations"`
}

func (s *Server) postCheckRun(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	repo, ns, err := s.loadRepoAndNS(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	var req postCheckRunRequest
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}
	if req.HeadSHA == "" || req.CheckSuite.Key == "" || req.CheckRun.Key == "" || req.CheckRun.Name == "" {
		badRequest(w, "missing required fields: headSha, checkSuite.key, checkRun.key, checkRun.name")
		return
	}

	actingUser := ""
	if tok, ok := auth.ContextToken(r.Context()); ok {
		actingUser = tok.Subject()
	}

	suiteStore := store.CheckSuite{
		RepoID:      repo.ID,
		SHA:         req.HeadSHA,
		BaseSHA:     req.BaseSHA,
		Key:         req.CheckSuite.Key,
		Name:        req.CheckSuite.Name,
		DetailsURL:  req.CheckSuite.DetailsURL,
		ExternalID:  req.CheckSuite.ExternalID,
		ActorUserID: actingUser,
	}
	suite, err := s.store.UpsertCheckSuite(r.Context(), suiteStore)
	if err != nil {
		internalError(w, "upsert check suite: "+err.Error())
		return
	}

	runStore := store.CheckRun{
		RepoID:            repo.ID,
		CheckSuiteID:      suite.ID,
		SHA:               req.HeadSHA,
		BaseSHA:           req.BaseSHA,
		Key:               req.CheckRun.Key,
		Name:              req.CheckRun.Name,
		Status:            req.CheckRun.Status,
		Conclusion:        req.CheckRun.Conclusion,
		DetailsURL:        req.CheckRun.DetailsURL,
		ExternalID:        req.CheckRun.ExternalID,
		ExternalUpdatedAt: req.CheckRun.ExternalUpdatedAt,
		StartedAt:         req.CheckRun.StartedAt,
		CompletedAt:       req.CheckRun.CompletedAt,
		DeadlineAt:        req.CheckRun.DeadlineAt,
		IsRerequestable:   req.CheckRun.IsRerequestable,
		ActorUserID:       actingUser,
	}
	if req.CheckRun.Output != nil {
		runStore.OutputTitle = req.CheckRun.Output.Title
		runStore.OutputSummary = req.CheckRun.Output.Summary
		runStore.OutputText = req.CheckRun.Output.Text
	}

	storedRun, outcome, err := s.store.UpsertCheckRun(r.Context(), runStore)
	if err != nil {
		internalError(w, "upsert check run: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"checkSuite": toCheckSuiteDTO(suite, repo, ns),
		"checkRun":   toCheckRunDTO(storedRun, repo, ns),
		"outcome":    outcome,
	})
}

func (s *Server) batchUpsertCheckRuns(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	repo, ns, err := s.loadRepoAndNS(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	var req batchUpsertCheckRunsRequest
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}
	if req.HeadSHA == "" || req.CheckSuite.Key == "" || len(req.CheckRuns) == 0 {
		badRequest(w, "missing required fields: headSha, checkSuite.key, checkRuns")
		return
	}

	actingUser := ""
	if tok, ok := auth.ContextToken(r.Context()); ok {
		actingUser = tok.Subject()
	}

	suiteStore := store.CheckSuite{
		RepoID:      repo.ID,
		SHA:         req.HeadSHA,
		BaseSHA:     req.BaseSHA,
		Key:         req.CheckSuite.Key,
		Name:        req.CheckSuite.Name,
		DetailsURL:  req.CheckSuite.DetailsURL,
		ExternalID:  req.CheckSuite.ExternalID,
		ActorUserID: actingUser,
	}
	suite, err := s.store.UpsertCheckSuite(r.Context(), suiteStore)
	if err != nil {
		internalError(w, "upsert check suite: "+err.Error())
		return
	}

	var runsDTO []map[string]any
	var resultsDTO []map[string]any

	for _, cr := range req.CheckRuns {
		runStore := store.CheckRun{
			RepoID:            repo.ID,
			CheckSuiteID:      suite.ID,
			SHA:               req.HeadSHA,
			BaseSHA:           req.BaseSHA,
			Key:               cr.Key,
			Name:              cr.Name,
			Status:            cr.Status,
			Conclusion:        cr.Conclusion,
			DetailsURL:        cr.DetailsURL,
			ExternalID:        cr.ExternalID,
			ExternalUpdatedAt: cr.ExternalUpdatedAt,
			StartedAt:         cr.StartedAt,
			CompletedAt:       cr.CompletedAt,
			DeadlineAt:        cr.DeadlineAt,
			IsRerequestable:   cr.IsRerequestable,
			ActorUserID:       actingUser,
		}
		if cr.Output != nil {
			runStore.OutputTitle = cr.Output.Title
			runStore.OutputSummary = cr.Output.Summary
			runStore.OutputText = cr.Output.Text
		}

		storedRun, outcome, err := s.store.UpsertCheckRun(r.Context(), runStore)
		if err != nil {
			internalError(w, "upsert check run: "+err.Error())
			return
		}

		runDTO := toCheckRunDTO(storedRun, repo, ns)
		runsDTO = append(runsDTO, runDTO)
		resultsDTO = append(resultsDTO, map[string]any{
			"checkRun": runDTO,
			"outcome":  outcome,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"checkSuite": toCheckSuiteDTO(suite, repo, ns),
		"checkRuns":  runsDTO,
		"results":    resultsDTO,
	})
}

func (s *Server) getCheckRun(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	repo, ns, err := s.loadRepoAndNS(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	checkRunID := r.PathValue("checkRunId")
	run, err := s.store.GetCheckRun(r.Context(), checkRunID)
	if err != nil {
		notFound(w, "check run not found")
		return
	}
	if run.RepoID != repo.ID {
		notFound(w, "check run not found")
		return
	}

	writeJSON(w, http.StatusOK, toCheckRunDTO(run, repo, ns))
}

func (s *Server) createCheckRunAnnotations(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	repo, _, err := s.loadRepoAndNS(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	checkRunID := r.PathValue("checkRunId")
	run, err := s.store.GetCheckRun(r.Context(), checkRunID)
	if err != nil || run.RepoID != repo.ID {
		notFound(w, "check run not found")
		return
	}

	var req createAnnotationsRequest
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}

	var storeAnns []store.CheckRunAnnotation
	for _, a := range req.Annotations {
		sa := store.CheckRunAnnotation{
			AnnotationLevel: a.AnnotationLevel,
			Message:         a.Message,
			Title:           a.Title,
			RawDetails:      a.RawDetails,
		}
		if a.Location != nil {
			sa.Path = a.Location.Path
			sa.StartLine = a.Location.StartLine
			sa.EndLine = a.Location.EndLine
			if a.Location.Columns != nil {
				sa.StartColumn = a.Location.Columns.StartColumn
				sa.EndColumn = a.Location.Columns.EndColumn
			}
		}
		storeAnns = append(storeAnns, sa)
	}

	created, err := s.store.CreateCheckRunAnnotations(r.Context(), checkRunID, storeAnns)
	if err != nil {
		internalError(w, "create annotations: "+err.Error())
		return
	}

	var dtos []map[string]any
	for _, a := range created {
		dtos = append(dtos, toAnnotationDTO(a))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"annotations": dtos,
	})
}

func (s *Server) listCheckRunAnnotations(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	repo, _, err := s.loadRepoAndNS(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	checkRunID := r.PathValue("checkRunId")
	run, err := s.store.GetCheckRun(r.Context(), checkRunID)
	if err != nil || run.RepoID != repo.ID {
		notFound(w, "check run not found")
		return
	}

	anns, err := s.store.ListCheckRunAnnotations(r.Context(), checkRunID, 100, 0)
	if err != nil {
		internalError(w, "list annotations: "+err.Error())
		return
	}

	var dtos []map[string]any
	for _, a := range anns {
		dtos = append(dtos, toAnnotationDTO(a))
	}
	if dtos == nil {
		dtos = []map[string]any{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"annotations":   dtos,
		"nextPageToken": "",
	})
}

func (s *Server) rerequestCheckRun(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	repo, ns, err := s.loadRepoAndNS(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	checkRunID := r.PathValue("checkRunId")
	run, err := s.store.GetCheckRun(r.Context(), checkRunID)
	if err != nil || run.RepoID != repo.ID {
		notFound(w, "check run not found")
		return
	}

	actingUser := ""
	if tok, ok := auth.ContextToken(r.Context()); ok {
		actingUser = tok.Subject()
	}

	updatedRun, err := s.store.RerequestCheckRun(r.Context(), checkRunID, actingUser)
	if err != nil {
		if errors.Is(err, store.ErrPrecondition) {
			badRequest(w, err.Error())
			return
		}
		if errors.Is(err, store.ErrAlreadyExists) {
			conflict(w, err.Error())
			return
		}
		internalError(w, "rerequest check run: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toCheckRunDTO(updatedRun, repo, ns))
}

func (s *Server) getCheckSuite(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	repo, ns, err := s.loadRepoAndNS(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	checkSuiteID := r.PathValue("checkSuiteId")
	suite, err := s.store.GetCheckSuite(r.Context(), checkSuiteID)
	if err != nil || suite.RepoID != repo.ID {
		notFound(w, "check suite not found")
		return
	}

	writeJSON(w, http.StatusOK, toCheckSuiteDTO(suite, repo, ns))
}

func (s *Server) listCheckRunsForSuite(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	repo, ns, err := s.loadRepoAndNS(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	checkSuiteID := r.PathValue("checkSuiteId")
	suite, err := s.store.GetCheckSuite(r.Context(), checkSuiteID)
	if err != nil || suite.RepoID != repo.ID {
		notFound(w, "check suite not found")
		return
	}

	runs, err := s.store.ListCheckRunsForSuite(r.Context(), checkSuiteID, 100, 0)
	if err != nil {
		internalError(w, "list check runs: "+err.Error())
		return
	}

	var dtos []map[string]any
	for _, run := range runs {
		dtos = append(dtos, toCheckRunDTO(run, repo, ns))
	}
	if dtos == nil {
		dtos = []map[string]any{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"checkRuns":     dtos,
		"nextPageToken": "",
	})
}

func (s *Server) listCheckRunsForCommit(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	repo, ns, err := s.loadRepoAndNS(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	sha := r.PathValue("sha")
	checkName := r.URL.Query().Get("checkName")
	status := r.URL.Query().Get("status")

	runs, err := s.store.ListCheckRunsForCommit(r.Context(), repo.ID, sha, checkName, status, 100, 0)
	if err != nil {
		internalError(w, "list check runs for commit: "+err.Error())
		return
	}

	var dtos []map[string]any
	for _, run := range runs {
		dtos = append(dtos, toCheckRunDTO(run, repo, ns))
	}
	if dtos == nil {
		dtos = []map[string]any{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"checkRuns":     dtos,
		"nextPageToken": "",
	})
}

func (s *Server) listCheckSuitesForCommit(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	repo, ns, err := s.loadRepoAndNS(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	sha := r.PathValue("sha")
	suites, err := s.store.ListCheckSuitesForCommit(r.Context(), repo.ID, sha, 100, 0)
	if err != nil {
		internalError(w, "list check suites for commit: "+err.Error())
		return
	}

	var dtos []map[string]any
	for _, suite := range suites {
		dtos = append(dtos, toCheckSuiteDTO(suite, repo, ns))
	}
	if dtos == nil {
		dtos = []map[string]any{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"checkSuites":   dtos,
		"nextPageToken": "",
	})
}

func (s *Server) loadRepoAndNS(ctx context.Context, owner, name string) (store.Repo, store.Namespace, error) {
	repo, err := s.store.GetRepo(ctx, owner, name)
	if err != nil {
		return store.Repo{}, store.Namespace{}, err
	}
	ns, err := s.store.GetNamespace(ctx, repo.NamespaceID)
	if err != nil {
		return store.Repo{}, store.Namespace{}, err
	}
	return repo, ns, nil
}

func toCheckSuiteDTO(suite store.CheckSuite, repo store.Repo, ns store.Namespace) map[string]any {
	dto := map[string]any{
		"id": suite.ID,
		"repository": map[string]any{
			"id":   repo.ID,
			"name": repo.Name,
			"owner": map[string]any{
				"id":   ns.ID,
				"slug": ns.Slug,
				"type": ns.Type,
			},
		},
		"sha":        suite.SHA,
		"key":        suite.Key,
		"name":       suite.Name,
		"detailsUrl": suite.DetailsURL,
		"createdAt":  suite.CreatedAt,
		"updatedAt":  suite.UpdatedAt,
		"externalId": suite.ExternalID,
	}
	if suite.BaseSHA != "" {
		dto["baseSha"] = suite.BaseSHA
	}
	if suite.ActorUserID != "" {
		dto["actor"] = map[string]any{
			"user": map[string]any{
				"id": suite.ActorUserID,
			},
		}
	}
	return dto
}

func toCheckRunDTO(run store.CheckRun, repo store.Repo, ns store.Namespace) map[string]any {
	dto := map[string]any{
		"id": run.ID,
		"repository": map[string]any{
			"id":   repo.ID,
			"name": repo.Name,
			"owner": map[string]any{
				"id":   ns.ID,
				"slug": ns.Slug,
				"type": ns.Type,
			},
		},
		"checkSuite": map[string]any{
			"id": run.CheckSuiteID,
		},
		"sha":               run.SHA,
		"key":               run.Key,
		"name":              run.Name,
		"status":            run.Status,
		"conclusion":        run.Conclusion,
		"detailsUrl":        run.DetailsURL,
		"externalUpdatedAt": run.ExternalUpdatedAt,
		"startedAt":         run.StartedAt,
		"completedAt":       run.CompletedAt,
		"createdAt":         run.CreatedAt,
		"updatedAt":         run.UpdatedAt,
		"externalId":        run.ExternalID,
		"isRerequestable":   run.IsRerequestable,
	}
	if run.BaseSHA != "" {
		dto["baseSha"] = run.BaseSHA
	}
	if run.OutputTitle != "" || run.OutputSummary != "" || run.OutputText != "" {
		dto["output"] = map[string]any{
			"title":   run.OutputTitle,
			"summary": run.OutputSummary,
			"text":    run.OutputText,
		}
	}
	if run.RerequestedAt != "" {
		dto["rerequestedAt"] = run.RerequestedAt
		if run.RerequestedByUserID != "" {
			dto["rerequestedBy"] = map[string]any{
				"user": map[string]any{
					"id": run.RerequestedByUserID,
				},
			}
		}
	}
	if run.ActorUserID != "" {
		dto["actor"] = map[string]any{
			"user": map[string]any{
				"id": run.ActorUserID,
			},
		}
	}
	return dto
}

func toAnnotationDTO(a store.CheckRunAnnotation) map[string]any {
	dto := map[string]any{
		"id":              a.ID,
		"checkRunId":      a.CheckRunID,
		"annotationLevel": a.AnnotationLevel,
		"message":         a.Message,
		"title":           a.Title,
		"rawDetails":      a.RawDetails,
		"createdAt":       a.CreatedAt,
		"updatedAt":       a.UpdatedAt,
	}
	if a.Path != "" {
		loc := map[string]any{
			"path":      a.Path,
			"startLine": a.StartLine,
			"endLine":   a.EndLine,
		}
		if a.StartColumn > 0 || a.EndColumn > 0 {
			loc["columns"] = map[string]any{
				"startColumn": a.StartColumn,
				"endColumn":   a.EndColumn,
			}
		}
		dto["location"] = loc
	}
	return dto
}

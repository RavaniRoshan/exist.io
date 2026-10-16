package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// CLI is the runner for the origin CLI client.
type CLI struct {
	Endpoint    string
	Token       string
	DefaultRepo string
	HTTPClient  *http.Client
	Stdout      io.Writer
	Stderr      io.Writer
}

// New creates a new CLI instance configured from the environment.
func New() *CLI {
	endpoint := os.Getenv("EXIST_ENDPOINT")
	if endpoint == "" {
		endpoint = os.Getenv("CURSOR_ORIGIN_ENDPOINT")
	}
	if endpoint == "" {
		endpoint = "http://localhost:8080/v1/origin"
	}
	endpoint = strings.TrimRight(endpoint, "/")

	token := os.Getenv("EXIST_TOKEN")
	if token == "" {
		token = os.Getenv("CURSOR_AUTH_TOKEN")
	}
	if token == "" {
		token = os.Getenv("CURSOR_API_KEY")
	}

	defaultRepo := os.Getenv("EXIST_REPO")
	if defaultRepo == "" {
		defaultRepo = os.Getenv("ORIGIN_REPO")
	}

	return &CLI{
		Endpoint:    endpoint,
		Token:       token,
		DefaultRepo: defaultRepo,
		HTTPClient:  http.DefaultClient,
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
	}
}

// Run executes the command specified by args.
func (c *CLI) Run(args []string) error {
	if len(args) == 0 {
		return c.printHelp()
	}

	// Global flags check
	if args[0] == "--help" || args[0] == "-h" {
		return c.printHelp()
	}
	if args[0] == "--version" || args[0] == "-v" {
		fmt.Fprintln(c.Stdout, "exist version 2026.10.14")
		return nil
	}

	switch args[0] {
	case "auth":
		return c.runAuth(args[1:])
	case "repo":
		return c.runRepo(args[1:])
	case "pr":
		return c.runPR(args[1:])
	case "app":
		return c.runApp(args[1:])
	case "api":
		return c.runAPI(args[1:])
	case "conformance":
		return c.runConformance(args[1:])
	default:
		return fmt.Errorf("unknown command %q. Run 'exist --help' for usage", args[0])
	}
}

func (c *CLI) printHelp() error {
	help := `exist <command>

Commands:
  exist auth               Manage authentication
  exist repo               Manage repositories
  exist pr                 Manage pull requests
  exist app                Manage apps
  exist api <endpoint>     Make an authenticated request to REST API
  exist conformance        Run API conformance test suite against existd

Options:
  --version                 Show version number
  --help                    Show help
`
	fmt.Fprint(c.Stdout, help)
	return nil
}

// Request helper
func (c *CLI) doRequest(method, path string, body any) ([]byte, int, error) {
	url := c.Endpoint + path
	var bodyReader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		bodyReader = bytes.NewReader(raw)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, 0, err
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	return respBytes, resp.StatusCode, err
}

// Auth commands
func (c *CLI) runAuth(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: origin auth <status|token|login|logout>")
	}
	switch args[0] {
	case "status":
		if c.Token != "" {
			fmt.Fprintf(c.Stdout, "Logged in to %s (token configured)\n", c.Endpoint)
		} else {
			fmt.Fprintf(c.Stdout, "Not logged in (target: %s)\n", c.Endpoint)
		}
		return nil
	case "token":
		if c.Token != "" {
			fmt.Fprintln(c.Stdout, c.Token)
			return nil
		}
		return errors.New("no auth token configured")
	case "login":
		fmt.Fprintln(c.Stdout, "Set CURSOR_AUTH_TOKEN or CURSOR_API_KEY in your environment.")
		return nil
	case "logout":
		c.Token = ""
		fmt.Fprintln(c.Stdout, "Logged out.")
		return nil
	default:
		return fmt.Errorf("unknown auth subcommand %q", args[0])
	}
}

// Repo commands
func (c *CLI) runRepo(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: origin repo <list|view|create|delete|mirror>")
	}
	switch args[0] {
	case "list":
		fs := flag.NewFlagSet("repo list", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		ns := fs.String("namespace", "", "Namespace slug")
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		path := "/repos"
		if *ns != "" {
			path = "/repos/" + *ns
		}
		data, code, err := c.doRequest(http.MethodGet, path, nil)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		fmt.Fprintln(c.Stdout, string(data))
		return nil

	case "view":
		if len(args) < 2 {
			return errors.New("usage: origin repo view <owner/name>")
		}
		owner, name, err := splitRepo(args[1])
		if err != nil {
			return err
		}
		data, code, err := c.doRequest(http.MethodGet, fmt.Sprintf("/repos/%s/%s", owner, name), nil)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		fmt.Fprintln(c.Stdout, string(data))
		return nil

	case "create":
		fs := flag.NewFlagSet("repo create", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		branch := fs.String("default-branch", "main", "Default branch")
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		if fs.NArg() < 1 {
			return errors.New("usage: origin repo create <owner/name>")
		}
		owner, name, err := splitRepo(fs.Arg(0))
		if err != nil {
			return err
		}
		reqBody := map[string]any{
			"repo": map[string]any{
				"name":          name,
				"defaultBranch": *branch,
			},
		}
		data, code, err := c.doRequest(http.MethodPost, "/repos/"+owner, reqBody)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		fmt.Fprintln(c.Stdout, string(data))
		return nil

	case "delete":
		if len(args) < 2 {
			return errors.New("usage: origin repo delete <owner/name>")
		}
		owner, name, err := splitRepo(args[1])
		if err != nil {
			return err
		}
		data, code, err := c.doRequest(http.MethodDelete, fmt.Sprintf("/repos/%s/%s", owner, name), nil)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		fmt.Fprintf(c.Stdout, "Deleted repository %s/%s\n", owner, name)
		return nil

	case "mirror":
		return c.runRepoMirror(args[1:])

	default:
		return fmt.Errorf("unknown repo subcommand %q", args[0])
	}
}

func (c *CLI) runRepoMirror(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: origin repo mirror <status|sync|detach> <owner/name>")
	}
	sub := args[0]
	if len(args) < 2 {
		return fmt.Errorf("usage: origin repo mirror %s <owner/name>", sub)
	}
	owner, name, err := splitRepo(args[1])
	if err != nil {
		return err
	}

	switch sub {
	case "status":
		data, code, err := c.doRequest(http.MethodGet, fmt.Sprintf("/repos/%s/%s/mirror", owner, name), nil)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		fmt.Fprintln(c.Stdout, string(data))
		return nil

	case "sync":
		ref := "refs/heads/main"
		if len(args) >= 4 && args[2] == "--ref" {
			ref = args[3]
		}
		reqBody := map[string]any{
			"ref":  ref,
			"wait": true,
		}
		data, code, err := c.doRequest(http.MethodPost, fmt.Sprintf("/repos/%s/%s:syncMirror", owner, name), reqBody)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		fmt.Fprintln(c.Stdout, string(data))
		return nil

	case "detach":
		data, code, err := c.doRequest(http.MethodDelete, fmt.Sprintf("/repos/%s/%s/mirror", owner, name), nil)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		fmt.Fprintf(c.Stdout, "Detached mirror for %s/%s\n", owner, name)
		return nil

	default:
		return fmt.Errorf("unknown mirror subcommand %q", sub)
	}
}

// Pull request commands
func (c *CLI) runPR(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: origin pr <list|view|create|merge|close>")
	}
	switch args[0] {
	case "list":
		fs := flag.NewFlagSet("pr list", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		repoFlag := fs.String("R", c.DefaultRepo, "Repository (owner/name)")
		repoLong := fs.String("repo", c.DefaultRepo, "Repository (owner/name)")
		status := fs.String("status", "open", "Filter by status")
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		repoStr := *repoFlag
		if *repoLong != "" && *repoLong != c.DefaultRepo {
			repoStr = *repoLong
		}
		owner, name, err := splitRepo(repoStr)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/repos/%s/%s/pulls?status=%s", owner, name, *status)
		data, code, err := c.doRequest(http.MethodGet, path, nil)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		fmt.Fprintln(c.Stdout, string(data))
		return nil

	case "view":
		fs := flag.NewFlagSet("pr view", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		repoFlag := fs.String("R", c.DefaultRepo, "Repository (owner/name)")
		repoLong := fs.String("repo", c.DefaultRepo, "Repository (owner/name)")
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		if fs.NArg() < 1 {
			return errors.New("usage: origin pr view <number> [-R owner/name]")
		}
		num := fs.Arg(0)
		repoStr := *repoFlag
		if *repoLong != "" && *repoLong != c.DefaultRepo {
			repoStr = *repoLong
		}
		owner, name, err := splitRepo(repoStr)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/repos/%s/%s/pulls/%s", owner, name, num)
		data, code, err := c.doRequest(http.MethodGet, path, nil)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		fmt.Fprintln(c.Stdout, string(data))
		return nil

	case "create":
		fs := flag.NewFlagSet("pr create", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		repoFlag := fs.String("R", c.DefaultRepo, "Repository (owner/name)")
		repoLong := fs.String("repo", c.DefaultRepo, "Repository (owner/name)")
		title := fs.String("title", "", "Pull request title")
		head := fs.String("head", "", "Head branch")
		base := fs.String("base", "main", "Base branch")
		body := fs.String("body", "", "Pull request description")
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		repoStr := *repoFlag
		if *repoLong != "" && *repoLong != c.DefaultRepo {
			repoStr = *repoLong
		}
		owner, name, err := splitRepo(repoStr)
		if err != nil {
			return err
		}
		if *title == "" || *head == "" {
			return errors.New("flags --title and --head are required")
		}

		reqBody := map[string]any{
			"title":       *title,
			"head":        *head,
			"base":        *base,
			"description": *body,
		}
		path := fmt.Sprintf("/repos/%s/%s/pulls", owner, name)
		data, code, err := c.doRequest(http.MethodPost, path, reqBody)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		fmt.Fprintln(c.Stdout, string(data))
		return nil

	case "merge":
		fs := flag.NewFlagSet("pr merge", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		repoFlag := fs.String("R", c.DefaultRepo, "Repository (owner/name)")
		repoLong := fs.String("repo", c.DefaultRepo, "Repository (owner/name)")
		method := fs.String("method", "merge", "Merge method (merge, squash)")
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		if fs.NArg() < 1 {
			return errors.New("usage: origin pr merge <number> [-R owner/name]")
		}
		num := fs.Arg(0)
		repoStr := *repoFlag
		if *repoLong != "" && *repoLong != c.DefaultRepo {
			repoStr = *repoLong
		}
		owner, name, err := splitRepo(repoStr)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/repos/%s/%s/pulls/%s/merge", owner, name, num)
		reqBody := map[string]any{
			"mergeMethod": *method,
		}
		data, code, err := c.doRequest(http.MethodPost, path, reqBody)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		fmt.Fprintln(c.Stdout, string(data))
		return nil

	default:
		return fmt.Errorf("unknown pr subcommand %q", args[0])
	}
}

// App commands
func (c *CLI) runApp(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: origin app <list|create>")
	}
	switch args[0] {
	case "list":
		fs := flag.NewFlagSet("app list", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		ns := fs.String("namespace", "", "Namespace slug")
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		if *ns == "" {
			return errors.New("flag --namespace is required")
		}
		data, code, err := c.doRequest(http.MethodGet, fmt.Sprintf("/namespaces/%s/apps", *ns), nil)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		fmt.Fprintln(c.Stdout, string(data))
		return nil

	case "create":
		fs := flag.NewFlagSet("app create", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		ns := fs.String("namespace", "", "Namespace slug")
		name := fs.String("name", "", "App display name")
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		if *ns == "" || *name == "" {
			return errors.New("flags --namespace and --name are required")
		}
		reqBody := map[string]any{
			"app": map[string]any{
				"displayName": *name,
			},
		}
		data, code, err := c.doRequest(http.MethodPost, fmt.Sprintf("/namespaces/%s/apps", *ns), reqBody)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		fmt.Fprintln(c.Stdout, string(data))
		return nil

	default:
		return fmt.Errorf("unknown app subcommand %q", args[0])
	}
}

// Arbitrary REST API call (origin api <path>)
func (c *CLI) runAPI(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: origin api <path> [-X METHOD] [-d DATA]")
	}
	fs := flag.NewFlagSet("api", flag.ContinueOnError)
	fs.SetOutput(c.Stderr)
	method := fs.String("X", "GET", "HTTP method")
	dataFlag := fs.String("d", "", "JSON request body")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return errors.New("endpoint path required")
	}
	path := fs.Arg(0)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	var body any
	if *dataFlag != "" {
		var parsed any
		if err := json.Unmarshal([]byte(*dataFlag), &parsed); err == nil {
			body = parsed
		} else {
			body = *dataFlag
		}
	}

	res, code, err := c.doRequest(strings.ToUpper(*method), path, body)
	if err != nil {
		return err
	}
	if code >= 400 {
		return fmt.Errorf("server error (%d): %s", code, string(res))
	}
	fmt.Fprintln(c.Stdout, string(res))
	return nil
}

func splitRepo(repo string) (string, string, error) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid repository target %q: expected owner/repo", repo)
	}
	return parts[0], parts[1], nil
}

func reorderFlags(args []string) []string {
	var flags, pos []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			flags = append(flags, arg)
			if !strings.Contains(arg, "=") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				flags = append(flags, args[i+1])
				i++
			}
		} else {
			pos = append(pos, arg)
		}
	}
	return append(flags, pos...)
}

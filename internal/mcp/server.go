package mcp

import (
	"encoding/json"
	"fmt"
	"net/http"

	"exist.io/internal/gitstore"
	"exist.io/internal/store"
)

// Server implements the Model Context Protocol (MCP) HTTP JSON-RPC 2.0 server.
type Server struct {
	store    *store.DB
	git      *gitstore.Store
	allTools []ToolDef
	toolMap  map[string]ToolDef
}

// New creates a new MCP Server.
func New(db *store.DB, git *gitstore.Store) *Server {
	tools := BuiltinTools()
	tmap := make(map[string]ToolDef, len(tools))
	for _, t := range tools {
		tmap[t.Name] = t
	}
	return &Server{
		store:    db,
		git:      git,
		allTools: tools,
		toolMap:  tmap,
	}
}

// JSON-RPC request and response envelopes
type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      any       `json:"id,omitempty"`
	Result  any       `json:"result,omitempty"`
	Error   *rpcError `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type toolCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type toolCallContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type toolCallResult struct {
	Content []toolCallContent `json:"content"`
	IsError bool              `json:"isError"`
}

// Handler returns an HTTP handler for the MCP endpoint.
// When readonly is true, only read-only tools are listed and allowed.
func (s *Server) Handler(readonly bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed: MCP expects POST", http.StatusMethodNotAllowed)
			return
		}

		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeRPCError(w, nil, -32700, "Parse error: "+err.Error())
			return
		}

		ctx := r.Context()
		switch req.Method {
		case "initialize":
			writeRPCResult(w, req.ID, map[string]any{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]any{
					"tools": map[string]any{},
				},
				"serverInfo": map[string]any{
					"name":    "exist.io",
					"version": "1.0.0",
				},
			})

		case "notifications/initialized":
			// MCP initialized notification; no response needed for notifications
			w.WriteHeader(http.StatusNoContent)

		case "ping":
			writeRPCResult(w, req.ID, map[string]any{})

		case "tools/list":
			var exposed []ToolDef
			for _, t := range s.allTools {
				if readonly && !t.ReadOnly {
					continue
				}
				exposed = append(exposed, t)
			}
			writeRPCResult(w, req.ID, map[string]any{
				"tools": exposed,
			})

		case "tools/call":
			var params toolCallParams
			if len(req.Params) > 0 {
				if err := json.Unmarshal(req.Params, &params); err != nil {
					writeRPCError(w, req.ID, -32602, "Invalid params: "+err.Error())
					return
				}
			}

			tool, ok := s.toolMap[params.Name]
			if !ok {
				writeRPCResult(w, req.ID, toolCallResult{
					Content: []toolCallContent{{Type: "text", Text: fmt.Sprintf("Unknown tool %q", params.Name)}},
					IsError: true,
				})
				return
			}

			if readonly && !tool.ReadOnly {
				writeRPCResult(w, req.ID, toolCallResult{
					Content: []toolCallContent{{Type: "text", Text: fmt.Sprintf("Tool %q is a write operation not permitted on /mcp/readonly", params.Name)}},
					IsError: true,
				})
				return
			}

			if params.Arguments == nil {
				params.Arguments = make(map[string]any)
			}

			res, err := tool.Handler(ctx, s, params.Arguments)
			if err != nil {
				writeRPCResult(w, req.ID, toolCallResult{
					Content: []toolCallContent{{Type: "text", Text: err.Error()}},
					IsError: true,
				})
				return
			}

			resBytes, _ := json.MarshalIndent(res, "", "  ")
			writeRPCResult(w, req.ID, toolCallResult{
				Content: []toolCallContent{{Type: "text", Text: string(resBytes)}},
				IsError: false,
			})

		default:
			writeRPCError(w, req.ID, -32601, fmt.Sprintf("Method %q not found", req.Method))
		}
	})
}

func writeRPCResult(w http.ResponseWriter, id any, result any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(rpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	})
}

func writeRPCError(w http.ResponseWriter, id any, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(rpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &rpcError{
			Code:    code,
			Message: message,
		},
	})
}

// Package symbolsearch provides pure-Go semantic code symbol indexing and search
// across Git repository trees without external indexers or runtime dependencies.
package symbolsearch

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"exist.io/internal/gitstore"
)

// Symbol represents an extracted code symbol (function, struct, class, interface, type).
type Symbol struct {
	Name      string `json:"name"`
	Kind      string `json:"kind"` // function | method | struct | interface | type | class | constant
	File      string `json:"file"`
	Line      int    `json:"line"`
	Signature string `json:"signature,omitempty"`
	Container string `json:"container,omitempty"` // Parent package, class, or struct
	Doc       string `json:"doc,omitempty"`
}

// Regex patterns for TypeScript/JavaScript symbol extraction
var (
	reTSFunc      = regexp.MustCompile(`(?m)^(?:export\s+)?(?:default\s+)?(?:async\s+)?function\s+([A-Za-z0-9_$]+)\s*\(([^)]*)\)`)
	reTSArrow     = regexp.MustCompile(`(?m)^(?:export\s+)?const\s+([A-Za-z0-9_$]+)\s*=\s*(?:async\s*)?\(([^)]*)\)\s*(?::\s*[^=]+)?=>`)
	reTSClass     = regexp.MustCompile(`(?m)^(?:export\s+)?(?:default\s+)?class\s+([A-Za-z0-9_$]+)`)
	reTSInterface = regexp.MustCompile(`(?m)^(?:export\s+)?interface\s+([A-Za-z0-9_$]+)`)
	reTSType      = regexp.MustCompile(`(?m)^(?:export\s+)?type\s+([A-Za-z0-9_$]+)\s*=`)

	// Python patterns
	rePyFunc  = regexp.MustCompile(`(?m)^([ \t]*)def\s+([A-Za-z0-9_]+)\s*\(([^)]*)\)`)
	rePyClass = regexp.MustCompile(`(?m)^class\s+([A-Za-z0-9_]+)`)
)

// ExtractGoSymbols parses Go source files using standard library go/parser and go/token AST.
func ExtractGoSymbols(filePath string, src []byte) ([]Symbol, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, src, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var symbols []Symbol
	pkgName := node.Name.Name

	for _, decl := range node.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			pos := fset.Position(d.Pos())
			kind := "function"
			container := pkgName
			sig := d.Name.Name + "()"

			if d.Recv != nil && len(d.Recv.List) > 0 {
				kind = "method"
				recvType := exprToString(d.Recv.List[0].Type)
				container = recvType
				sig = fmt.Sprintf("(%s).%s()", recvType, d.Name.Name)
			}

			doc := ""
			if d.Doc != nil {
				doc = strings.TrimSpace(d.Doc.Text())
			}

			symbols = append(symbols, Symbol{
				Name:      d.Name.Name,
				Kind:      kind,
				File:      filePath,
				Line:      pos.Line,
				Signature: sig,
				Container: container,
				Doc:       doc,
			})

		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					pos := fset.Position(s.Pos())
					kind := "type"
					switch s.Type.(type) {
					case *ast.StructType:
						kind = "struct"
					case *ast.InterfaceType:
						kind = "interface"
					}

					doc := ""
					if s.Doc != nil {
						doc = strings.TrimSpace(s.Doc.Text())
					} else if d.Doc != nil {
						doc = strings.TrimSpace(d.Doc.Text())
					}

					symbols = append(symbols, Symbol{
						Name:      s.Name.Name,
						Kind:      kind,
						File:      filePath,
						Line:      pos.Line,
						Signature: fmt.Sprintf("type %s %s", s.Name.Name, kind),
						Container: pkgName,
						Doc:       doc,
					})

				case *ast.ValueSpec:
					kind := "variable"
					if d.Tok == token.CONST {
						kind = "constant"
					}
					pos := fset.Position(s.Pos())
					for _, name := range s.Names {
						symbols = append(symbols, Symbol{
							Name:      name.Name,
							Kind:      kind,
							File:      filePath,
							Line:      pos.Line,
							Signature: name.Name,
							Container: pkgName,
						})
					}
				}
			}
		}
	}

	return symbols, nil
}

func exprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		return "*" + exprToString(e.X)
	case *ast.SelectorExpr:
		return exprToString(e.X) + "." + e.Sel.Name
	default:
		return fmt.Sprintf("%T", expr)
	}
}

// ExtractTSSymbols extracts TypeScript and JavaScript symbols.
func ExtractTSSymbols(filePath string, src []byte) []Symbol {
	var symbols []Symbol
	lines := strings.Split(string(src), "\n")

	for lineIdx, line := range lines {
		lineNum := lineIdx + 1
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
			continue
		}

		if m := reTSFunc.FindStringSubmatch(line); len(m) > 1 {
			symbols = append(symbols, Symbol{
				Name:      m[1],
				Kind:      "function",
				File:      filePath,
				Line:      lineNum,
				Signature: fmt.Sprintf("function %s(%s)", m[1], strings.TrimSpace(m[2])),
			})
			continue
		}

		if m := reTSArrow.FindStringSubmatch(line); len(m) > 1 {
			symbols = append(symbols, Symbol{
				Name:      m[1],
				Kind:      "function",
				File:      filePath,
				Line:      lineNum,
				Signature: fmt.Sprintf("const %s = (%s)", m[1], strings.TrimSpace(m[2])),
			})
			continue
		}

		if m := reTSClass.FindStringSubmatch(line); len(m) > 1 {
			symbols = append(symbols, Symbol{
				Name:      m[1],
				Kind:      "class",
				File:      filePath,
				Line:      lineNum,
				Signature: fmt.Sprintf("class %s", m[1]),
			})
			continue
		}

		if m := reTSInterface.FindStringSubmatch(line); len(m) > 1 {
			symbols = append(symbols, Symbol{
				Name:      m[1],
				Kind:      "interface",
				File:      filePath,
				Line:      lineNum,
				Signature: fmt.Sprintf("interface %s", m[1]),
			})
			continue
		}

		if m := reTSType.FindStringSubmatch(line); len(m) > 1 {
			symbols = append(symbols, Symbol{
				Name:      m[1],
				Kind:      "type",
				File:      filePath,
				Line:      lineNum,
				Signature: fmt.Sprintf("type %s", m[1]),
			})
			continue
		}
	}

	return symbols
}

// ExtractPythonSymbols extracts Python symbols.
func ExtractPythonSymbols(filePath string, src []byte) []Symbol {
	var symbols []Symbol
	lines := strings.Split(string(src), "\n")

	for lineIdx, line := range lines {
		lineNum := lineIdx + 1

		if m := rePyClass.FindStringSubmatch(line); len(m) > 1 {
			symbols = append(symbols, Symbol{
				Name:      m[1],
				Kind:      "class",
				File:      filePath,
				Line:      lineNum,
				Signature: fmt.Sprintf("class %s", m[1]),
			})
			continue
		}

		if m := rePyFunc.FindStringSubmatch(line); len(m) > 2 {
			kind := "function"
			if len(m[1]) > 0 {
				kind = "method"
			}
			symbols = append(symbols, Symbol{
				Name:      m[2],
				Kind:      kind,
				File:      filePath,
				Line:      lineNum,
				Signature: fmt.Sprintf("def %s(%s)", m[2], strings.TrimSpace(m[3])),
			})
			continue
		}
	}

	return symbols
}

// Search searches code symbols across the repository tree matching the query and kind filter.
func Search(ctx context.Context, gs *gitstore.Store, ownerSlug, repoName, ref, query, kindFilter string, limit int) ([]Symbol, error) {
	if limit <= 0 {
		limit = 50
	}

	// 1. Resolve ref
	targetHash, err := gs.ResolveRef(ownerSlug, repoName, ref)
	if err != nil {
		return nil, fmt.Errorf("symbolsearch: resolve ref %q: %w", ref, err)
	}

	// 2. Read tree recursively
	_, entries, err := gs.ReadTree(ownerSlug, repoName, targetHash.String(), "", true)
	if err != nil {
		return nil, fmt.Errorf("symbolsearch: read tree: %w", err)
	}

	var allSymbols []Symbol
	qLower := strings.ToLower(query)
	kLower := strings.ToLower(kindFilter)

	for _, entry := range entries {
		if entry.Type != "blob" {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Path))
		switch ext {
		case ".go":
			data, _, err := gs.ReadBlob(ownerSlug, repoName, entry.SHA)
			if err != nil {
				continue
			}
			syms, _ := ExtractGoSymbols(entry.Path, data)
			allSymbols = append(allSymbols, syms...)

		case ".ts", ".tsx", ".js", ".jsx":
			data, _, err := gs.ReadBlob(ownerSlug, repoName, entry.SHA)
			if err != nil {
				continue
			}
			syms := ExtractTSSymbols(entry.Path, data)
			allSymbols = append(allSymbols, syms...)

		case ".py":
			data, _, err := gs.ReadBlob(ownerSlug, repoName, entry.SHA)
			if err != nil {
				continue
			}
			syms := ExtractPythonSymbols(entry.Path, data)
			allSymbols = append(allSymbols, syms...)
		}
	}

	// 3. Filter symbols
	var filtered []Symbol
	for _, sym := range allSymbols {
		if kLower != "" && strings.ToLower(sym.Kind) != kLower {
			continue
		}
		if qLower != "" {
			nameMatch := strings.Contains(strings.ToLower(sym.Name), qLower)
			sigMatch := strings.Contains(strings.ToLower(sym.Signature), qLower)
			if !nameMatch && !sigMatch {
				continue
			}
		}
		filtered = append(filtered, sym)
	}

	// 4. Rank and sort: exact match > prefix match > substring match
	if qLower != "" {
		sort.Slice(filtered, func(i, j int) bool {
			ni := strings.ToLower(filtered[i].Name)
			nj := strings.ToLower(filtered[j].Name)

			if ni == qLower && nj != qLower {
				return true
			}
			if nj == qLower && ni != qLower {
				return false
			}

			iPref := strings.HasPrefix(ni, qLower)
			jPref := strings.HasPrefix(nj, qLower)
			if iPref && !jPref {
				return true
			}
			if jPref && !iPref {
				return false
			}

			return filtered[i].File < filtered[j].File
		})
	}

	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	return filtered, nil
}

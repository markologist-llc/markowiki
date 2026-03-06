package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

//go:embed index.html
var indexHTML []byte

var version = "1.0.0"
var appName = "MarkoWiki"

// ── Data types ────────────────────────────────────────────────────────────────

type Vault struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

type Settings struct {
	Autosave      bool   `json:"autosave"`
	AutosaveDelay int    `json:"autosave_delay"` // seconds
	Dark          bool   `json:"dark"`
	Accent        string `json:"accent"`
}

type Block struct {
	ID      string              `json:"id"`
	Title   string              `json:"title"`
	Schema  string              `json:"schema"`
	Recents map[string][]string `json:"recents,omitempty"`
}

type Config struct {
	Vaults   []Vault  `json:"vaults"`
	Settings Settings `json:"settings"`
	Blocks   []Block  `json:"blocks"`
}

type TreeNode struct {
	Name     string      `json:"name"`
	Type     string      `json:"type"`
	Path     string      `json:"path"`
	Children []*TreeNode `json:"children,omitempty"`
}

type SearchResult struct {
	VaultID   string `json:"vault_id"`
	VaultName string `json:"vault_name"`
	Path      string `json:"path"`
	Name      string `json:"name"`
	Snippet   string `json:"snippet"`
}

// ── Config file ───────────────────────────────────────────────────────────────

func configPath() string {
	if d := os.Getenv("MW_DATA_DIR"); d != "" {
		return filepath.Join(d, "config.json")
	}
	exe, _ := os.Executable()
	return filepath.Join(filepath.Dir(exe), "config.json")
}

func defaultConfig() Config {
	return Config{
		Vaults: []Vault{},
		Blocks: []Block{},
		Settings: Settings{
			Autosave:      false,
			AutosaveDelay: 2,
			Dark:          true,
			Accent:        "forest",
		},
	}
}

func loadConfig() (Config, error) {
	data, err := os.ReadFile(configPath())
	if os.IsNotExist(err) {
		return defaultConfig(), nil
	}
	if err != nil {
		return defaultConfig(), err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return defaultConfig(), err
	}
	// Fill in missing settings with defaults
	if cfg.Settings.Accent == "" {
	cfg.Settings.Accent = "forest"
	}
	if cfg.Settings.AutosaveDelay == 0 {
		cfg.Settings.AutosaveDelay = 2
	}
	return cfg, nil
}

func saveConfig(cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath(), data, 0644)
}

func getVault(id string) (*Vault, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}
	for _, v := range cfg.Vaults {
		if v.ID == id {
			return &v, nil
		}
	}
	return nil, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"detail": msg})
}

func safePath(vaultPath, relPath string) (string, bool) {
	full := filepath.Clean(filepath.Join(vaultPath, relPath))
	if !strings.HasPrefix(full, filepath.Clean(vaultPath)+string(os.PathSeparator)) &&
		full != filepath.Clean(vaultPath) {
		return "", false
	}
	return full, true
}

func slugify(s string) string {
	re := regexp.MustCompile(`[^a-z0-9_]`)
	return re.ReplaceAllString(strings.ToLower(s), "_")
}

// ── File tree ─────────────────────────────────────────────────────────────────

func buildTree(base, current string) ([]*TreeNode, error) {
	entries, err := os.ReadDir(current)
	if err != nil {
		return nil, err
	}
	var nodes []*TreeNode
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if e.IsDir() {
			rel, _ := filepath.Rel(base, filepath.Join(current, e.Name()))
			children, _ := buildTree(base, filepath.Join(current, e.Name()))
			nodes = append(nodes, &TreeNode{
				Name:     e.Name(),
				Type:     "folder",
				Path:     rel,
				Children: children,
			})
		}
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if !e.IsDir() {
			ext := strings.ToLower(filepath.Ext(e.Name()))
			if ext == ".md" || ext == ".markdown" || ext == ".txt" {
				rel, _ := filepath.Rel(base, filepath.Join(current, e.Name()))
				nodes = append(nodes, &TreeNode{
					Name: e.Name(),
					Type: "file",
					Path: rel,
				})
			}
		}
	}
	return nodes, nil
}

// ── Handlers ──────────────────────────────────────────────────────────────────

// GET/POST /api/vaults
func handleVaults(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodGet:
		cfg, err := loadConfig()
		if err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, cfg.Vaults)

	case http.MethodPost:
		var body struct {
			Name string `json:"name"`
			Path string `json:"path"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, 400, "invalid JSON")
			return
		}
		body.Name = strings.TrimSpace(body.Name)
		body.Path = strings.TrimSpace(body.Path)
		if body.Name == "" || body.Path == "" {
			writeError(w, 400, "name and path required")
			return
		}
		abs, err := filepath.Abs(body.Path)
		if err != nil {
			writeError(w, 400, "invalid path")
			return
		}
		info, err := os.Stat(abs)
		if err != nil || !info.IsDir() {
			writeError(w, 400, "path does not exist or is not a directory: "+body.Path)
			return
		}
		cfg, _ := loadConfig()
		id := fmt.Sprintf("%s_%d", slugify(body.Name), len(cfg.Vaults))
		vault := Vault{ID: id, Name: body.Name, Path: abs}
		cfg.Vaults = append(cfg.Vaults, vault)
		if err := saveConfig(cfg); err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, vault)

	default:
		writeError(w, 405, "method not allowed")
	}
}

// DELETE /api/vaults/{id}
func handleVaultDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, 405, "method not allowed")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/vaults/")
	cfg, _ := loadConfig()
	filtered := cfg.Vaults[:0]
	for _, v := range cfg.Vaults {
		if v.ID != id {
			filtered = append(filtered, v)
		}
	}
	cfg.Vaults = filtered
	saveConfig(cfg)
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// GET /api/vaults/{id}/tree
func handleVaultTree(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		writeError(w, 400, "bad url")
		return
	}
	id := parts[3]
	vault, _ := getVault(id)
	if vault == nil {
		writeError(w, 404, "vault not found")
		return
	}
	tree, err := buildTree(vault.Path, vault.Path)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	if tree == nil {
		tree = []*TreeNode{}
	}
	writeJSON(w, 200, tree)
}

// GET/PUT/DELETE /api/file
func handleFile(w http.ResponseWriter, r *http.Request) {
	vaultID := r.URL.Query().Get("vault_id")
	relPath := r.URL.Query().Get("path")
	vault, _ := getVault(vaultID)
	if vault == nil {
		writeError(w, 404, "vault not found")
		return
	}
	full, ok := safePath(vault.Path, relPath)
	if !ok {
		writeError(w, 403, "access denied")
		return
	}
	switch r.Method {
	case http.MethodGet:
		data, err := os.ReadFile(full)
		if err != nil {
			writeError(w, 404, "file not found")
			return
		}
		writeJSON(w, 200, map[string]string{"content": string(data), "path": relPath})
	case http.MethodPut:
		var body struct {
			Content string `json:"content"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		os.MkdirAll(filepath.Dir(full), 0755)
		if err := os.WriteFile(full, []byte(body.Content), 0644); err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, map[string]bool{"ok": true})
	case http.MethodDelete:
		if err := os.Remove(full); err != nil {
			writeError(w, 404, "file not found")
			return
		}
		writeJSON(w, 200, map[string]bool{"ok": true})
	default:
		writeError(w, 405, "method not allowed")
	}
}

// POST /api/file/create
func handleFileCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, 405, "method not allowed")
		return
	}
	vaultID := r.URL.Query().Get("vault_id")
	vault, _ := getVault(vaultID)
	if vault == nil {
		writeError(w, 404, "vault not found")
		return
	}
	var body struct {
		Path string `json:"path"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	rel := strings.Trim(body.Path, "/")
	if !strings.HasSuffix(rel, ".md") {
		rel += ".md"
	}
	full, ok := safePath(vault.Path, rel)
	if !ok {
		writeError(w, 403, "access denied")
		return
	}
	if _, err := os.Stat(full); err == nil {
		writeError(w, 409, "file already exists")
		return
	}
	os.MkdirAll(filepath.Dir(full), 0755)
	title := strings.TrimSuffix(filepath.Base(rel), ".md")
	if err := os.WriteFile(full, []byte(fmt.Sprintf("# %s\n\n", title)), 0644); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"path": rel})
}

// GET /api/search
func handleSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(r.URL.Query().Get("q"))
	vaultID := r.URL.Query().Get("vault_id")
	if q == "" {
		writeJSON(w, 200, []SearchResult{})
		return
	}
	cfg, _ := loadConfig()
	var results []SearchResult
	for _, vault := range cfg.Vaults {
		if vaultID != "" && vault.ID != vaultID {
			continue
		}
		filepath.WalkDir(vault.Path, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if ext != ".md" && ext != ".markdown" {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			content := string(data)
			rel, _ := filepath.Rel(vault.Path, path)
			if strings.Contains(strings.ToLower(content), q) ||
				strings.Contains(strings.ToLower(d.Name()), q) {
				snippet := ""
				idx := strings.Index(strings.ToLower(content), q)
				if idx >= 0 {
					start := idx - 60
					if start < 0 {
						start = 0
					}
					end := idx + 120
					if end > len(content) {
						end = len(content)
					}
					snippet = strings.ReplaceAll(content[start:end], "\n", " ")
				}
				results = append(results, SearchResult{
					VaultID:   vault.ID,
					VaultName: vault.Name,
					Path:      rel,
					Name:      d.Name(),
					Snippet:   snippet,
				})
			}
			return nil
		})
	}
	if results == nil {
		results = []SearchResult{}
	}
	writeJSON(w, 200, results)
}

// ── Blocks ────────────────────────────────────────────────────────────────────

func extractPlaceholders(schema string) []string {
	re := regexp.MustCompile(`\{([^}]+)\}`)
	matches := re.FindAllStringSubmatch(schema, -1)
	seen := map[string]bool{}
	var out []string
	for _, m := range matches {
		if !seen[m[1]] {
			seen[m[1]] = true
			out = append(out, m[1])
		}
	}
	return out
}

// GET/POST /api/blocks
func handleBlocks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg, err := loadConfig()
		if err != nil {
			writeError(w, 500, err.Error())
			return
		}
		if cfg.Blocks == nil {
			cfg.Blocks = []Block{}
		}
		writeJSON(w, 200, cfg.Blocks)

	case http.MethodPost:
		var b Block
		if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
			writeError(w, 400, "invalid JSON")
			return
		}
		b.Title = strings.TrimSpace(b.Title)
		b.Schema = strings.TrimSpace(b.Schema)
		if b.Title == "" || b.Schema == "" {
			writeError(w, 400, "title and schema required")
			return
		}
		cfg, _ := loadConfig()
		b.ID = fmt.Sprintf("blk_%s_%d", slugify(b.Title), len(cfg.Blocks))
		if b.Recents == nil {
			b.Recents = map[string][]string{}
		}
		cfg.Blocks = append(cfg.Blocks, b)
		if err := saveConfig(cfg); err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, b)

	default:
		writeError(w, 405, "method not allowed")
	}
}

// PUT/DELETE /api/blocks/{id}
func handleBlockByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/blocks/")

	switch r.Method {
	case http.MethodPut:
		var b Block
		if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
			writeError(w, 400, "invalid JSON")
			return
		}
		cfg, _ := loadConfig()
		for i, blk := range cfg.Blocks {
			if blk.ID == id {
				b.ID = id
				if b.Recents == nil {
					b.Recents = blk.Recents // preserve recents on edit
				}
				cfg.Blocks[i] = b
				saveConfig(cfg)
				writeJSON(w, 200, b)
				return
			}
		}
		writeError(w, 404, "block not found")

	case http.MethodDelete:
		cfg, _ := loadConfig()
		filtered := cfg.Blocks[:0]
		for _, blk := range cfg.Blocks {
			if blk.ID != id {
				filtered = append(filtered, blk)
			}
		}
		cfg.Blocks = filtered
		saveConfig(cfg)
		writeJSON(w, 200, map[string]bool{"ok": true})

	default:
		writeError(w, 405, "method not allowed")
	}
}

// POST /api/blocks/{id}/recents  — update recent values for placeholders
func handleBlockRecents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, 405, "method not allowed")
		return
	}
	// path: /api/blocks/{id}/recents
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		writeError(w, 400, "bad url")
		return
	}
	id := parts[3]
	var values map[string]string // placeholder -> value used
	if err := json.NewDecoder(r.Body).Decode(&values); err != nil {
		writeError(w, 400, "invalid JSON")
		return
	}
	cfg, _ := loadConfig()
	for i, blk := range cfg.Blocks {
		if blk.ID == id {
			if blk.Recents == nil {
				blk.Recents = map[string][]string{}
			}
			for k, v := range values {
				if v == "" {
					continue
				}
				list := blk.Recents[k]
				// Remove if already exists (move to front)
				newList := []string{v}
				for _, existing := range list {
					if existing != v {
						newList = append(newList, existing)
					}
				}
				if len(newList) > 3 {
					newList = newList[:3]
				}
				blk.Recents[k] = newList
			}
			cfg.Blocks[i] = blk
			saveConfig(cfg)
			writeJSON(w, 200, blk.Recents)
			return
		}
	}
	writeError(w, 404, "block not found")
}

// GET/PUT /api/settings
func handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodGet:
		cfg, err := loadConfig()
		if err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, cfg.Settings)

	case http.MethodPut:
		var s Settings
		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			writeError(w, 400, "invalid JSON")
			return
		}
		cfg, _ := loadConfig()
		cfg.Settings = s
		if err := saveConfig(cfg); err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, cfg.Settings)

	default:
		writeError(w, 405, "method not allowed")
	}
}

// GET / — serve embedded index.html
func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(indexHTML)
}

// ── Router ────────────────────────────────────────────────────────────────────

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Printf("%s v%s\n", appName, version)
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/vaults", handleVaults)
	mux.HandleFunc("/api/vaults/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/tree") {
			handleVaultTree(w, r)
		} else {
			handleVaultDelete(w, r)
		}
	})
	mux.HandleFunc("/api/blocks", handleBlocks)
	mux.HandleFunc("/api/blocks/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/recents") {
			handleBlockRecents(w, r)
		} else {
			handleBlockByID(w, r)
		}
	})
	mux.HandleFunc("/api/file/create", handleFileCreate)
	mux.HandleFunc("/api/file", handleFile)
	mux.HandleFunc("/api/search", handleSearch)
	mux.HandleFunc("/api/settings", handleSettings)
	mux.HandleFunc("/", handleIndex)

	port := "8000"
	fmt.Printf("\n  %s v%s\n", appName, version)
	fmt.Printf("  → http://localhost:%s\n\n", port)

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

//go:embed chat/*
var chatFS embed.FS

const (
	defaultPort       = "8080"
	defaultTimeout    = 60 * time.Second
	maxQueryLength    = 1000
	defaultSandboxDir = "/workspace/scion"
	defaultSystemMD   = "/etc/docs-agent/system-prompt.md"
	defaultModel      = "gemini-3.1-flash-lite-preview"
)

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// runCommand executes a command and returns its combined stdout and stderr.
// Override in tests to mock command execution.
var runCommand = func(ctx context.Context, name string, args []string, env []string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}
	out, err := cmd.Output()
	return string(out), err
}

type askRequest struct {
	Query string `json:"query"`
}

type askResponse struct {
	Answer string `json:"answer"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type healthResponse struct {
	Status string `json:"status"`
}

type refreshResponse struct {
	Status string `json:"status"`
	Output string `json:"output,omitempty"`
}

func getTimeout() time.Duration {
	if v := os.Getenv("DOCS_AGENT_TIMEOUT"); v != "" {
		if secs, err := strconv.Atoi(v); err == nil && secs > 0 {
			return time.Duration(secs) * time.Second
		}
	}
	return defaultTimeout
}

func getSandboxDir() string {
	if v := os.Getenv("DOCS_AGENT_SANDBOX_DIR"); v != "" {
		return v
	}
	return defaultSandboxDir
}

func getSystemMD() string {
	if v := os.Getenv("GEMINI_SYSTEM_MD"); v != "" {
		return v
	}
	return defaultSystemMD
}

func getModel() string {
	if v := os.Getenv("DOCS_AGENT_MODEL"); v != "" {
		return v
	}
	return defaultModel
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

func handleAsk(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 2048))
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read request body")
		return
	}

	var req askRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	query := strings.TrimSpace(req.Query)
	if query == "" {
		writeError(w, http.StatusBadRequest, "query must not be empty")
		return
	}
	if len(query) > maxQueryLength {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("query must not exceed %d characters", maxQueryLength))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), getTimeout())
	defer cancel()

	args := []string{
		"--prompt", query,
		"--model", getModel(),
		"--sandbox_dir", getSandboxDir(),
	}
	env := []string{
		"GEMINI_SYSTEM_MD=" + getSystemMD(),
	}

	output, err := runCommand(ctx, "gemini", args, env)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			writeError(w, http.StatusGatewayTimeout, "request timed out")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get response from Gemini")
		return
	}

	answer := ansiRegexp.ReplaceAllString(output, "")
	writeJSON(w, http.StatusOK, askResponse{Answer: answer})
}

func handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	data, err := chatFS.ReadFile("chat/index.html")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "chat widget not found")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}

func handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	sandboxDir := getSandboxDir()
	output, err := runCommand(ctx, "git", []string{"-C", sandboxDir, "pull", "--ff-only"}, nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "git pull failed: "+output)
		return
	}

	writeJSON(w, http.StatusOK, refreshResponse{Status: "ok", Output: strings.TrimSpace(output)})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}

// corsMiddleware wraps a handler to add CORS headers for cross-origin access.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/ask", handleAsk)
	mux.HandleFunc("/chat", handleChat)
	mux.HandleFunc("/refresh", handleRefresh)
	mux.HandleFunc("/health", handleHealth)

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	handler := corsMiddleware(mux)

	log.Printf("docs-agent listening on :%s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

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

package harness

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GoogleCloudPlatform/scion/pkg/api"
)

func TestPiInjectAgentInstructions(t *testing.T) {
	agentHome := t.TempDir()
	p := &Pi{}
	content := []byte("# Agent Instructions\nDo good work.")

	if err := p.InjectAgentInstructions(agentHome, content); err != nil {
		t.Fatalf("InjectAgentInstructions failed: %v", err)
	}

	target := filepath.Join(agentHome, ".pi", "agent", "AGENTS.md")
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("expected file at %s: %v", target, err)
	}
	if string(data) != string(content) {
		t.Errorf("content mismatch: got %q, want %q", string(data), string(content))
	}
}

func TestPiInjectSystemPrompt(t *testing.T) {
	agentHome := t.TempDir()
	p := &Pi{}
	content := []byte("You are a helpful assistant.")

	if err := p.InjectSystemPrompt(agentHome, content); err != nil {
		t.Fatalf("InjectSystemPrompt failed: %v", err)
	}

	target := filepath.Join(agentHome, ".pi", "agent", "SYSTEM.md")
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("expected file at %s: %v", target, err)
	}
	if string(data) != string(content) {
		t.Errorf("content mismatch: got %q, want %q", string(data), string(content))
	}
}

func TestPiResolveAuth_AnthropicAPIKey(t *testing.T) {
	p := &Pi{}
	auth := api.AuthConfig{AnthropicAPIKey: "sk-ant-test"}
	result, err := p.ResolveAuth(auth)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Method != "api-key" {
		t.Errorf("Method = %q, want %q", result.Method, "api-key")
	}
	if result.EnvVars["ANTHROPIC_API_KEY"] != "sk-ant-test" {
		t.Errorf("ANTHROPIC_API_KEY = %q, want %q", result.EnvVars["ANTHROPIC_API_KEY"], "sk-ant-test")
	}
}

func TestPiResolveAuth_OpenAIAPIKey(t *testing.T) {
	p := &Pi{}
	auth := api.AuthConfig{OpenAIAPIKey: "sk-openai-test"}
	result, err := p.ResolveAuth(auth)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Method != "api-key" {
		t.Errorf("Method = %q, want %q", result.Method, "api-key")
	}
	if result.EnvVars["OPENAI_API_KEY"] != "sk-openai-test" {
		t.Errorf("OPENAI_API_KEY = %q, want %q", result.EnvVars["OPENAI_API_KEY"], "sk-openai-test")
	}
}

func TestPiResolveAuth_AuthFile(t *testing.T) {
	p := &Pi{}
	auth := api.AuthConfig{PiAuthFile: "/home/user/.pi/agent/auth.json"}
	result, err := p.ResolveAuth(auth)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Method != "auth-file" {
		t.Errorf("Method = %q, want %q", result.Method, "auth-file")
	}
	if len(result.Files) != 1 {
		t.Fatalf("expected 1 file mapping, got %d", len(result.Files))
	}
}

func TestPiResolveAuth_PreferenceOrder(t *testing.T) {
	p := &Pi{}
	// AnthropicAPIKey should win over OpenAIAPIKey and auth file
	auth := api.AuthConfig{
		AnthropicAPIKey: "anthropic",
		OpenAIAPIKey:    "openai",
		PiAuthFile:      "/auth.json",
	}
	result, err := p.ResolveAuth(auth)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Method != "api-key" {
		t.Errorf("AnthropicAPIKey should win; Method = %q, want %q", result.Method, "api-key")
	}

	// OpenAIAPIKey should win over auth file
	auth = api.AuthConfig{
		OpenAIAPIKey: "openai",
		PiAuthFile:   "/auth.json",
	}
	result, err = p.ResolveAuth(auth)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Method != "api-key" {
		t.Errorf("OpenAIAPIKey should win over auth file; Method = %q, want %q", result.Method, "api-key")
	}
}

func TestPiResolveAuth_NoCreds(t *testing.T) {
	p := &Pi{}
	result, err := p.ResolveAuth(api.AuthConfig{})
	if err != nil {
		t.Fatalf("expected no error for empty AuthConfig (local models supported): %v", err)
	}
	if result.Method != "none" {
		t.Errorf("expected method %q for no-creds case, got %q", "none", result.Method)
	}
}

func TestPiGetCommand_Basic(t *testing.T) {
	p := &Pi{}
	cmd := p.GetCommand("do the thing", false, nil)
	if len(cmd) < 3 {
		t.Fatalf("expected at least 3 args, got %v", cmd)
	}
	if cmd[0] != "pi" {
		t.Errorf("cmd[0] = %q, want %q", cmd[0], "pi")
	}
	if cmd[1] != "--print" {
		t.Errorf("cmd[1] = %q, want %q", cmd[1], "--print")
	}
	if cmd[2] != "do the thing" {
		t.Errorf("cmd[2] = %q, want %q", cmd[2], "do the thing")
	}
}

func TestPiGetCommand_Resume(t *testing.T) {
	p := &Pi{}
	cmd := p.GetCommand("", true, nil)
	found := false
	for _, a := range cmd {
		if a == "--continue" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected --continue in command: %v", cmd)
	}
}

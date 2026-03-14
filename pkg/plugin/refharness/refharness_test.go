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

package refharness

import (
	"context"
	"embed"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/GoogleCloudPlatform/scion/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefHarness_Name(t *testing.T) {
	h := New(slog.Default())
	assert.Equal(t, "refharness", h.Name())
}

func TestRefHarness_AdvancedCapabilities(t *testing.T) {
	h := New(slog.Default())
	caps := h.AdvancedCapabilities()
	assert.Equal(t, "refharness", caps.Harness)
	assert.Equal(t, api.SupportYes, caps.Limits.MaxDuration.Support)
}

func TestRefHarness_GetEnv(t *testing.T) {
	h := New(slog.Default())
	env := h.GetEnv("agent1", "/home/agent1", "user")
	assert.Equal(t, "agent1", env["SCION_AGENT_NAME"])
	assert.Equal(t, "/home/agent1", env["REFHARNESS_HOME"])
}

func TestRefHarness_GetCommand(t *testing.T) {
	h := New(slog.Default())

	cmd := h.GetCommand("do stuff", false, []string{"--verbose"})
	assert.Equal(t, []string{"refharness-agent", "--task", "do stuff", "--verbose"}, cmd)

	cmd = h.GetCommand("", true, nil)
	assert.Equal(t, []string{"refharness-agent", "--resume"}, cmd)
}

func TestRefHarness_GetHarnessEmbedsFS_ReturnsEmpty(t *testing.T) {
	h := New(slog.Default())
	fs, base := h.GetHarnessEmbedsFS()
	assert.Equal(t, embed.FS{}, fs)
	assert.Empty(t, base)
}

func TestRefHarness_Provision(t *testing.T) {
	h := New(slog.Default())
	home := t.TempDir()

	err := h.Provision(context.Background(), "agent1", "/dir", home, "/workspace")
	require.NoError(t, err)
	assert.True(t, h.Provisioned)

	// Verify config file was written
	configPath := filepath.Join(home, ".refharness", "config.yaml")
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "agent_name: agent1")
	assert.Contains(t, string(data), "workspace: /workspace")
}

func TestRefHarness_InjectAgentInstructions(t *testing.T) {
	h := New(slog.Default())
	home := t.TempDir()

	content := []byte("# Agent Instructions\nDo the thing.")
	err := h.InjectAgentInstructions(home, content)
	require.NoError(t, err)
	assert.Equal(t, content, h.InjectedInstructions)

	// Verify file was written
	data, err := os.ReadFile(filepath.Join(home, ".refharness", "instructions.md"))
	require.NoError(t, err)
	assert.Equal(t, content, data)
}

func TestRefHarness_InjectSystemPrompt(t *testing.T) {
	h := New(slog.Default())
	home := t.TempDir()

	content := []byte("You are a helpful assistant.")
	err := h.InjectSystemPrompt(home, content)
	require.NoError(t, err)
	assert.Equal(t, content, h.InjectedSystemPrompt)

	// Verify file and HasSystemPrompt
	assert.True(t, h.HasSystemPrompt(home))
}

func TestRefHarness_HasSystemPrompt_False(t *testing.T) {
	h := New(slog.Default())
	assert.False(t, h.HasSystemPrompt(t.TempDir()))
}

func TestRefHarness_ResolveAuth(t *testing.T) {
	h := New(slog.Default())
	resolved, err := h.ResolveAuth(api.AuthConfig{
		AnthropicAPIKey: "sk-test",
		GeminiAPIKey:    "gem-test",
	})
	require.NoError(t, err)
	assert.Equal(t, "passthrough", resolved.Method)
	assert.Equal(t, "sk-test", resolved.EnvVars["ANTHROPIC_API_KEY"])
	assert.Equal(t, "gem-test", resolved.EnvVars["GEMINI_API_KEY"])
}

func TestRefHarness_ApplyAuthSettings(t *testing.T) {
	h := New(slog.Default())
	err := h.ApplyAuthSettings("/home/agent1", &api.ResolvedAuth{Method: "passthrough"})
	require.NoError(t, err)
	assert.Equal(t, "/home/agent1", h.AppliedAuthHome)
}

func TestRefHarness_ApplyTelemetrySettings(t *testing.T) {
	h := New(slog.Default())
	enabled := true
	err := h.ApplyTelemetrySettings("/home/agent1", &api.TelemetryConfig{Enabled: &enabled}, nil)
	require.NoError(t, err)
	assert.Equal(t, "/home/agent1", h.AppliedTelemetryHome)
}

func TestRefHarness_GetInfo(t *testing.T) {
	h := New(slog.Default())
	info, err := h.GetInfo()
	require.NoError(t, err)
	assert.Equal(t, "refharness", info.Name)
	assert.Equal(t, "0.1.0", info.Version)
	assert.Contains(t, info.Capabilities, "auth_settings")
	assert.Contains(t, info.Capabilities, "telemetry_settings")
}

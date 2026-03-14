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
	"net"
	"net/rpc"
	"testing"

	"github.com/GoogleCloudPlatform/scion/pkg/api"
	"github.com/GoogleCloudPlatform/scion/pkg/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// startRefHarnessRPCServer starts a real RefHarness behind an RPC server and returns a client.
// This exercises the full RPC transport path: client -> net/rpc -> HarnessRPCServer -> RefHarness.
func startRefHarnessRPCServer(t *testing.T) (*plugin.HarnessRPCClient, *RefHarness) {
	t.Helper()

	impl := New(slog.Default())

	server := rpc.NewServer()
	rpcServer := &plugin.HarnessRPCServer{Impl: impl}
	require.NoError(t, server.RegisterName("Plugin", rpcServer))

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { listener.Close() })

	go server.Accept(listener)

	client, err := rpc.Dial("tcp", listener.Addr().String())
	require.NoError(t, err)
	t.Cleanup(func() { client.Close() })

	return plugin.NewHarnessRPCClient(client), impl
}

func TestRPCIntegration_Metadata(t *testing.T) {
	client, _ := startRefHarnessRPCServer(t)

	assert.Equal(t, "refharness", client.Name())
	assert.Equal(t, ".refharness", client.DefaultConfigDir())
	assert.Equal(t, ".refharness/skills", client.SkillsDir())
	assert.Equal(t, "refharness", client.GetEmbedDir())
	assert.Equal(t, "C-c", client.GetInterruptKey())
}

func TestRPCIntegration_AdvancedCapabilities(t *testing.T) {
	client, _ := startRefHarnessRPCServer(t)

	caps := client.AdvancedCapabilities()
	assert.Equal(t, "refharness", caps.Harness)
	assert.Equal(t, api.SupportYes, caps.Limits.MaxDuration.Support)
}

func TestRPCIntegration_GetEnv(t *testing.T) {
	client, _ := startRefHarnessRPCServer(t)

	env := client.GetEnv("coder", "/home/coder", "coder")
	assert.Equal(t, "coder", env["SCION_AGENT_NAME"])
	assert.Equal(t, "/home/coder", env["REFHARNESS_HOME"])
}

func TestRPCIntegration_GetCommand(t *testing.T) {
	client, _ := startRefHarnessRPCServer(t)

	cmd := client.GetCommand("build feature", false, []string{"--debug"})
	assert.Equal(t, []string{"refharness-agent", "--task", "build feature", "--debug"}, cmd)
}

func TestRPCIntegration_GetHarnessEmbedsFS(t *testing.T) {
	client, _ := startRefHarnessRPCServer(t)

	fs, base := client.GetHarnessEmbedsFS()
	assert.Equal(t, embed.FS{}, fs)
	assert.Empty(t, base)
}

func TestRPCIntegration_Provision(t *testing.T) {
	client, impl := startRefHarnessRPCServer(t)
	home := t.TempDir()

	err := client.Provision(context.Background(), "agent1", "/dir", home, "/workspace")
	require.NoError(t, err)
	assert.True(t, impl.Provisioned)
}

func TestRPCIntegration_InjectAgentInstructions(t *testing.T) {
	client, impl := startRefHarnessRPCServer(t)
	home := t.TempDir()

	content := []byte("# Instructions\nDo the thing.")
	err := client.InjectAgentInstructions(home, content)
	require.NoError(t, err)
	assert.Equal(t, content, impl.InjectedInstructions)
}

func TestRPCIntegration_InjectSystemPrompt(t *testing.T) {
	client, impl := startRefHarnessRPCServer(t)
	home := t.TempDir()

	content := []byte("You are a helpful assistant.")
	err := client.InjectSystemPrompt(home, content)
	require.NoError(t, err)
	assert.Equal(t, content, impl.InjectedSystemPrompt)
}

func TestRPCIntegration_HasSystemPrompt(t *testing.T) {
	client, _ := startRefHarnessRPCServer(t)
	home := t.TempDir()

	// No system prompt yet
	assert.False(t, client.HasSystemPrompt(home))

	// Inject one
	err := client.InjectSystemPrompt(home, []byte("prompt"))
	require.NoError(t, err)

	// Now it should exist
	assert.True(t, client.HasSystemPrompt(home))
}

func TestRPCIntegration_ResolveAuth(t *testing.T) {
	client, _ := startRefHarnessRPCServer(t)

	resolved, err := client.ResolveAuth(api.AuthConfig{
		AnthropicAPIKey: "sk-test-key",
	})
	require.NoError(t, err)
	assert.Equal(t, "passthrough", resolved.Method)
	assert.Equal(t, "sk-test-key", resolved.EnvVars["ANTHROPIC_API_KEY"])
}

func TestRPCIntegration_ApplyAuthSettings(t *testing.T) {
	client, impl := startRefHarnessRPCServer(t)

	err := client.ApplyAuthSettings("/home/agent1", &api.ResolvedAuth{Method: "passthrough"})
	require.NoError(t, err)
	assert.Equal(t, "/home/agent1", impl.AppliedAuthHome)
}

func TestRPCIntegration_ApplyTelemetrySettings(t *testing.T) {
	client, impl := startRefHarnessRPCServer(t)

	enabled := true
	err := client.ApplyTelemetrySettings("/home/agent1", &api.TelemetryConfig{Enabled: &enabled}, map[string]string{"FOO": "bar"})
	require.NoError(t, err)
	assert.Equal(t, "/home/agent1", impl.AppliedTelemetryHome)
}

func TestRPCIntegration_GetTelemetryEnv(t *testing.T) {
	client, _ := startRefHarnessRPCServer(t)

	env := client.GetTelemetryEnv()
	assert.Equal(t, "http://localhost:4317", env["OTEL_EXPORTER_OTLP_ENDPOINT"])
}

func TestRPCIntegration_GetInfo(t *testing.T) {
	client, _ := startRefHarnessRPCServer(t)

	info, err := client.GetInfo()
	require.NoError(t, err)
	// HarnessRPCServer.GetInfo populates Name from Impl.Name(), not from a
	// separate GetInfo() method (unlike broker plugins), so Version is empty.
	assert.Equal(t, "refharness", info.Name)
}

func TestRPCIntegration_FullLifecycle(t *testing.T) {
	client, impl := startRefHarnessRPCServer(t)
	home := t.TempDir()

	// 1. GetInfo
	info, err := client.GetInfo()
	require.NoError(t, err)
	assert.Equal(t, "refharness", info.Name)

	// 2. Check metadata
	assert.Equal(t, "refharness", client.Name())
	assert.Equal(t, ".refharness", client.DefaultConfigDir())

	// 3. Provision
	err = client.Provision(context.Background(), "lifecycle-agent", "/dir", home, "/workspace")
	require.NoError(t, err)
	assert.True(t, impl.Provisioned)

	// 4. Inject instructions
	err = client.InjectAgentInstructions(home, []byte("# Instructions"))
	require.NoError(t, err)

	// 5. Inject system prompt
	err = client.InjectSystemPrompt(home, []byte("You are helpful."))
	require.NoError(t, err)
	assert.True(t, client.HasSystemPrompt(home))

	// 6. Resolve auth
	resolved, err := client.ResolveAuth(api.AuthConfig{AnthropicAPIKey: "sk-key"})
	require.NoError(t, err)
	assert.Equal(t, "passthrough", resolved.Method)

	// 7. Apply auth settings (optional capability)
	err = client.ApplyAuthSettings(home, resolved)
	require.NoError(t, err)
	assert.Equal(t, home, impl.AppliedAuthHome)

	// 8. Apply telemetry settings (optional capability)
	enabled := true
	err = client.ApplyTelemetrySettings(home, &api.TelemetryConfig{Enabled: &enabled}, nil)
	require.NoError(t, err)
	assert.Equal(t, home, impl.AppliedTelemetryHome)

	// 9. Get env and command
	env := client.GetEnv("lifecycle-agent", home, "user")
	assert.Equal(t, "lifecycle-agent", env["SCION_AGENT_NAME"])

	cmd := client.GetCommand("build it", false, nil)
	assert.Equal(t, []string{"refharness-agent", "--task", "build it"}, cmd)
}

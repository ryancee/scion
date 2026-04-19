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
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/GoogleCloudPlatform/scion/pkg/api"
	piEmbeds "github.com/GoogleCloudPlatform/scion/pkg/harness/pi"
)

type Pi struct{}

func (p *Pi) Name() string {
	return "pi"
}

func (p *Pi) AdvancedCapabilities() api.HarnessAdvancedCapabilities {
	return api.HarnessAdvancedCapabilities{
		Harness: "pi",
		Limits: api.HarnessLimitCapabilities{
			MaxTurns:      api.CapabilityField{Support: api.SupportNo, Reason: "This harness has no hook dialect for turn events"},
			MaxModelCalls: api.CapabilityField{Support: api.SupportNo, Reason: "This harness has no hook dialect for model events"},
			MaxDuration:   api.CapabilityField{Support: api.SupportYes},
		},
		Telemetry: api.HarnessTelemetryCapabilities{
			EnabledConfig: api.CapabilityField{Support: api.SupportNo, Reason: "Native telemetry config is not supported for this harness"},
			NativeEmitter: api.CapabilityField{Support: api.SupportNo, Reason: "Native telemetry forwarding is not wired for this harness"},
		},
		Prompts: api.HarnessPromptCapabilities{
			SystemPrompt:      api.CapabilityField{Support: api.SupportYes},
			AgentInstructions: api.CapabilityField{Support: api.SupportYes},
		},
		Auth: api.HarnessAuthCapabilities{
			APIKey:   api.CapabilityField{Support: api.SupportYes},
			AuthFile: api.CapabilityField{Support: api.SupportYes},
			VertexAI: api.CapabilityField{Support: api.SupportNo, Reason: "Vertex AI auth is not supported for this harness"},
		},
	}
}

func (p *Pi) GetEnv(agentName string, agentHome string, unixUsername string) map[string]string {
	return map[string]string{}
}

func (p *Pi) GetCommand(task string, resume bool, baseArgs []string) []string {
	args := []string{"pi", "--print"}
	if resume {
		args = append(args, "--continue")
	}
	if task != "" {
		args = append(args, task)
	}
	args = append(args, baseArgs...)
	return args
}

func (p *Pi) DefaultConfigDir() string {
	return ".pi/agent"
}

func (p *Pi) SkillsDir() string {
	return ".pi/agent/skills"
}

func (p *Pi) HasSystemPrompt(agentHome string) bool {
	_, err := os.Stat(filepath.Join(agentHome, ".pi", "agent", "SYSTEM.md"))
	return err == nil
}

func (p *Pi) Provision(ctx context.Context, agentName, agentDir, agentHome, agentWorkspace string) error {
	return nil
}

func (p *Pi) GetEmbedDir() string {
	return "pi"
}

func (p *Pi) GetInterruptKey() string {
	return "C-c"
}

func (p *Pi) GetHarnessEmbedsFS() (embed.FS, string) {
	return piEmbeds.EmbedsFS, "embeds"
}

func (p *Pi) GetTelemetryEnv() map[string]string {
	return nil
}

func (p *Pi) InjectAgentInstructions(agentHome string, content []byte) error {
	target := filepath.Join(agentHome, ".pi", "agent", "AGENTS.md")
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return fmt.Errorf("failed to create pi agent dir: %w", err)
	}
	return os.WriteFile(target, content, 0644)
}

func (p *Pi) InjectSystemPrompt(agentHome string, content []byte) error {
	target := filepath.Join(agentHome, ".pi", "agent", "SYSTEM.md")
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return fmt.Errorf("failed to create pi agent dir: %w", err)
	}
	return os.WriteFile(target, content, 0644)
}

func (p *Pi) ResolveAuth(auth api.AuthConfig) (*api.ResolvedAuth, error) {
	// Explicit selection support
	if auth.SelectedType != "" {
		switch auth.SelectedType {
		case "api-key":
			key := auth.AnthropicAPIKey
			if key == "" {
				key = auth.OpenAIAPIKey
			}
			if key == "" {
				return nil, fmt.Errorf("pi: auth type %q selected but no API key found; set ANTHROPIC_API_KEY or OPENAI_API_KEY", auth.SelectedType)
			}
			envKey := "ANTHROPIC_API_KEY"
			if auth.AnthropicAPIKey == "" {
				envKey = "OPENAI_API_KEY"
			}
			return &api.ResolvedAuth{
				Method:  "api-key",
				EnvVars: map[string]string{envKey: key},
			}, nil
		case "auth-file":
			if auth.PiAuthFile == "" {
				return nil, fmt.Errorf("pi: auth type %q selected but no auth file found; expected ~/.pi/agent/auth.json", auth.SelectedType)
			}
			return &api.ResolvedAuth{
				Method: "auth-file",
				Files: []api.FileMapping{
					{SourcePath: auth.PiAuthFile, ContainerPath: "~/.pi/agent/auth.json"},
				},
			}, nil
		default:
			return nil, fmt.Errorf("pi: unknown auth type %q; valid types are: api-key, auth-file", auth.SelectedType)
		}
	}

	// Auto-detect preference order: AnthropicAPIKey → OpenAIAPIKey → PiAuthFile → error

	if auth.AnthropicAPIKey != "" {
		return &api.ResolvedAuth{
			Method:  "api-key",
			EnvVars: map[string]string{"ANTHROPIC_API_KEY": auth.AnthropicAPIKey},
		}, nil
	}

	if auth.OpenAIAPIKey != "" {
		return &api.ResolvedAuth{
			Method:  "api-key",
			EnvVars: map[string]string{"OPENAI_API_KEY": auth.OpenAIAPIKey},
		}, nil
	}

	if auth.PiAuthFile != "" {
		return &api.ResolvedAuth{
			Method: "auth-file",
			Files: []api.FileMapping{
				{SourcePath: auth.PiAuthFile, ContainerPath: "~/.pi/agent/auth.json"},
			},
		}, nil
	}

	return nil, fmt.Errorf("pi: no valid auth method found; set ANTHROPIC_API_KEY or OPENAI_API_KEY, or provide auth credentials at ~/.pi/agent/auth.json")
}

# Scion

Scion is a container-based orchestration tool for managing concurrent LLM-based code agents. It enables parallel execution of specialized sub-agents with isolated identities, credentials, and workspaces.

## Installation

You can install Scion directly from the source using Go:

```bash
go install github.com/ptone/scion-agent@latest
```

Ensure that your `$GOPATH/bin` is in your system `$PATH`.

## Concepts

- **Agent**: An isolated container running an LLM-driven task. Each agent has its own home directory, workspace, and credentials.
- **Grove**: A "grove" is a project workspace where agents live. It corresponds to a `.scion` directory which contains agent configurations and templates.
- **Template**: A blueprint for an agent, defining its base configuration, system prompt, and tools.

## Quick Start

### 1. Initialize a Grove

Navigate to your project root and initialize a new Scion grove. This creates the `.scion` directory structure.

```bash
cd my-project
scion grove init
```

### 2. Start an Agent

Launch a new agent to perform a specific task. By default, this runs in the background.

```bash
# Start a generic agent named "coder"
scion start coder "Refactor the authentication middleware in pkg/auth"

# Start a specialized agent using a template
scion start auditor "Audit the user input validation" --type security-auditor

# Start and immediately attach to the session
scion start debug "Help me debug this error" --attach
```

### 3. List Running Agents

View all active agents in the current grove.

```bash
scion list
```

To see agents across all groves:

```bash
scion list --all
```

### 4. Interact with an Agent

You can attach to a running agent's interactive session (TTY).

```bash
scion attach coder
```

Use `Ctrl+P, Ctrl+Q` to detach without stopping the container (if using Docker/default runtime behavior), or simply exit the shell to stop the agent.

### 5. View Logs

Check the logs of a background agent.

```bash
scion logs coder
```

### 6. Stop and Cleanup

Stop a running agent:

```bash
scion stop coder
```

Delete an agent (removes the container, agent directory, and git worktree):

```bash
scion delete coder
```

## Configuration

Scion uses a `settings.json` file located in `~/.gemini/settings.json` for global configuration, such as API keys and auth preferences.

## License

This project is licensed under the Apache License, Version 2.0. See the [LICENSE](LICENSE) file for details.
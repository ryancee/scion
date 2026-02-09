---
title: Team Workflow
description: Connecting to a Scion Hub for team collaboration.
---

While Scion works great in "Solo" mode for individual developers, its true power is realized in a "Hosted" architecture where teams can share state, infrastructure, and agent configurations.

## Connecting to a Hub

To join a team environment, you first need to connect your local CLI to the team's Scion Hub.

```bash
# Set the Hub endpoint globally
scion config set hub.endpoint https://scion.yourcompany.com --global

# Enable Hub integration
scion hub enable

# Login via the browser
scion hub auth login
```

Once authenticated, your CLI will route agent operations through the Hub instead of running them purely locally.

## Verifying Connection

You can check your connection status and authentication details at any time:

```bash
scion hub status
```

This will show:
- Whether Hub integration is enabled.
- The configured Hub endpoint.
- Your authentication method (OAuth, Dev Auth, etc.).
- Your user identity, role, and token expiration.
- **Grove Context**: Whether the current directory is registered with the Hub and which brokers are available to it.

## Linking a Project (Grove)

In a team environment, a "Grove" represents a shared project or repository. Registration links your local repository to a central Grove ID on the Hub.

1.  Navigate to your project directory.
2.  Run the link command: `scion hub link`
3.  The Hub will attempt to match your project by its **Git remote**.
4.  If a match is found, your local `.scion` config is linked to the existing Grove.
5.  If no match is found, you can register it as a new Grove.

To see all groves you have access to on the Hub:
```bash
scion hub groves
```

## Shared Infrastructure (Brokers)

When you start an agent in a team workflow, the Hub dispatches it to an available **Runtime Broker**.

### Using Remote Brokers
A Runtime Broker is a compute node that has been registered with the Hub. To see available brokers:
```bash
scion hub brokers
```

### Providing Execution for a Grove
A broker only executes agents for the Groves it is a **provider** for. If you want to use your local machine as a broker for a team project:
1.  Start the broker: `scion broker start`
2.  Register the broker: `scion broker register`
3.  Add it as a provider for the current project: `scion broker provide`

## Managing Secrets

In a team workflow, secrets should be managed on the Hub rather than in local `.env` files. This ensures they are available to any broker that executes your agents.

```bash
# Set a secret for the current grove
scion hub secret set GITHUB_TOKEN=ghp_...

# The secret is securely stored in the Hub and injected into agents at runtime.
```


## Collaborating

- **Shared Visibility**: Use the Web Dashboard to see what agents your team is running.
- **Shared Templates**: Use centrally managed templates for consistent agent behavior across the team.
- **Attach to Remote Agents**: You can `scion attach` to an agent running on a remote Runtime Broker just as if it were local.

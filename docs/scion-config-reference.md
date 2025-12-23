# scion.json Configuration Reference

The `scion.json` file is used within templates and agent directories to configure how a Scion agent is provisioned and executed.

## Fields

### `template` (string)
The name of the template this configuration belongs to.
- **Example**: `"template": "default"`

### `unix_username` (string)
The username used for the primary user inside the container.
- **Default**: `node`
- **Example**: `"unix_username": "developer"`

### `image` (string)
The container image to use for the agent.
- **Default**: `gemini-cli-sandbox`
- **Example**: `"image": "my-custom-gemini-agent:latest"`

### `detached` (boolean)
Whether the agent should run in detached mode by default.
- **Default**: `true`
- **Example**: `"detached": false`

### `use_tmux` (boolean)
If set to `true`, the agent's main process will be wrapped in a `tmux` session. This enables persistent interactive sessions that can be detached and re-attached using the `scion attach` command.
- **Default**: `false`
- **Details**: When enabled, `scion` will attempt to use a version of the configured image with a `:tmux` tag if available (e.g. `gemini-cli-sandbox:tmux`).
- **Example**: `"use_tmux": true`

### `model` (string)
The Gemini model ID to use for the agent.
- **Default**: `"flash"`
- **Details**: This value is propagated to the agent container as the `GEMINI_MODEL` environment variable. The value `"flash"` is a special ID recognized by the Gemini CLI as the best available flash model.
- **Example**: `"model": "pro"`

### `agent` (object)
*Internal usage*: This object is typically populated by the Scion CLI during provisioning to track instance-specific state.
- **Fields**: `grove`, `name`, `status`.

## Inheritance
`scion` uses a template inheritance system. Configuration fields are merged from the `default` template, then the specified template type, and finally any overrides in the agent's own directory. The last value defined for a field takes precedence.
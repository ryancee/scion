package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ptone/gswarm/pkg/config"
	"github.com/ptone/gswarm/pkg/runtime"
	"github.com/ptone/gswarm/pkg/util"
	"github.com/spf13/cobra"
)

var (
	agentName    string
	templateName string
	agentImage   string
	noAuth       bool
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start <task>",
	Short: "Launch a new gswarm agent",
	Long: `Provision and launch a new isolated Gemini agent to perform a specific task.
The agent will be created from a template and run in a detached container.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		task := args[0]

		if agentName == "" {
			// Generate a unique name if not provided
			agentName = fmt.Sprintf("agent-%d", os.Getpid())
		}

		fmt.Printf("Starting agent '%s' for task: %s\n", agentName, task)

		// 1. Prepare agent home directory
		agentsDir, err := config.GetProjectAgentsDir()
		if err != nil {
			return err
		}
		agentHome := filepath.Join(agentsDir, agentName, "home")
		if err := os.MkdirAll(agentHome, 0755); err != nil {
			return fmt.Errorf("failed to create agent home: %w", err)
		}

		// 2. Load and copy templates
		chain, err := config.GetTemplateChain(templateName)
		if err != nil {
			return fmt.Errorf("failed to load template: %w", err)
		}

		// Track image from templates
		resolvedImage := ""

		for _, tpl := range chain {
			fmt.Printf("Applying template: %s\n", tpl.Name)
			if err := util.CopyDir(tpl.Path, agentHome); err != nil {
				return fmt.Errorf("failed to copy template %s: %w", tpl.Name, err)
			}

			// Load gswarm.json from this template to see if it specifies an image
			tplCfg, err := tpl.LoadConfig()
			if err == nil && tplCfg.Image != "" {
				resolvedImage = tplCfg.Image
			}
		}

		// Flag takes ultimate precedence
		if agentImage != "" {
			resolvedImage = agentImage
		}
		if resolvedImage == "" {
			resolvedImage = "gemini-cli-sandbox"
		}

		// 3. Propagate credentials
		var auth config.AuthConfig
		if !noAuth {
			auth = config.DiscoverAuth()
		}

		// 4. Launch container
		rt := runtime.GetRuntime()
		runCfg := runtime.RunConfig{
			Name:    agentName,
			Image:   resolvedImage,
			HomeDir: agentHome,
			Auth:    auth,
			Env: []string{
				fmt.Sprintf("GEMINI_INITIAL_PROMPT=%s", task),
			},
			Labels: map[string]string{
				"gswarm.agent": "true",
				"gswarm.name":  agentName,
			},
		}
		
				id, err := rt.RunDetached(context.Background(), runCfg)
					if err != nil {
						return fmt.Errorf("failed to launch container: %w", err)
					}
		
					fmt.Printf("Agent '%s' launched successfully (ID: %s)\n", agentName, id)
					return nil
			},
		}
		
		func init() {
			rootCmd.AddCommand(startCmd)
			startCmd.Flags().StringVarP(&agentName, "name", "n", "", "Name of the agent")
			startCmd.Flags().StringVarP(&templateName, "type", "t", "default", "Template to use")
				startCmd.Flags().StringVarP(&agentImage, "image", "i", "", "Container image to use (overrides template)")
				startCmd.Flags().BoolVar(&noAuth, "no-auth", false, "Disable authentication propagation")
			}
			

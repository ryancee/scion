package runtime

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGetRuntime(t *testing.T) {
	// Save original env
	orig := os.Getenv("GEMINI_SANDBOX")
	defer os.Setenv("GEMINI_SANDBOX", orig)

	t.Run("EnvVar_Container", func(t *testing.T) {
		os.Setenv("GEMINI_SANDBOX", "container")
		r := GetRuntime("")
		if _, ok := r.(*AppleContainerRuntime); !ok {
			t.Errorf("expected *AppleContainerRuntime, got %T", r)
		}
	})

	t.Run("EnvVar_Local", func(t *testing.T) {
		os.Setenv("GEMINI_SANDBOX", "local")
		r := GetRuntime("")
		if runtime.GOOS == "darwin" {
			if _, ok := r.(*AppleContainerRuntime); !ok {
				t.Errorf("expected *AppleContainerRuntime on darwin, got %T", r)
			}
		} else {
			if _, ok := r.(*DockerRuntime); !ok {
				t.Errorf("expected *DockerRuntime on %s, got %T", runtime.GOOS, r)
			}
		}
	})

	t.Run("EnvVar_Remote", func(t *testing.T) {
		os.Setenv("GEMINI_SANDBOX", "remote")
		r := GetRuntime("")
		// Remote resolves to kubernetes, which currently falls back to docker in the factory switch
		if _, ok := r.(*DockerRuntime); !ok {
			t.Errorf("expected *DockerRuntime (fallback from kubernetes), got %T", r)
		}
	})

	t.Run("EnvVar_Docker", func(t *testing.T) {
		os.Setenv("GEMINI_SANDBOX", "docker")
		r := GetRuntime("")
		if _, ok := r.(*DockerRuntime); !ok {
			t.Errorf("expected *DockerRuntime, got %T", r)
		}
	})

	t.Run("Default_AutoDetect", func(t *testing.T) {
		os.Unsetenv("GEMINI_SANDBOX")
		// This depends on what's available in the environment running the test.
		// We just want to ensure it doesn't panic and returns a valid runtime.
		r := GetRuntime("")
		if r == nil {
			t.Error("expected a runtime instance, got nil")
		}
	})

	t.Run("Settings_File", func(t *testing.T) {
		os.Unsetenv("GEMINI_SANDBOX")
		
		// Mock HOME for settings
		tmpHome := t.TempDir()
		origHome := os.Getenv("HOME")
		os.Setenv("HOME", tmpHome)
		defer os.Setenv("HOME", origHome)

		geminiDir := filepath.Join(tmpHome, ".gemini")
		if err := os.MkdirAll(geminiDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Write settings to force "container"
		// Note: The factory logic checks for string or bool.
		err := os.WriteFile(filepath.Join(geminiDir, "settings.json"), 
			[]byte(`{"tools": {"sandbox": "container"}}`), 0644)
		if err != nil {
			t.Fatal(err)
		}

		r := GetRuntime("")
		if _, ok := r.(*AppleContainerRuntime); !ok {
			t.Errorf("expected *AppleContainerRuntime from settings, got %T", r)
		}
	})
}

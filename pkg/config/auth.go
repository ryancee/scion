package config

import (
	"os"
)

type AuthConfig struct {
	GeminiAPIKey               string
	GoogleAPIKey               string
	VertexAPIKey               string
	GoogleAppCredentials       string
	GoogleCloudProject         string
}

func DiscoverAuth() AuthConfig {
	auth := AuthConfig{
		GeminiAPIKey:         os.Getenv("GEMINI_API_KEY"),
		GoogleAPIKey:         os.Getenv("GOOGLE_API_KEY"),
		VertexAPIKey:         os.Getenv("VERTEX_API_KEY"),
		GoogleAppCredentials: os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
		GoogleCloudProject:   os.Getenv("GOOGLE_CLOUD_PROJECT"),
	}

	if auth.GoogleCloudProject == "" {
		auth.GoogleCloudProject = os.Getenv("GCP_PROJECT")
	}

	// Fallback to settings.json for Gemini API Key if none found in env
	if auth.GeminiAPIKey == "" && auth.GoogleAPIKey == "" {
		if settings, err := GetGeminiSettings(); err == nil {
			if settings.ApiKey != "" {
				auth.GeminiAPIKey = settings.ApiKey
			}
		}
	}

	return auth
}

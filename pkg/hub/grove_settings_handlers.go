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

package hub

import (
	"net/http"
	"strconv"

	"github.com/GoogleCloudPlatform/scion/pkg/hubclient"
	"github.com/GoogleCloudPlatform/scion/pkg/store"
)

// Annotation keys for grove settings stored in grove annotations.
const (
	groveSettingDefaultTemplate      = "scion.io/default-template"
	groveSettingDefaultHarnessConfig = "scion.io/default-harness-config"
	groveSettingTelemetryEnabled     = "scion.io/telemetry-enabled"
	groveSettingActiveProfile        = "scion.io/active-profile"
)

// handleGroveSettings handles GET/PUT on /api/v1/groves/{groveId}/settings.
func (s *Server) handleGroveSettings(w http.ResponseWriter, r *http.Request, groveID string) {
	ctx := r.Context()

	grove, err := s.store.GetGrove(ctx, groveID)
	if err != nil {
		if err == store.ErrNotFound {
			NotFound(w, "Grove")
			return
		}
		writeErrorFromErr(w, err, "")
		return
	}

	identity := GetIdentityFromContext(ctx)
	if identity == nil {
		Unauthorized(w)
		return
	}

	switch r.Method {
	case http.MethodGet:
		if userIdent, ok := identity.(UserIdentity); ok {
			decision := s.authzService.CheckAccess(ctx, userIdent, Resource{
				Type:    "grove",
				ID:      grove.ID,
				OwnerID: grove.OwnerID,
			}, ActionRead)
			if !decision.Allowed {
				Forbidden(w)
				return
			}
		}

		writeJSON(w, http.StatusOK, groveSettingsFromAnnotations(grove))

	case http.MethodPut:
		if userIdent, ok := identity.(UserIdentity); ok {
			decision := s.authzService.CheckAccess(ctx, userIdent, Resource{
				Type:    "grove",
				ID:      grove.ID,
				OwnerID: grove.OwnerID,
			}, ActionUpdate)
			if !decision.Allowed {
				Forbidden(w)
				return
			}
		} else {
			Forbidden(w)
			return
		}

		var req hubclient.GroveSettings
		if err := readJSON(r, &req); err != nil {
			BadRequest(w, "Invalid request body: "+err.Error())
			return
		}

		applyGroveSettingsToAnnotations(grove, &req)

		if err := s.store.UpdateGrove(ctx, grove); err != nil {
			writeErrorFromErr(w, err, "")
			return
		}

		s.events.PublishGroveUpdated(ctx, grove)
		writeJSON(w, http.StatusOK, groveSettingsFromAnnotations(grove))

	default:
		MethodNotAllowed(w)
	}
}

// groveSettingsFromAnnotations reads grove settings from the grove's annotations map.
func groveSettingsFromAnnotations(grove *store.Grove) *hubclient.GroveSettings {
	settings := &hubclient.GroveSettings{}
	if grove.Annotations == nil {
		return settings
	}

	settings.DefaultTemplate = grove.Annotations[groveSettingDefaultTemplate]
	settings.DefaultHarnessConfig = grove.Annotations[groveSettingDefaultHarnessConfig]
	settings.ActiveProfile = grove.Annotations[groveSettingActiveProfile]

	if val, ok := grove.Annotations[groveSettingTelemetryEnabled]; ok {
		if b, err := strconv.ParseBool(val); err == nil {
			settings.TelemetryEnabled = &b
		}
	}

	return settings
}

// applyGroveSettingsToAnnotations writes grove settings into the grove's annotations map.
func applyGroveSettingsToAnnotations(grove *store.Grove, settings *hubclient.GroveSettings) {
	if grove.Annotations == nil {
		grove.Annotations = make(map[string]string)
	}

	setOrDelete(grove.Annotations, groveSettingDefaultTemplate, settings.DefaultTemplate)
	setOrDelete(grove.Annotations, groveSettingDefaultHarnessConfig, settings.DefaultHarnessConfig)
	setOrDelete(grove.Annotations, groveSettingActiveProfile, settings.ActiveProfile)

	if settings.TelemetryEnabled != nil {
		grove.Annotations[groveSettingTelemetryEnabled] = strconv.FormatBool(*settings.TelemetryEnabled)
	} else {
		delete(grove.Annotations, groveSettingTelemetryEnabled)
	}
}

// setOrDelete sets an annotation key to value, or deletes it if value is empty.
func setOrDelete(m map[string]string, key, value string) {
	if value == "" {
		delete(m, key)
	} else {
		m[key] = value
	}
}

package dmtools

import (
	"fmt"

	"dungeons/internal/adventure"
)

// NewStartSessionTool creates a tool to start a new game session.
func NewStartSessionTool(adv *adventure.Adventure) *SimpleTool {
	return &SimpleTool{
		name:        "start_session",
		description: "Start a new game session. This MUST be called at the beginning of each play session to properly track events and journal entries.",
		schema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			// Start the session
			session, err := adv.StartSession()
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   err.Error(),
				}, nil
			}

			return map[string]interface{}{
				"success":    true,
				"session_id": session.ID,
				"started_at": session.StartedAt.Format("2006-01-02 15:04:05"),
				"display":    fmt.Sprintf("Session %d démarrée", session.ID),
			}, nil
		},
	}
}

// NewEndSessionTool creates a tool to end the current game session.
func NewEndSessionTool(adv *adventure.Adventure) *SimpleTool {
	return &SimpleTool{
		name:        "end_session",
		description: "End the current game session with a summary. This MUST be called when players finish playing to properly close the session and organize the journal.",
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"summary": map[string]interface{}{
					"type":        "string",
					"description": "Summary of what happened during this session (in French)",
				},
			},
			"required": []string{"summary"},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			summary := params["summary"].(string)

			// End the session
			session, err := adv.EndSession(summary)
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   err.Error(),
				}, nil
			}

			return map[string]interface{}{
				"success":    true,
				"session_id": session.ID,
				"duration":   session.Duration,
				"summary":    session.Summary,
				"display":    fmt.Sprintf("Session %d terminée - Durée: %s", session.ID, session.Duration),
			}, nil
		},
	}
}

// NewGetSessionInfoTool creates a tool to get information about the current session.
func NewGetSessionInfoTool(adv *adventure.Adventure) *SimpleTool {
	return &SimpleTool{
		name:        "get_session_info",
		description: "Get information about the current active session (if any)",
		schema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			session, err := adv.GetCurrentSession()
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   err.Error(),
				}, nil
			}

			if session == nil {
				return map[string]interface{}{
					"success":        true,
					"session_active": false,
					"display":        "Aucune session active",
				}, nil
			}

			return map[string]interface{}{
				"success":        true,
				"session_active": true,
				"session_id":     session.ID,
				"started_at":     session.StartedAt.Format("2006-01-02 15:04:05"),
				"display":        fmt.Sprintf("Session %d en cours depuis %s", session.ID, session.StartedAt.Format("15:04")),
			}, nil
		},
	}
}

package uagent

import "time"

type UserAgentResponse struct {
	ID        int            `json:"id"`
	Proxy     map[string]any `json:"proxy"`
	Status    int            `json:"status"`
	UpdatedAt time.Time      `json:"updated_at"`
	UserAgent map[string]any `json:"user_agent"`
}

package domain

import "time"

type ServiceStatus string

const (
	StatusActive      ServiceStatus = "active"      // StatusActive - активность сервера, полная доступность к функциям сервера
	StatusInactive    ServiceStatus = "inactive"    // StatusInactive - сервер доступен, но не доступны его функции
	StatusUnavailable ServiceStatus = "unavailable" // StatusUnavailable - сервер полностью недоступен
)

type WebHookType string

const (
	Switch    WebHookType = "switch"    // Switch - переключатель. Например: вкл/выкл, A/B, короче можно выбирать между двух
	Indicator WebHookType = "indicator" // Indicator - используется базово для ServiceStatus (статус сервиса), можно сделать свои реализации этого
)

type WebHook struct {
	ID         string      `json:"webhook_id"`
	Name       string      `json:"webhook_name"`
	Path       string      `json:"webhook_path"`
	Type       WebHookType `json:"webhook_type"`
	Method     string      `json:"webhook_method"`
	Executions int         `json:"webhook_executions"`
	LastCall   time.Time   `json:"webhook_last_call"`
}

type Service struct {
	ID        string        `json:"service_id"`
	Name      string        `json:"service_name"`
	Status    ServiceStatus `json:"service_status"`
	WebHooks  []WebHook     `json:"webhooks"`
	CreatedAt time.Time     `json:"service_created_at"`
	UpdatedAt time.Time     `json:"service_updated_at"`
}

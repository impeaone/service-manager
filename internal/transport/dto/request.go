package dto

type ServicesResponse struct {
	Services []ServiceResponse `json:"services"`
}

type ServiceResponse struct {
	ID        string            `json:"service_id"`
	Name      string            `json:"service_name"`
	Status    string            `json:"service_status"`
	WebHooks  []WebHookResponse `json:"service_webhooks"`
	CreatedAt string            `json:"service_created_at,omitempty"`
}

type WebHookResponse struct {
	ID         string `json:"webhook_id"`
	Name       string `json:"webhook_name"`
	Type       string `json:"webhook_type"`
	Path       string `json:"webhook_path"`
	Method     string `json:"webhook_method"`
	Executions int    `json:"webhook_executions"`
	LastCalled string `json:"webhook_last_called,omitempty"`
}

type WebHookRequest struct {
	Status     string                 `json:"status"`
	Body       map[string]interface{} `json:"body"`
	StatusCode int                    `json:"status_code"`
}

type DeleteServiceResponse struct {
	ID string `json:"id"`
}

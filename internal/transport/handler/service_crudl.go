package handler

import (
	"ServiceManager/internal/domain"
	"ServiceManager/internal/service"
	"ServiceManager/internal/transport/dto"
	"ServiceManager/pkg/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const DefaultRequestTimeout = time.Second * 10

var ErrorHTMLinsteadJSON = fmt.Errorf("invalid character '<' looking for beginning of value")

type APIHandler struct {
	serviceManager service.ServiceManager
	logger         *slog.Logger
	ctx            context.Context
}

func NewAPIHandler(ctx context.Context, serviceManager service.ServiceManager) *APIHandler {
	return &APIHandler{serviceManager: serviceManager, ctx: ctx}
}

func (a *APIHandler) GetServices(w http.ResponseWriter, _ *http.Request) {
	// Получаем все наши сервисы
	services, err := a.serviceManager.GetAllServices()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Формируем ответ
	response := dto.ServicesResponse{
		Services: make([]dto.ServiceResponse, 0, len(services)),
	}

	for _, serv := range services {
		serviceResp := dto.ServiceResponse{
			ID:        serv.ID,
			Name:      serv.Name,
			Status:    string(serv.Status),
			WebHooks:  make([]dto.WebHookResponse, 0, len(serv.WebHooks)),
			CreatedAt: serv.CreatedAt.String(),
		}

		for _, wh := range serv.WebHooks {
			hookResp := dto.WebHookResponse{
				ID:         wh.ID,
				Name:       wh.Name,
				Path:       wh.Path,
				Method:     wh.Method,
				Executions: wh.Executions,
			}
			if !wh.LastCall.IsZero() {
				hookResp.LastCalled = wh.LastCall.String()
			}
			serviceResp.WebHooks = append(serviceResp.WebHooks, hookResp)
		}
		response.Services = append(response.Services, serviceResp)
	}

	utils.SendJSON(w, response, http.StatusOK)
	return
}

func (a *APIHandler) GetService(w http.ResponseWriter, r *http.Request) {

	serviceID := r.PathValue("service_id")
	if serviceID == "" {
		resp := map[string]interface{}{"error": "invalid service id"}
		utils.SendJSON(w, resp, http.StatusNotFound)
		return
	}

	servic, err := a.serviceManager.GetService(serviceID)
	if err != nil {
		resp := map[string]interface{}{"error": err.Error()}
		utils.SendJSON(w, resp, http.StatusNotFound)
		return
	}

	var serviceResp = dto.ServiceResponse{
		ID:        servic.ID,
		Name:      servic.Name,
		Status:    string(servic.Status),
		WebHooks:  make([]dto.WebHookResponse, len(servic.WebHooks)),
		CreatedAt: servic.CreatedAt.String(),
	}
	for i, wh := range servic.WebHooks {
		serviceResp.WebHooks[i] = dto.WebHookResponse{
			ID:         wh.ID,
			Name:       wh.Name,
			Path:       wh.Path,
			Method:     wh.Method,
			Type:       string(wh.Type),
			Executions: wh.Executions,
			LastCalled: wh.LastCall.String(),
		}
	}

	utils.SendJSON(w, serviceResp, http.StatusOK)
	return
}

func (a *APIHandler) DeleteService(w http.ResponseWriter, r *http.Request) {

	serviceID := r.PathValue("service_id")
	if serviceID == "" {
		resp := map[string]interface{}{"error": "invalid service id"}
		utils.SendJSON(w, resp, http.StatusNotFound)
		return
	}

	if err := a.serviceManager.DeleteService(serviceID); err != nil {
		resp := map[string]interface{}{"error": err.Error()}
		utils.SendJSON(w, resp, http.StatusInternalServerError)
		return
	}

	utils.SendJSON(w, dto.DeleteServiceResponse{
		ID: serviceID,
	}, http.StatusNoContent)
	return
}

func (a *APIHandler) AddService(w http.ResponseWriter, r *http.Request) {
	// Читаем тело запроса
	data, err := io.ReadAll(r.Body)
	defer func() { _ = r.Body.Close() }()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var servic dto.ServiceResponse
	if err = json.Unmarshal(data, &servic); err != nil {
		resp := map[string]interface{}{"error": err.Error()}
		utils.SendJSON(w, resp, http.StatusBadRequest)
		return
	}
	// Сохраняем новый сервис (и получаем его уже с id)
	newService, err := a.serviceManager.CreateService(servic)
	if err != nil {
		resp := map[string]interface{}{"error": err.Error()}
		utils.SendJSON(w, resp, http.StatusBadRequest)
		return
	}
	// Собираем ответ и выдаем
	var newServiceResponse = dto.ServiceResponse{
		ID:        newService.ID,
		Name:      newService.Name,
		Status:    string(newService.Status),
		WebHooks:  make([]dto.WebHookResponse, len(newService.WebHooks)),
		CreatedAt: newService.CreatedAt.String(),
	}

	for i, wh := range newService.WebHooks {
		newServiceResponse.WebHooks[i] = dto.WebHookResponse{
			ID:         wh.ID,
			Name:       wh.Name,
			Path:       wh.Path,
			Type:       string(wh.Type),
			Method:     wh.Method,
			Executions: wh.Executions,
			LastCalled: wh.LastCall.String(),
		}
	}

	utils.SendJSON(w, newServiceResponse, http.StatusOK)
	return
}

func (a *APIHandler) UpdateService(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	defer func() { _ = r.Body.Close() }()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var servic dto.ServiceResponse
	if err = json.Unmarshal(data, &servic); err != nil {
		resp := map[string]interface{}{"error": err.Error()}
		utils.SendJSON(w, resp, http.StatusBadRequest)
		return
	}

	updService, err := a.serviceManager.UpdateService(servic)
	if err != nil {
		resp := map[string]interface{}{"error": err.Error()}
		utils.SendJSON(w, resp, http.StatusBadRequest)
		return
	}

	var updServiceResponse = dto.ServiceResponse{
		ID:        updService.ID,
		Name:      updService.Name,
		Status:    string(updService.Status),
		WebHooks:  make([]dto.WebHookResponse, len(updService.WebHooks)),
		CreatedAt: updService.CreatedAt.String(),
	}

	for i, wh := range updService.WebHooks {
		updServiceResponse.WebHooks[i] = dto.WebHookResponse{
			ID:         wh.ID,
			Name:       wh.Name,
			Path:       wh.Path,
			Type:       string(wh.Type),
			Method:     wh.Method,
			Executions: wh.Executions,
			LastCalled: wh.LastCall.String(),
		}
	}

	utils.SendJSON(w, updServiceResponse, http.StatusOK)
	return
}

// ExecuteWebHook example http://localhost:8080/api/services/execute?service_id=123&webhook_id=123
func (a *APIHandler) ExecuteWebHook(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()

	serviceID := r.URL.Query().Get("service_id")
	webhookID := r.URL.Query().Get("webhook_id")

	if serviceID == "" || webhookID == "" {
		resp := map[string]interface{}{"error": "service_id or webhook_id is empty"}
		utils.SendJSON(w, resp, http.StatusBadRequest)
		return
	}

	serv, err := a.serviceManager.GetService(serviceID)
	if err != nil {
		resp := map[string]interface{}{"error": err.Error()}
		utils.SendJSON(w, resp, http.StatusBadRequest)
		return
	}

	var targetWebHook domain.WebHook
	for _, wh := range serv.WebHooks {
		if wh.ID == webhookID {
			targetWebHook = wh
			break
		}
	}

	hookID := targetWebHook.ID
	hookPath := targetWebHook.Path
	hookMethod := targetWebHook.Method

	req, err := a.webHookRequest(hookPath, hookMethod, r.Body)
	if err != nil {
		resp := map[string]interface{}{"error": err.Error()}
		utils.SendJSON(w, resp, http.StatusBadRequest)
		return
	}

	go func() {
		_ = a.serviceManager.IncrementWebHook(serviceID, hookID)
	}()

	utils.SendJSON(w, req, http.StatusOK)
	return
}

func (a *APIHandler) webHookRequest(path, method string, body io.Reader) (dto.WebHookRequest, error) {
	ctx, cancel := context.WithTimeout(a.ctx, DefaultRequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, path, body)
	if err != nil {
		return dto.WebHookRequest{}, err
	}

	client := new(http.Client)

	errChan := make(chan error, 1)
	respChan := make(chan *http.Response, 1)

	go func() {
		resp, err := client.Do(req)
		if err != nil {
			errChan <- err
		}
		respChan <- resp
	}()

	select {
	case <-ctx.Done():
		return dto.WebHookRequest{}, fmt.Errorf("request timeout: %w", ctx.Err())
	case err = <-errChan:
		return dto.WebHookRequest{}, fmt.Errorf("request failed: %w", err)
	case resp := <-respChan:
		bodyReq, err := utils.ReadCloserToJSONMap(resp.Body)
		if err != nil && !errors.As(err, &ErrorHTMLinsteadJSON) {
			return dto.WebHookRequest{}, err
		}

		return dto.WebHookRequest{
			StatusCode: resp.StatusCode,
			Body:       bodyReq,
			Status:     resp.Status,
		}, nil
	}

}

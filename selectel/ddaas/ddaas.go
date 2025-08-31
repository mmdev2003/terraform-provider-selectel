package ddaas

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	// appName specifies an application name.
	appName = "ddaas-go"

	// appVersion specifies an application version.
	appVersion = "0.1.0"

	// userAgent contains a basic user agent that will be used in queries.
	userAgent = appName + "/" + appVersion
)

func New(token, endpoint string) (*API, error) {
	return &API{
		HTTPClient: http.DefaultClient,
		Token:      token,
		Endpoint:   endpoint,
		UserAgent:  userAgent,
	}, nil
}

type API struct {
	HTTPClient *http.Client
	Token      string
	Endpoint   string
	UserAgent  string
}

type DedicatedServerAPIError struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Detail  string `json:"detail,omitempty"`
}

func (e *DedicatedServerAPIError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("API error %d: %s (%s)", e.Code, e.Message, e.Detail)
	}
	return fmt.Sprintf("API error %d: %s", e.Code, e.Message)
}

type Status string

const (
	StatusActive      Status = "ACTIVE"
	StatusBuilding    Status = "BUILDING"
	StatusRebooting   Status = "REBOOTING"
	StatusReinstall   Status = "REINSTALL"
	StatusDeleted     Status = "DELETED"
	StatusError       Status = "ERROR"
	StatusMaintenance Status = "MAINTENANCE"
)

// Location представляет локацию/пул серверов
type Location struct {
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	LocationID  int    `json:"location_id"`
	Description string `json:"description"`
	Enable      bool   `json:"enable"`
}

// Configuration представляет конфигурацию сервера
type Configuration struct {
	UUID         string `json:"uuid"`
	Name         string `json:"name"`
	TariffLine   string `json:"tariff_line"`
	Model        string `json:"model"`
	CPU          string `json:"cpu,omitempty"`
	RAM          string `json:"ram,omitempty"`
	Storage      string `json:"storage,omitempty"`
	LocationUUID string `json:"location_uuid,omitempty"`
}

// Tariff представляет тариф
type Tariff struct {
	UUID              string `json:"uuid"`
	Name              string `json:"name"`
	Period            string `json:"period"`
	Price             string `json:"price"`
	Currency          string `json:"currency"`
	ConfigurationUUID string `json:"configuration_uuid,omitempty"`
}

// OSImage представляет образ операционной системы
type OSImage struct {
	UUID         string            `json:"uuid"`
	Name         string            `json:"name"`
	Family       string            `json:"family"`
	Version      string            `json:"version"`
	Architecture string            `json:"architecture"`
	Parameters   map[string]string `json:"parameters,omitempty"`
}

// Network представляет сеть
type Network struct {
	UUID         string `json:"uuid"`
	Name         string `json:"name"`
	Type         string `json:"type"` // public, private
	LocationUUID string `json:"location_uuid"`
	VLAN         int    `json:"vlan,omitempty"`
}

// DedicatedServer основная структура сервера
type DedicatedServer struct {
	UUID              string                 `json:"uuid"`
	Name              string                 `json:"name"`
	Status            Status                 `json:"status"`
	ProjectID         string                 `json:"project_id"`
	LocationUUID      string                 `json:"location_uuid"`
	ServiceUUID       string                 `json:"service_uuid"`
	ConfigurationUUID string                 `json:"configuration_uuid"`
	TariffUUID        string                 `json:"tariff_uuid"`
	OSImageUUID       string                 `json:"os_image_uuid"`
	IPAddresses       []IPAddress            `json:"ip_addresses,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
	OsParams          map[string]interface{} `json:"os_params,omitempty"`
}

// IPAddress структура IP адреса
type IPAddress struct {
	Type    string `json:"type"` // public, private
	IP      string `json:"ip"`
	Netmask string `json:"netmask"`
	Gateway string `json:"gateway"`
}

// DedicatedServerCreateOpts параметры создания сервера
type DedicatedServerCreateOpts struct {
	ProjectID          string                 `json:"project_id"`
	LocationUUID       string                 `json:"location_uuid"`
	ConfigurationUUID  string                 `json:"configuration_uuid"`
	TariffUUID         string                 `json:"tariff_uuid"`
	OSImageUUID        string                 `json:"os_image_uuid"`
	Name               string                 `json:"name,omitempty"`
	PublicNetworkUUID  string                 `json:"public_network_uuid,omitempty"`
	PrivateNetworkUUID string                 `json:"private_network_uuid,omitempty"`
	OsParams           map[string]interface{} `json:"os_params,omitempty"`
}

// DedicatedServerUpdateOpts параметры обновления сервера
type DedicatedServerUpdateOpts struct {
	OSImageUUID string                 `json:"os_image_uuid,omitempty"`
	OsParams    map[string]interface{} `json:"os_params,omitempty"`
}

// DedicatedServerQueryParams параметры поиска серверов
type DedicatedServerQueryParams struct {
	UUID      string `json:"uuid,omitempty"`
	ProjectID string `json:"project_id,omitempty"`
	Name      string `json:"name,omitempty"`
	Status    Status `json:"status,omitempty"`
}

const (
	// API endpoints
	DedicatedServerURI = "/servers/v2/resource"
	LocationURI        = "/servers/v2/location"
	ConfigurationURI   = "/servers/v2/service"
	TariffURI          = "/servers/v2/tariff"
	OSImageURI         = "/servers/v2/boot/template/os/new"
	NetworkURI         = "/servers/v2/network"
)

// Методы для работы с серверами
func (api *API) DedicatedServers(ctx context.Context, params *DedicatedServerQueryParams) ([]DedicatedServer, error) {
	uri := DedicatedServerURI
	if params != nil {
		queryParams, err := setQueryParams(uri, params)
		if err != nil {
			return []DedicatedServer{}, err
		}
		uri = queryParams
	}

	resp, err := api.makeRequest(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return []DedicatedServer{}, err
	}

	var result struct {
		Result []DedicatedServer `json:"result"`
	}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return []DedicatedServer{}, fmt.Errorf("error during Unmarshal: %w", err)
	}

	return result.Result, nil
}

func (api *API) DedicatedServer(ctx context.Context, serverUUID string) (DedicatedServer, error) {
	uri := fmt.Sprintf("%s/%s", DedicatedServerURI, serverUUID)

	resp, err := api.makeRequest(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return DedicatedServer{}, err
	}

	var result struct {
		Result DedicatedServer `json:"result"`
	}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return DedicatedServer{}, fmt.Errorf("error during Unmarshal: %w", err)
	}

	return result.Result, nil
}

func (api *API) CreateDedicatedServer(ctx context.Context, opts DedicatedServerCreateOpts) (DedicatedServer, error) {
	requestBody, err := json.Marshal(opts)
	if err != nil {
		return DedicatedServer{}, fmt.Errorf("error marshalling params to JSON: %w", err)
	}

	resp, err := api.makeRequest(ctx, http.MethodPost, DedicatedServerURI, requestBody)
	if err != nil {
		return DedicatedServer{}, err
	}

	var result struct {
		Result DedicatedServer `json:"result"`
	}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return DedicatedServer{}, fmt.Errorf("error during Unmarshal: %w", err)
	}

	return result.Result, nil
}

func (api *API) UpdateDedicatedServer(ctx context.Context, serverUUID string, opts DedicatedServerUpdateOpts) (DedicatedServer, error) {
	uri := fmt.Sprintf("%s/%s", DedicatedServerURI, serverUUID)

	requestBody, err := json.Marshal(opts)
	if err != nil {
		return DedicatedServer{}, fmt.Errorf("error marshalling params to JSON: %w", err)
	}

	resp, err := api.makeRequest(ctx, http.MethodPatch, uri, requestBody)
	if err != nil {
		return DedicatedServer{}, err
	}

	var result struct {
		Result DedicatedServer `json:"result"`
	}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return DedicatedServer{}, fmt.Errorf("error during Unmarshal: %w", err)
	}

	return result.Result, nil
}

func (api *API) DeleteDedicatedServer(ctx context.Context, serverUUID string) error {
	uri := fmt.Sprintf("%s/%s", DedicatedServerURI, serverUUID)

	_, err := api.makeRequest(ctx, http.MethodDelete, uri, nil)
	return err
}

// Методы для работы с локациями
func (api *API) Locations(ctx context.Context) ([]Location, error) {
	resp, err := api.makeRequest(ctx, http.MethodGet, LocationURI, nil)
	if err != nil {
		return []Location{}, err
	}

	var result struct {
		Result []Location `json:"result"`
	}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return []Location{}, fmt.Errorf("error during Unmarshal: %w", err)
	}

	return result.Result, nil
}

func (api *API) Location(ctx context.Context, locationUUID string) (Location, error) {
	locations, err := api.Locations(ctx)
	if err != nil {
		return Location{}, err
	}

	for _, location := range locations {
		if location.UUID == locationUUID {
			return location, nil
		}
	}

	return Location{}, fmt.Errorf("location with UUID %s not found", locationUUID)
}

// Методы для работы с конфигурациями
func (api *API) Configurations(ctx context.Context, locationUUID string) ([]Configuration, error) {
	uri := ConfigurationURI
	if locationUUID != "" {
		uri = fmt.Sprintf("%s?location_uuid=%s", ConfigurationURI, locationUUID)
	}

	resp, err := api.makeRequest(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return []Configuration{}, err
	}

	var result struct {
		Result []Configuration `json:"result"`
	}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return []Configuration{}, fmt.Errorf("error during Unmarshal: %w", err)
	}

	return result.Result, nil
}

func (api *API) Configuration(ctx context.Context, configUUID string) (Configuration, error) {
	configurations, err := api.Configurations(ctx, "")
	if err != nil {
		return Configuration{}, err
	}

	for _, config := range configurations {
		if config.UUID == configUUID {
			return config, nil
		}
	}

	return Configuration{}, fmt.Errorf("configuration with UUID %s not found", configUUID)
}

// Методы для работы с тарифами
func (api *API) Tariffs(ctx context.Context, configUUID string) ([]Tariff, error) {
	uri := TariffURI
	if configUUID != "" {
		uri = fmt.Sprintf("%s?configuration_uuid=%s", TariffURI, configUUID)
	}

	resp, err := api.makeRequest(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return []Tariff{}, err
	}

	var result struct {
		Result []Tariff `json:"result"`
	}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return []Tariff{}, fmt.Errorf("error during Unmarshal: %w", err)
	}

	return result.Result, nil
}

func (api *API) Tariff(ctx context.Context, tariffUUID string) (Tariff, error) {
	tariffs, err := api.Tariffs(ctx, "")
	if err != nil {
		return Tariff{}, err
	}

	for _, tariff := range tariffs {
		if tariff.UUID == tariffUUID {
			return tariff, nil
		}
	}

	return Tariff{}, fmt.Errorf("tariff with UUID %s not found", tariffUUID)
}

// Методы для работы с образами ОС
func (api *API) OSImages(ctx context.Context, locationUUID, serviceUUID string) ([]OSImage, error) {
	uri := fmt.Sprintf("%s?location_uuid=%s&service_uuid=%s", OSImageURI, locationUUID, serviceUUID)

	resp, err := api.makeRequest(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return []OSImage{}, err
	}

	var result struct {
		Result []OSImage `json:"result"`
	}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return []OSImage{}, fmt.Errorf("error during Unmarshal: %w", err)
	}

	return result.Result, nil
}

func (api *API) OSImage(ctx context.Context, osImageUUID, locationUUID, serviceUUID string) (*OSImage, error) {
	images, err := api.OSImages(ctx, locationUUID, serviceUUID)
	if err != nil {
		return &OSImage{}, err
	}

	for _, image := range images {
		if image.UUID == osImageUUID {
			return &image, nil
		}
	}

	return &OSImage{}, fmt.Errorf("OS image with UUID %s not found", osImageUUID)
}

// Методы для работы с сетями
func (api *API) Networks(ctx context.Context, locationUUID string) ([]Network, error) {
	uri := NetworkURI
	if locationUUID != "" {
		uri = fmt.Sprintf("%s?location_uuid=%s", NetworkURI, locationUUID)
	}

	resp, err := api.makeRequest(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return []Network{}, err
	}

	var result struct {
		Result []Network `json:"result"`
	}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return []Network{}, fmt.Errorf("error during Unmarshal: %w", err)
	}

	return result.Result, nil
}

func (api *API) Network(ctx context.Context, networkUUID string) (Network, error) {
	networks, err := api.Networks(ctx, "")
	if err != nil {
		return Network{}, err
	}

	for _, network := range networks {
		if network.UUID == networkUUID {
			return network, nil
		}
	}

	return Network{}, fmt.Errorf("network with UUID %s not found", networkUUID)
}

// Вспомогательные методы
func (api *API) makeRequest(ctx context.Context, method, uri string, params interface{}) ([]byte, error) {
	jsonBody, err := handleParams(params)
	if err != nil {
		return nil, err
	}

	var resp *http.Response
	var respErr error
	var reqBody io.Reader
	var respBody []byte

	if jsonBody != nil {
		reqBody = bytes.NewReader(jsonBody)
	}

	resp, respErr = api.request(ctx, method, uri, reqBody)
	if respErr != nil || resp.StatusCode >= http.StatusInternalServerError {
		if respErr == nil {
			respBody, err = io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				respErr = fmt.Errorf("could not read response body: %w", err)
			} else {
				respErr = fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
			}
		}
		return nil, respErr
	}

	respBody, err = io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("could not read response body: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, handleStatusCode(resp.StatusCode, respBody, uri)
	}

	return respBody, nil
}

func (api *API) request(ctx context.Context, method, uri string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, api.Endpoint+uri, body)
	if err != nil {
		return nil, fmt.Errorf("HTTP request creation failed: %w", err)
	}

	// Используем заголовок X-Token для авторизации как показано в документации
	req.Header.Set("User-Agent", api.UserAgent)
	req.Header.Set("X-Token", api.Token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := api.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	return resp, nil
}

func handleParams(params interface{}) ([]byte, error) {
	if params == nil {
		return nil, nil
	}

	if paramBytes, ok := params.([]byte); ok {
		return paramBytes, nil
	}

	jsonBody, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("error marshalling params to JSON: %w", err)
	}

	return jsonBody, nil
}

func handleStatusCode(statusCode int, body []byte, uri string) error {
	if statusCode >= http.StatusInternalServerError {
		return fmt.Errorf("http status %d: service failed. URI: %s Body: %s", statusCode, uri, string(body))
	}

	errBody := &DedicatedServerAPIError{}
	err := json.Unmarshal(body, &errBody)
	if err != nil {
		return fmt.Errorf("can't unmarshal response (status %d): %s, %w", statusCode, string(body), err)
	}
	return errBody
}

func setQueryParams(uri string, params interface{}) (string, error) {
	v := url.Values{}

	var queryParams map[string]interface{}
	jsonParams, err := json.Marshal(params)
	if err != nil {
		return "", fmt.Errorf("error marshalling params to JSON: %w", err)
	}

	err = json.Unmarshal(jsonParams, &queryParams)
	if err != nil {
		return "", fmt.Errorf("error during Unmarshal: %w", err)
	}

	for key, value := range queryParams {
		if value != nil && value != "" {
			v.Set(key, fmt.Sprintf("%v", value))
		}
	}

	if len(v) > 0 {
		uri = uri + "?" + v.Encode()
	}

	return uri, nil
}

func convertFieldFromStringToType(fieldValue string) interface{} {
	if val, err := strconv.Atoi(fieldValue); err == nil {
		return val
	} else if val, err := strconv.ParseFloat(fieldValue, 64); err == nil {
		return val
	} else if val, err := strconv.ParseBool(fieldValue); err == nil {
		return val
	}
	return fieldValue
}

// WaitForServerStatus ожидает определенного статуса сервера
func (api *API) WaitForServerStatus(ctx context.Context, serverUUID string, targetStatus Status, timeout time.Duration) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			return fmt.Errorf("timeout waiting for server %s to reach status %s", serverUUID, targetStatus)
		case <-ticker.C:
			server, err := api.DedicatedServer(ctx, serverUUID)
			if err != nil {
				return fmt.Errorf("error checking server status: %w", err)
			}

			if server.Status == targetStatus {
				return nil
			}

			if server.Status == StatusError {
				return fmt.Errorf("server %s entered error state", serverUUID)
			}
		}
	}
}

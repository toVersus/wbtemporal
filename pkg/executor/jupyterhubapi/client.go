// Implemetation of JupyterHub API client
// https://jupyterhub.readthedocs.io/en/stable/reference/rest-api.html#/
package jupyterhubapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

var (
	ErrResourceNotFound       = errors.New("resource not found")
	ErrResourceAlreadyExists  = errors.New("resource already exists")
	ErrResourceAlreadyRunning = errors.New("resource already running")
)

const (
	usersAPIPath       = "/hub/api/users"
	userAPIPath        = "/hub/api/users/%s"
	userServersAPIPath = "/hub/api/users/%s/servers"
	userServerAPIPath  = "/hub/api/users/%s/servers/%s"
)

type Client struct {
	baseURL string
	token   string

	client *http.Client
}

func NewClient(ctx context.Context, baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

type User struct {
	Name         string            `json:"name"`
	Admin        bool              `json:"admin"`
	Roles        []string          `json:"roles"`
	Groups       []string          `json:"groups"`
	Server       string            `json:"server,omitempty"`
	Pending      string            `json:"pending,omitempty"`
	LastActivity string            `json:"last_activity,omitempty"`
	Servers      map[string]Server `jsonpath:"server,omitempty"`
	AuthState    interface{}       `json:"auth_state,omitempty"`
}

type Server struct {
	Name         string        `json:"name"`
	Ready        bool          `json:"ready"`
	Stopped      bool          `json:"stopped"`
	Pending      PendingAction `json:"pending"`
	URL          string        `json:"url"`
	ProgressUrl  string        `json:"progress_url"`
	Started      string        `json:"started"`
	LastActivity string        `json:"last_activity"`
	State        interface{}   `json:"state,omitempty"`
	UserOptions  interface{}   `json:"user_options,omitempty"`
}

type PendingAction string

const (
	PendingActionSpawn PendingAction = "spawn"
	PendingActionStop  PendingAction = "stop"
)

const (
	ServerStatusReady   = "Ready"
	ServerStatusStopped = "Stopped"
	ServerStatusPending = "Pending"
)

func getServerStatus(server Server) string {
	var status string
	if server.Ready {
		status = ServerStatusReady
	}
	if server.Stopped {
		status = ServerStatusStopped
	}
	if server.Pending == PendingActionSpawn || server.Pending == PendingActionStop {
		status = ServerStatusPending
	}
	return status
}

type GetUserOption struct {
	User string
}

func (c *Client) GetUser(ctx context.Context, opt *GetUserOption) (User, error) {
	targetPath := fmt.Sprintf(userAPIPath, opt.User)
	targetURL, err := url.JoinPath(c.baseURL, targetPath)
	if err != nil {
		return User{}, fmt.Errorf("failed to generate target URL for JupyterHub API: %w", err)
	}
	got, err := c.doRequest(ctx, http.MethodGet, targetURL)
	if err != nil {
		return User{}, fmt.Errorf("failed to get user: %w", err)
	}
	var user User
	if err := json.Unmarshal(got, &user); err != nil {
		return User{}, fmt.Errorf("failed to unmarshal response of get user: %w", err)
	}
	return user, nil
}

type CreateUserOption struct {
	User string
}

func (c *Client) CreateUser(ctx context.Context, opt *CreateUserOption) (User, error) {
	targetPath := fmt.Sprintf(userAPIPath, opt.User)
	targetURL, err := url.JoinPath(c.baseURL, targetPath)
	if err != nil {
		return User{}, fmt.Errorf("failed to generate target URL for JupyterHub API: %w", err)
	}
	got, err := c.doRequest(ctx, http.MethodPost, targetURL)
	if err != nil {
		return User{}, fmt.Errorf("failed to create user: %w", err)
	}
	var user User
	if err := json.Unmarshal(got, &user); err != nil {
		return User{}, fmt.Errorf("failed to unmarshal response of get user: %w", err)
	}
	return user, nil
}

// type GetUserServerOption struct {
// 	Server string
// 	User   string
// }

// func (c *Client) GetUserServer(ctx context.Context, opt *GetUserServerOption) (Server, error) {
// 	targetPath := fmt.Sprintf(userServerAPIPath, opt.User, opt.Server)
// 	targetURL, err := url.JoinPath(c.baseURL, targetPath)
// 	if err != nil {
// 		return Server{}, fmt.Errorf("failed to generate target URL for JupyterHub API: %w", err)
// 	}
// 	got, err := c.doRequest(ctx, http.MethodGet, targetURL)
// 	if err != nil {
// 		return Server{}, fmt.Errorf("failed to get user server: %w", err)
// 	}
// 	var server Server
// 	if err := json.Unmarshal(got, &server); err != nil {
// 		return Server{}, fmt.Errorf("failed to unmarshal response of create user server: %w", err)
// 	}
// 	return server, nil
// }

type CreateUserServerOption struct {
	Server string
	User   string
}

func (c *Client) CreateUserServer(ctx context.Context, opt *CreateUserServerOption) error {
	targetPath := fmt.Sprintf(userServerAPIPath, opt.User, opt.Server)
	targetURL, err := url.JoinPath(c.baseURL, targetPath)
	if err != nil {
		return fmt.Errorf("failed to generate target URL for JupyterHub API: %w", err)
	}
	_, err = c.doRequest(ctx, http.MethodPost, targetURL)
	if err != nil {
		return fmt.Errorf("failed to create user server: %w", err)
	}
	// var server Server
	// if err := json.Unmarshal(got, &server); err != nil {
	// 	return Server{}, fmt.Errorf("failed to unmarshal response of create user server: %w", err)
	// }
	return nil
}

func (c *Client) doRequest(ctx context.Context, method string, targetURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for JupyterHub API: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("token %s", c.token))
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to JupyterHub API: %w", err)
	}

	switch method {
	case http.MethodGet:
		if resp.StatusCode == http.StatusNotFound {
			return nil, ErrResourceNotFound
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("GET operation failed, %s got %s status code from JupyterHub API", targetURL, resp.Status)
		}
	case http.MethodPost:
		if resp.StatusCode == http.StatusConflict {
			return nil, ErrResourceAlreadyExists
		}
		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
			return nil, fmt.Errorf("POST operation failed, %s got %s status code from JupyterHub API", targetURL, resp.Status)
		}
	case http.MethodDelete:
		if resp.StatusCode == http.StatusNotFound {
			return nil, ErrResourceNotFound
		}
		if resp.StatusCode != http.StatusNoContent {
			return nil, fmt.Errorf("DELETE operation failed, %s got %s status code from JupyterHub API", targetURL, resp.Status)
		}
	}
	defer resp.Body.Close()

	got, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from JupyterHub API: %w", err)
	}
	return got, nil
}

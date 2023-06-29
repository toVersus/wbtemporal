package jupyterhubapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	api "github.com/toVersus/wbtemporal/pkg/api/jupyterhub"
)

const (
	UserServerStatusReady   = "Ready"
	UserServerStatusPending = "Pending"
	UserServerStatusStopped = "Stopped"
)

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrServerNotFound = errors.New("server not found")
)

type notebook struct {
	*api.Client
	baseURL    string
	apiBaseURL string
}

func NewExecutor(ctx context.Context, baseURL, token string) (Executor, error) {
	apiBaseURL, err := url.JoinPath(baseURL, "/hub/api")
	if err != nil {
		return nil, fmt.Errorf("failed to generate JupyterHub API base URL: %v", err)
	}

	client, err := api.NewClient(apiBaseURL, api.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
		return nil
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to create JupyterHub API client: %w", err)
	}

	return &notebook{Client: client, baseURL: baseURL, apiBaseURL: apiBaseURL}, nil
}

func (n *notebook) GetUser(ctx context.Context, option *Option) (*api.User, error) {
	resp, err := n.GetUsersName(ctx, option.User)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrUserNotFound
	} else if resp.StatusCode > 299 && resp.StatusCode < 200 {
		return nil, fmt.Errorf("unexpected status code returned from getting user: %s", resp.Status)
	}

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	var user api.User
	if err := json.Unmarshal(result, &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
	}
	return &user, nil
}

func (n *notebook) CreateUser(ctx context.Context, option *Option) (*api.User, error) {
	resp, err := n.PostUsersName(ctx, option.User)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}
	if resp.StatusCode > 299 && resp.StatusCode < 200 {
		return nil, fmt.Errorf("unexpected status code returned from creating user: %v", resp.Status)
	}

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	var user api.User
	if err := json.Unmarshal(result, &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
	}
	return &user, nil
}

func (n *notebook) GetUserServer(ctx context.Context, option *Option) (*Status, error) {
	user, err := n.GetUser(ctx, option)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create user: %v", err)
	}

	for name, server := range *user.Servers {
		if name != option.Server {
			continue
		}
		var status string
		if *server.Ready {
			status = UserServerStatusReady
		} else if *server.Pending == api.ServerPendingSpawn || *server.Pending == api.ServerPendingStop {
			status = UserServerStatusPending
		} else if *server.Stopped {
			status = UserServerStatusStopped
		}
		serverURL, err := url.JoinPath(n.baseURL, *server.Url)
		if err != nil {
			return nil, fmt.Errorf("failed to generate server URL: %v", err)
		}

		return &Status{
			Name:   name,
			URL:    serverURL,
			Status: status,
		}, nil
	}
	return nil, ErrServerNotFound
}

func (n *notebook) CreateUserServer(ctx context.Context, option *Option) error {
	user, err := n.GetUser(ctx, option)
	if err != nil {
		return err
	}
	servers := *user.Servers
	if len(servers) != 0 {
		for _, server := range servers {
			if *server.Name == option.Server && *server.Ready {
				return nil
			}
		}
	}

	resp, err := n.PostUsersNameServersServerName(ctx, option.User, option.Server, nil)
	if err != nil {
		return fmt.Errorf("failed to create server: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 && resp.StatusCode < 200 {
		return fmt.Errorf("failed to create server: %v", resp.Status)
	}

	return nil
}

func (n *notebook) DeleteUserServer(ctx context.Context, option *Option) error {
	user, err := n.GetUser(ctx, option)
	if err != nil {
		return err
	}
	servers := *user.Servers
	if len(servers) != 0 {
		for _, server := range servers {
			if *server.Name == option.Server && *server.Stopped {
				return nil
			}
		}
	}

	resp, err := n.DeleteUsersNameServersServerName(ctx, option.User, option.Server)
	if err != nil {
		return fmt.Errorf("failed to delete server: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 && resp.StatusCode < 200 {
		return fmt.Errorf("failed to delete server: %v", resp.Status)
	}

	return nil
}

func (n *notebook) IsUserServerReady(ctx context.Context, option *Option) (bool, error) {
	user, err := n.GetUser(ctx, option)
	if err != nil {
		return false, err
	}

	for name, server := range *user.Servers {
		if name != option.Server {
			continue
		}
		return *server.Ready, nil
	}
	return false, fmt.Errorf("server %s not found", option.Server)
}

func (n *notebook) IsUserServerDeleted(ctx context.Context, option *Option) (bool, error) {
	user, err := n.GetUser(ctx, option)
	if err != nil {
		return false, err
	}

	deleted := true
	for name, _ := range *user.Servers {
		if name == option.Server {
			deleted = false
			continue
		}
	}
	return deleted, nil
}

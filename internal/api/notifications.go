package api

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	"github.com/direktiv/direktiv/internal/core"
	"github.com/go-chi/chi/v5"
)

// NOTE: We can potentially build a real notifications system if that seems useful.
// For now, this is just a port of the v1 linting API. The UI guys requested that I rename it to notifications.

type notificationsController struct {
	sManager core.SecretsManager
}

func (c *notificationsController) mountRouter(r chi.Router) {
	r.Get("/", c.list)
}

type apiNotification struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Count       *int   `json:"count,omitempty"`
	Level       string `json:"level"`
}

func (c *notificationsController) list(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")
	ctx := r.Context()

	secretIssues, err := c.lintSecrets(ctx, namespace)
	if err != nil {
		writeInternalError(w, err)
		return
	}

	notifications := make([]*apiNotification, 0)
	if len(secretIssues) > 0 {
		notifications = append(notifications, secretIssues...)
	}

	writeJSON(w, notifications)
}

func (c *notificationsController) lintSecrets(ctx context.Context, ns string) ([]*apiNotification, error) {
	secrets, err := c.sManager.GetAll(ctx, ns)
	if err != nil {
		return nil, err
	}

	issues := make([]*apiNotification, 0)
	keys := []string{}

	for _, secret := range secrets {
		if len(secret.Data) == 0 {
			keys = append(keys, secret.Name)
		}
	}

	if len(keys) == 0 {
		return nil, err
	}

	sort.Strings(keys)

	count := len(keys)

	issues = append(issues, &apiNotification{
		Level:       "warning",
		Type:        "uninitialized_secrets",
		Description: fmt.Sprintf(`secrets have not been initialized: %v`, keys),
		Count:       &count,
	})

	return issues, nil
}

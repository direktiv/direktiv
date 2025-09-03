package api

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	"github.com/direktiv/direktiv/internal/datastore/datasql"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// NOTE: We can potentially build a real notifications system if that seems useful.
// For now, this is just a port of the v1 linting API. The UI guys requested that I rename it to notifications.

type notificationsController struct {
	db *gorm.DB
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

	notifications := make([]*apiNotification, 0)

	ctx := r.Context()

	db := c.db.WithContext(ctx).Begin()
	if db.Error != nil {
		writeInternalError(w, db.Error)
		return
	}
	defer db.Rollback()

	secretIssues, err := c.lintSecrets(ctx, db, namespace)
	if err != nil {
		writeInternalError(w, err)
		return
	}

	if len(secretIssues) > 0 {
		notifications = append(notifications, secretIssues...)
	}

	writeJSON(w, notifications)
}

func (c *notificationsController) lintSecrets(ctx context.Context, tx *gorm.DB, ns string) ([]*apiNotification, error) {
	secrets, err := datasql.NewStore(tx).Secrets().GetAll(ctx, ns)
	if err != nil {
		return nil, err
	}

	issues := make([]*apiNotification, 0)
	keys := []string{}

	for _, secret := range secrets {
		if secret.Data == nil {
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

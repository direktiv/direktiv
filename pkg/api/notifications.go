package api

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/go-chi/chi/v5"
)

// NOTE: We can potentially build a real notifications system if that seems useful.
// For now, this is just a port of the v1 linting API. The UI guys requested that I rename it to notifications.

type notificationsController struct {
	db *database.SQLStore
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
	ns := extractContextNamespace(r)

	notifications := make([]*apiNotification, 0)

	ctx := r.Context()

	db, err := c.db.BeginTx(ctx)
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Rollback()

	secretIssues, err := c.lintSecrets(ctx, db, ns)
	if err != nil {
		writeInternalError(w, err)
		return
	}

	if len(secretIssues) > 0 {
		notifications = append(notifications, secretIssues...)
	}

	writeJSON(w, notifications)
}

func (c *notificationsController) lintSecrets(ctx context.Context, tx *database.SQLStore, ns *datastore.Namespace) ([]*apiNotification, error) {
	secrets, err := tx.DataStore().Secrets().GetAll(ctx, ns.Name)
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

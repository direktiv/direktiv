package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"sort"

	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/direktiv/direktiv/pkg/flow"
	"github.com/direktiv/direktiv/pkg/secrets"
	"github.com/go-chi/chi/v5"
)

type vaultsController struct {
	db *database.DB
}

func (c *vaultsController) mountDefaultRouter(r chi.Router) {
	r.Get("/", c.getDefault)
	r.Post("/", c.setDefault)
}

func (c *vaultsController) getConfigsWithDefault(w http.ResponseWriter, r *http.Request) *secrets.Config {
	ctx := r.Context()
	namespace := extractContextNamespace(r)

	var config *secrets.Config

	configs, err := c.db.DataStore().SecretsConfigs().Get(ctx, namespace.Name)
	if err != nil {
		if !errors.Is(err, datastore.ErrNotFound) {
			writeDataStoreError(w, err)

			return nil
		}

		config = flow.DefaultSecretsConfig(namespace.Name)
	} else {
		if err := json.Unmarshal(configs.Configuration, &config); err != nil {
			writeInternalError(w, err)

			return nil
		}
	}

	return config
}

func (c *vaultsController) getDefault(w http.ResponseWriter, r *http.Request) {
	config := c.getConfigsWithDefault(w, r)
	if config == nil {
		return
	}

	writeJSON(w, map[string]string{
		"default": config.DefaultSource,
	})
}

type setDefaultRequest struct {
	Default string `json:"default"`
}

func (c *vaultsController) setDefault(w http.ResponseWriter, r *http.Request) {
	config := c.getConfigsWithDefault(w, r)
	if config == nil {
		return
	}

	// read setting from request body

	var input setDefaultRequest

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&input); err != nil {
		writeBadrequestError(w, err)

		return
	}

	// confirm that setting matches defined config

	var match bool

	for _, x := range config.SourceConfigs {
		if x.Name == input.Default {
			match = true

			break
		}
	}

	if !match {
		writeDataStoreError(w, datastore.ErrNotFound)

		return
	}

	config.DefaultSource = input.Default

	// write back to the database

	ctx := r.Context()
	namespace := extractContextNamespace(r)

	configData, err := json.Marshal(config)
	if err != nil {
		writeInternalError(w, err)

		return
	}

	if err := c.db.DataStore().SecretsConfigs().Set(ctx, &datastore.SecretsConfigs{
		Namespace:     namespace.Name,
		Configuration: configData,
	}); err != nil {
		writeDataStoreError(w, err)

		return
	}

	writeJSON(w, map[string]string{
		"default": config.DefaultSource,
	})
}

func (c *vaultsController) mountConfigsRouter(r chi.Router) {
	r.Get("/", c.listConfigs)
	r.Post("/", c.createConfigs)

	r.Get("/{vault}", c.getConfig)
	r.Delete("/{vault}", c.deleteConfig)
	r.Patch("/{vault}", c.patchConfig)
}

func (c *vaultsController) listConfigs(w http.ResponseWriter, r *http.Request) {
	config := c.getConfigsWithDefault(w, r)
	if config == nil {
		return
	}

	list := make([]map[string]string, len(config.SourceConfigs))

	for i, x := range config.SourceConfigs {
		list[i] = map[string]string{
			"name":   x.Name,
			"driver": x.Driver,
		}
	}

	writeJSON(w, list)
}

func (c *vaultsController) createConfigs(w http.ResponseWriter, r *http.Request) {
	config := c.getConfigsWithDefault(w, r)
	if config == nil {
		return
	}

	var sourceConfigs secrets.SourceConfigs

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&sourceConfigs); err != nil {
		writeBadrequestError(w, err)

		return
	}

	// validate

	for idx, sourceConfig := range sourceConfigs {
		if sourceConfig.Name == "" {
			writeBadrequestError(w, fmt.Errorf(`validation error on config [%d]: the vault must have a name`, idx))

			return
		}

		vault := sourceConfig.Name

		if vault == "local" {
			writeBadrequestError(w, fmt.Errorf(`validation error on config [%d]: the "local" vault cannot be overwritten`, idx))

			return
		}

		if sourceConfig.Driver == "" {
			writeBadrequestError(w, fmt.Errorf(`validation error on config [%d]: the vault must have a driver`, idx))

			return
		}

		driver, err := secrets.GetDriver(sourceConfig.Driver)
		if err != nil {
			writeBadrequestError(w, fmt.Errorf(`validation error on config [%d]: error resolving driver: %w`, idx, err))

			return
		}

		if err := driver.ValidateConfig(sourceConfig.Data); err != nil {
			writeBadrequestError(w, fmt.Errorf(`validation error on config [%d]: driver '%s' reports validation error on source config data: %w`, idx, sourceConfig.Driver, err))

			return
		}

		// check for duplicates

		found := -1

		for i := range config.SourceConfigs {
			if config.SourceConfigs[i].Name == vault {
				found = i

				break
			}
		}

		if found >= 0 {
			writeDataStoreError(w, fmt.Errorf(`validation error on config [%d]: %w`, idx, datastore.ErrDuplication))

			return
		}

		config.SourceConfigs = append(config.SourceConfigs, sourceConfig)
	}

	sort.Sort(config.SourceConfigs)

	// write back to the database

	ctx := r.Context()
	namespace := extractContextNamespace(r)

	configData, err := json.Marshal(config)
	if err != nil {
		writeInternalError(w, err)

		return
	}

	if err := c.db.DataStore().SecretsConfigs().Set(ctx, &datastore.SecretsConfigs{
		Namespace:     namespace.Name,
		Configuration: configData,
	}); err != nil {
		writeDataStoreError(w, err)

		return
	}

	writeOk(w)
}

func (c *vaultsController) getConfig(w http.ResponseWriter, r *http.Request) {
	config := c.getConfigsWithDefault(w, r)
	if config == nil {
		return
	}

	vault := chi.URLParam(r, "vault")

	found := -1

	for i := range config.SourceConfigs {
		if config.SourceConfigs[i].Name == vault {
			found = i

			break
		}
	}

	if found < 0 {
		writeDataStoreError(w, datastore.ErrNotFound)

		return
	}

	x := config.SourceConfigs[found]

	driver, err := secrets.GetDriver(x.Driver)
	if err != nil {
		writeInternalError(w, err)

		return
	}

	redacted, err := driver.RedactConfig(x.Data)
	if err != nil {
		writeInternalError(w, err)

		return
	}

	x.Data = redacted

	writeJSON(w, x)
}

func (c *vaultsController) deleteConfig(w http.ResponseWriter, r *http.Request) {
	config := c.getConfigsWithDefault(w, r)
	if config == nil {
		return
	}

	vault := chi.URLParam(r, "vault")

	if vault == "local" {
		writeBadrequestError(w, errors.New(`the "local" vault cannot be deleted`))

		return
	}

	found := -1

	for i := range config.SourceConfigs {
		if config.SourceConfigs[i].Name == vault {
			found = i

			break
		}
	}

	if found < 0 {
		writeDataStoreError(w, datastore.ErrNotFound)

		return
	}

	config.SourceConfigs = slices.Delete(config.SourceConfigs, found, found+1)

	if config.DefaultSource == vault {
		config.DefaultSource = "local"
	}

	// write back to the database

	ctx := r.Context()
	namespace := extractContextNamespace(r)

	configData, err := json.Marshal(config)
	if err != nil {
		writeInternalError(w, err)

		return
	}

	if err := c.db.DataStore().SecretsConfigs().Set(ctx, &datastore.SecretsConfigs{
		Namespace:     namespace.Name,
		Configuration: configData,
	}); err != nil {
		writeDataStoreError(w, err)

		return
	}

	writeOk(w)

	secrets.DeleteController(namespace.Name) // changing the source configs should cause a cache invalidation for the secrets controller. Since this should be a rare operation, I don't worry about performance impact, and I just wipe the entire controller and let it rebuild
}

func (c *vaultsController) patchConfig(w http.ResponseWriter, r *http.Request) {
	config := c.getConfigsWithDefault(w, r)
	if config == nil {
		return
	}

	vault := chi.URLParam(r, "vault")

	if vault == "local" {
		writeBadrequestError(w, errors.New(`the "local" vault cannot be patched`))

		return
	}

	found := -1

	for i := range config.SourceConfigs {
		if config.SourceConfigs[i].Name == vault {
			found = i

			break
		}
	}

	if found < 0 {
		writeDataStoreError(w, datastore.ErrNotFound)

		return
	}

	config.SourceConfigs = slices.Delete(config.SourceConfigs, found, found+1)

	var sourceConfig secrets.SourceConfig

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&sourceConfig); err != nil {
		writeBadrequestError(w, err)

		return
	}

	// validate

	if sourceConfig.Name == "" {
		writeBadrequestError(w, errors.New(`the vault must have a name`))

		return
	}

	if sourceConfig.Name == "local" {
		writeBadrequestError(w, errors.New(`the "local" vault cannot be overwritten`))

		return
	}

	if sourceConfig.Driver == "" {
		writeBadrequestError(w, errors.New(`the vault must have a driver`))

		return
	}

	driver, err := secrets.GetDriver(sourceConfig.Driver)
	if err != nil {
		writeBadrequestError(w, fmt.Errorf(`error resolving driver: %w`, err))

		return
	}

	if err := driver.ValidateConfig(sourceConfig.Data); err != nil {
		writeBadrequestError(w, fmt.Errorf(`driver '%s' reports validation error on source config data: %w`, sourceConfig.Driver, err))

		return
	}

	// check for duplicates

	found = -1

	for i := range config.SourceConfigs {
		if config.SourceConfigs[i].Name == vault {
			found = i

			break
		}
	}

	if found >= 0 {
		writeDataStoreError(w, datastore.ErrDuplication)

		return
	}

	config.SourceConfigs = append(config.SourceConfigs, sourceConfig)

	sort.Sort(config.SourceConfigs)

	if config.DefaultSource == vault {
		config.DefaultSource = sourceConfig.Name
	}

	// write back to the database

	ctx := r.Context()
	namespace := extractContextNamespace(r)

	configData, err := json.Marshal(config)
	if err != nil {
		writeInternalError(w, err)

		return
	}

	if err := c.db.DataStore().SecretsConfigs().Set(ctx, &datastore.SecretsConfigs{
		Namespace:     namespace.Name,
		Configuration: configData,
	}); err != nil {
		writeDataStoreError(w, err)

		return
	}

	writeOk(w)

	secrets.DeleteController(namespace.Name) // changing the source configs should cause a cache invalidation for the secrets controller. Since this should be a rare operation, I don't worry about performance impact, and I just wipe the entire controller and let it rebuild
}

func (c *vaultsController) mountCacheRouter(r chi.Router) {
	r.Get("/", c.listCache)
	r.Delete("/", c.clearCache)
	r.Post("/", c.testCache)

	r.Delete("/{id}", c.invalidateCachedSecret)
}

func (c *vaultsController) listCache(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	namespace := extractContextNamespace(r)

	controller, err := secrets.GetController(namespace.Name)
	if err != nil {
		writeDataStoreError(w, err)

		return
	}

	list, err := controller.List(ctx)
	if err != nil {
		writeInternalError(w, err)

		return
	}

	responseList := []any{}
	for _, entry := range list {
		x := map[string]string{
			"id":     entry.ID,
			"path":   entry.Path,
			"source": entry.Source,
		}

		if entry.Error != nil {
			x["error"] = entry.Error.Error()
		}

		responseList = append(responseList, x)
	}

	writeJSON(w, responseList)
}

func (c *vaultsController) clearCache(w http.ResponseWriter, r *http.Request) {
	namespace := extractContextNamespace(r)

	secrets.DeleteController(namespace.Name)

	writeOk(w)
}

func (c *vaultsController) testCache(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	namespace := extractContextNamespace(r)

	controller, err := secrets.GetController(namespace.Name)
	if err != nil {
		writeDataStoreError(w, err)

		return
	}

	var refs []secrets.SecretRef

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&refs); err != nil {
		writeBadrequestError(w, err)

		return
	}

	list, err := controller.Lookup(ctx, refs)
	if err != nil {
		writeInternalError(w, err)

		return
	}

	// TODO: alan, make this a force-lookup? So that it doesn't return a cached error

	responseList := []any{}
	for _, entry := range list {
		x := map[string]string{
			"id":     entry.ID,
			"path":   entry.Path,
			"source": entry.Source,
		}

		if entry.Error != nil {
			x["error"] = entry.Error.Error()
		}

		responseList = append(responseList, x)
	}

	writeJSON(w, responseList)
}

func (c *vaultsController) invalidateCachedSecret(w http.ResponseWriter, r *http.Request) {
	namespace := extractContextNamespace(r)

	// NOTE: Since this should be a rare operation, I don't worry about performance impact, and I just wipe the entire controller and let it rebuild
	// TODO: alan, make this invalidate just a single secret instead of the whole cache?
	// TODO: alan, or maybe this API just shouldn't exist...

	secrets.DeleteController(namespace.Name)

	writeOk(w)
}

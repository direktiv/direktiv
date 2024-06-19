package tracing

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/engine"
)

func getSlogAttributesFromActionContext(tracingAttr engine.ActionContext, additionalTags ...map[string]interface{}) []interface{} {
	var result []interface{}

	// Add attributes from the ActionContext
	v := reflect.ValueOf(tracingAttr)
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		fieldName := t.Field(i).Name
		fieldValue := v.Field(i).Interface()

		if fieldValue == reflect.Zero(v.Field(i).Type()).Interface() {
			continue // Skip empty fields
		}

		switch fieldValue.(type) {
		case string, int: // Only add string and int fields
			result = append(result, strings.ToLower(fieldName), fieldValue)
		}
	}

	// Add additional tags
	for _, tags := range additionalTags {
		for k, v := range tags {
			result = append(result, k, v)
		}
	}

	return result
}

func NewGetSlogAttributesWithError(tracingAttr engine.ActionContext, err error) []interface{} {
	return getSlogAttributesFromActionContext(tracingAttr,
		map[string]interface{}{"error": err, "status": core.LogErrStatus})
}

func NewGetSlogAttributesWithStatus(tracingAttr engine.ActionContext, status core.LogStatus) []interface{} {
	return getSlogAttributesFromActionContext(tracingAttr, map[string]interface{}{"status": status})
}

func NewGetSlogAttributesWithInstanceTrackWithError(tracingAttr engine.ActionContext, err error) []interface{} {
	return getSlogAttributesFromActionContext(tracingAttr,
		map[string]interface{}{
			"error":  err,
			"status": core.LogErrStatus,
			"track":  BuildInstanceTrackViaCallpath(tracingAttr.Callpath),
		})
}

func NewGetSlogAttributesWithNamepaceTrackWithStatus(tracingAttr engine.ActionContext, status core.LogStatus) []interface{} {
	return getSlogAttributesFromActionContext(tracingAttr,
		map[string]interface{}{
			"status": status,
			"track":  BuildNamespaceTrack(tracingAttr.Namespace),
		})
}

func BuildInstanceTrackViaCallpath(callpath string) string {
	return fmt.Sprintf("%v.%v", "instance", callpath)
}

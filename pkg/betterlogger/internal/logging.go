package internal

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// AddEvent adds an event to the active trace span if tracing is enabled.
func AddEvent(ctx context.Context, msg string, attrs ...slog.Attr) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		otelAttrs := make([]attribute.KeyValue, len(attrs))
		for i, attr := range attrs {
			otelAttrs[i] = attribute.KeyValue{
				Key:   attribute.Key(attr.Key),
				Value: attribute.StringValue(attr.Value.String()),
			}
		}
		span.AddEvent(msg, trace.WithAttributes(otelAttrs...))
	}
}

// LogWithTraceAndAttributes is an internal helper to call slog with consistent formatting.
func LogWithTraceAndAttributes(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	// Add tracing information
	attrs = AddTraceID(ctx, attrs)

	// Filter unsupported object types
	attrs = RemoveObjectsFromAttrs(attrs, nil)

	// Log with structured attributes
	slog.LogAttrs(ctx, level, msg, attrs...)

	// Add event to trace span
	AddEvent(ctx, msg, attrs...)
}

// AddTraceID adds trace and span IDs to the logging attributes.
func AddTraceID(ctx context.Context, attrs []slog.Attr) []slog.Attr {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().TraceID().IsValid() {
		attrs = append(attrs,
			slog.String("traceID", span.SpanContext().TraceID().String()),
			slog.String("spanID", span.SpanContext().SpanID().String()),
		)
	}

	return attrs
}

// RemoveObjectsFromAttrs filters attributes, converting unsupported types into strings to prevent crashes.
// Parameters:
// - attrs: The original list of attributes.
// - attrsNew: The list of already-processed attributes to append to.
// Returns:
// - A new list of attributes with unsupported types safely converted.
func RemoveObjectsFromAttrs(attrs []slog.Attr, attrsNew []slog.Attr) []slog.Attr {
	for _, attr := range attrs {
		var resolvedValue any

		// Resolve the value using reflection
		value := attr.Value.Any()
		valueType := reflect.TypeOf(value)

		if valueType != nil {
			switch valueType.Kind() {
			case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Float32, reflect.Float64, reflect.Bool:
				// Retain valid types as-is
				resolvedValue = value

			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				// Convert unsigned integers to their string representation
				resolvedValue = fmt.Sprint(value)

			case reflect.Uintptr, reflect.Complex64, reflect.Complex128, reflect.Array,
				reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice, reflect.Struct, reflect.UnsafePointer:
				// Convert unsupported or complex types to strings
				resolvedValue = fmt.Sprint(value)

			case reflect.Invalid:
				// Handle invalid kind explicitly
				resolvedValue = "invalid"

			default:
				// Catch-all for unexpected kinds
				resolvedValue = fmt.Sprintf("unknown kind: %v", valueType.Kind())
			}
		} else {
			// Handle nil or untyped values
			resolvedValue = "nil"
		}

		// Add the resolved attribute
		attrsNew = append(attrsNew, slog.Any(attr.Key, resolvedValue))
	}

	return attrsNew
}

// MergeAttributes combines a set of base attributes with additional attributes efficiently.
func MergeAttributes(base []slog.Attr, extra ...slog.Attr) []slog.Attr {
	// Allocate the slice once to reduce allocations during appends.
	attrs := make([]slog.Attr, len(base)+len(extra))
	copy(attrs, base)
	copy(attrs[len(base):], extra)

	return attrs
}

// // buildInstanceAttributes constructs the base attributes for instance logging.
// func buildInstanceAttributes(attr InstanceAttributes, additionalAttrs ...slog.Attr) []slog.Attr {
// 	baseAttrs := []slog.Attr{
// 		slog.String("namespace", attr.Namespace),
// 		slog.String("instance", attr.InstanceID),
// 		slog.String("workflow", attr.WorkflowPath),
// 		slog.String("track", fmt.Sprintf("%v.%v", "instance", attr.CallPath)),
// 	}

// 	return internal.MergeAttributes(baseAttrs, additionalAttrs...)
// }

// // buildInstanceMemoryAttributes constructs the base attributes for instance memory logging.
// func buildInstanceMemoryAttributes(attr InstanceMemoryAttributes, additionalAttrs ...slog.Attr) []slog.Attr {
// 	baseAttrs := buildInstanceAttributes(InstanceAttributes{
// 		coreInstanceAttr:        attr.coreInstanceAttr,
// 		coreNamespaceAttributes: attr.coreNamespaceAttributes,
// 	})
// 	memoryAttr := slog.String("state", attr.State)

// 	return internal.MergeAttributes(baseAttrs, append(additionalAttrs, memoryAttr)...)
// }

// // buildCloudEventAttributes constructs the base attributes for cloud event logging.
// func buildCloudEventAttributes(attr CloudEventBusAttributes, additionalAttrs ...slog.Attr) []slog.Attr {
// 	baseAttrs := []slog.Attr{
// 		slog.String("event_id", attr.EventID),
// 		slog.String("source", attr.Source),
// 		slog.String("subject", attr.Subject),
// 		slog.String("event_type", attr.EventType),
// 		slog.String("namespace", attr.Namespace),
// 		slog.String("track", fmt.Sprintf("%v.%v", "namespace", attr.Namespace)),
// 	}

// 	return internal.MergeAttributes(baseAttrs, additionalAttrs...)
// }

// // buildGatewayAttributes constructs the base attributes for gateway logging.
// func buildGatewayAttributes(attr GatewayAttributes, additionalAttrs ...slog.Attr) []slog.Attr {
// 	baseAttrs := []slog.Attr{
// 		slog.String("component", "gateway"),
// 		slog.String("gateway_plugin", attr.Plugin),
// 		slog.String("namespace", attr.Namespace),
// 		slog.String("route", attr.Route),
// 		slog.String("track", fmt.Sprintf("%v.%v", "route", attr.Route)),
// 	}

// 	return internal.MergeAttributes(baseAttrs, additionalAttrs...)
// }

// // buildGatewayAttributes constructs the base attributes for gateway logging.
// func buildNamespaceAttributes(namespace string, additionalAttrs ...slog.Attr) []slog.Attr {
// 	baseAttrs := []slog.Attr{
// 		slog.String("namespace", namespace),
// 		slog.String("track", fmt.Sprintf("%v.%v", "namespace", namespace)),
// 	}

// 	return internal.MergeAttributes(baseAttrs, additionalAttrs...)
// }

// // buildSyncAttributes constructs the base attributes for mirror logging.
// func buildSyncAttributes(attr SyncAttributes, additionalAttrs ...slog.Attr) []slog.Attr {
// 	baseAttrs := []slog.Attr{
// 		slog.String("activity", attr.SyncID),
// 		slog.String("namespace", attr.Namespace),
// 		slog.String("track", fmt.Sprintf("%v.%v", "activity", attr.SyncID)),
// 	}

// 	return internal.MergeAttributes(baseAttrs, additionalAttrs...)
// }

package instancestoresql

import (
	"fmt"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/instancestore"
)

func wheres(clauses ...string) string {
	if len(clauses) == 0 {
		return ""
	}
	if len(clauses) == 1 {
		return ` WHERE ` + clauses[0]
	}

	return ` WHERE (` + strings.Join(clauses, ") AND (") + `)`
}

func generateGetInstancesOrderings(opts *instancestore.ListOpts) (string, error) {
	if opts == nil || len(opts.Orders) == 0 {
		return ` ORDER BY ` + fieldCreatedAt + " " + desc, nil
	}

	keys := make(map[string]bool)
	orderStrings := []string{}
	for _, order := range opts.Orders {
		var s string
		switch order.Field {
		case instancestore.FieldCreatedAt:
			s = fieldCreatedAt
		default:
			return "", fmt.Errorf("order field '%s': %w", order.Field, instancestore.ErrBadListOpts)
		}

		if _, exists := keys[order.Field]; exists {
			return "", fmt.Errorf("duplicate order field '%s': %w", order.Field, instancestore.ErrBadListOpts)
		}

		keys[order.Field] = true

		if order.Descending {
			s += " " + desc
		}

		orderStrings = append(orderStrings, s)
	}

	return ` ORDER BY ` + strings.Join(orderStrings, ", "), nil
}

//nolint:gocognit,goconst
func generateGetInstancesFilters(opts *instancestore.ListOpts) ([]string, []interface{}, error) {
	if opts == nil {
		return []string{}, []interface{}{}, nil
	}

	clauses := []string{}
	vals := []interface{}{}
	for idx := range opts.Filters {
		filter := opts.Filters[idx]
		var clause string
		var val interface{}
		switch filter.Field {
		case fieldNamespaceID:
			if filter.Kind == instancestore.FilterKindMatch {
				clause = fieldNamespaceID + " = ?"
				val = filter.Value
			} else {
				return nil, nil, fmt.Errorf("filter kind '%s' for use with field '%s': %w", filter.Kind, filter.Field, instancestore.ErrBadListOpts)
			}
		case instancestore.FieldCreatedAt:
			if t, ok := filter.Value.(time.Time); ok {
				filter.Value = t.UTC()
			}

			if t, ok := filter.Value.(*time.Time); ok {
				filter.Value = t.UTC()
			}

			if filter.Kind == instancestore.FilterKindBefore {
				clause = fieldCreatedAt + " < ?"
				val = filter.Value
			} else if filter.Kind == instancestore.FilterKindAfter {
				clause = fieldCreatedAt + " > ?"
				val = filter.Value
			} else {
				return nil, nil, fmt.Errorf("filter kind '%s' for use with field '%s': %w", filter.Kind, filter.Field, instancestore.ErrBadListOpts)
			}

		case fieldDeadline:
			if t, ok := filter.Value.(time.Time); ok {
				filter.Value = t.UTC()
			}

			if t, ok := filter.Value.(*time.Time); ok {
				filter.Value = t.UTC()
			}

			if filter.Kind == instancestore.FilterKindBefore {
				clause = fieldDeadline + " < ?"
				val = filter.Value
			} else if filter.Kind == instancestore.FilterKindAfter {
				clause = fieldDeadline + " > ?"
				val = filter.Value
			} else {
				return nil, nil, fmt.Errorf("filter kind '%s' for use with field '%s': %w", filter.Kind, filter.Field, instancestore.ErrBadListOpts)
			}

		case instancestore.FieldWorkflowPath:
			if filter.Kind == instancestore.FilterKindMatch {
				clause = fieldWorkflowPath + " = ?"
				val = fmt.Sprintf("%s", filter.Value)
			} else if filter.Kind == instancestore.FilterKindPrefix {
				clause = fieldWorkflowPath + " LIKE ?"
				val = fmt.Sprintf("%s", filter.Value) + "%"
			} else if filter.Kind == instancestore.FilterKindContains {
				clause = fieldWorkflowPath + " LIKE ?"
				val = "%" + fmt.Sprintf("%s", filter.Value) + "%"
			} else {
				return nil, nil, fmt.Errorf("filter kind '%s' for use with field '%s': %w", filter.Kind, filter.Field, instancestore.ErrBadListOpts)
			}

		case instancestore.FieldStatus:
			if filter.Kind == instancestore.FilterKindMatch {
				clause = fieldStatus + " = ?"
				val = filter.Value
			} else if filter.Kind == "<" {
				clause = fieldStatus + " < ?"
				val = filter.Value
			} else {
				return nil, nil, fmt.Errorf("filter kind '%s' for use with field '%s': %w", filter.Kind, filter.Field, instancestore.ErrBadListOpts)
			}

		case instancestore.FieldInvoker:
			if filter.Kind == instancestore.FilterKindMatch {
				clause = fieldInvoker + " = ?"
				val = fmt.Sprintf("%s", filter.Value)
			} else if filter.Kind == instancestore.FilterKindContains {
				clause = fieldInvoker + " LIKE ?"
				val = "%" + fmt.Sprintf("%s", filter.Value) + "%"
			} else {
				return nil, nil, fmt.Errorf("filter kind '%s' for use with field '%s': %w", filter.Kind, filter.Field, instancestore.ErrBadListOpts)
			}

		default:
			return nil, nil, fmt.Errorf("filter field '%s': %w", filter.Field, instancestore.ErrBadListOpts)
		}

		clauses = append(clauses, clause)
		vals = append(vals, val)
	}

	return clauses, vals, nil
}

func generateInsertQuery(columns []string) string {
	into := strings.Join(columns, ", ")
	valPlaceholders := strings.Repeat("?, ", len(columns)-1) + "?"
	query := fmt.Sprintf(`INSERT INTO %s(%s) VALUES (%s)`, table, into, valPlaceholders)

	return query
}

func generateGetInstancesQueries(columns []string, opts *instancestore.ListOpts) (string, string, []interface{}, error) {
	clauses, vals, err := generateGetInstancesFilters(opts)
	if err != nil {
		return "", "", nil, err
	}

	orderings, err := generateGetInstancesOrderings(opts)
	if err != nil {
		return "", "", nil, err
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM %s`, table)
	countQuery += wheres(clauses...)

	query := fmt.Sprintf(`SELECT %s FROM %s`, strings.Join(columns, ", "), table)
	query += wheres(clauses...)
	query += orderings

	if opts != nil && opts.Limit > 0 {
		query += fmt.Sprintf(` LIMIT %d`, opts.Limit)

		if opts.Offset > 0 {
			query += fmt.Sprintf(` OFFSET %d`, opts.Offset)
		}
	}

	return countQuery, query, vals, nil
}

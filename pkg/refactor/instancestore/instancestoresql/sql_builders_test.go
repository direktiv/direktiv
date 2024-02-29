//nolint:testpackage
package instancestoresql

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/instancestore"
)

func Test_wheres(t *testing.T) {
	res := wheres()

	if res != `` {
		t.Errorf("wheres failed with no inputs: expected '%s', but got '%s'", "", res)
	}

	res = wheres(`A = ?`)
	expect := ` WHERE A = ?`
	if res != expect {
		t.Errorf("wheres failed with one input: expected '%s', but got '%s'", expect, res)
	}

	res = wheres(
		`A = ?`,
		`B > ?`,
	)
	expect = ` WHERE (A = ?) AND (B > ?)`
	if res != expect {
		t.Errorf("wheres failed with multiple inputs: expected '%s', but got '%s'", expect, res)
	}
}

func Test_generateGetInstancesOrderings(t *testing.T) {
	res, err := generateGetInstancesOrderings(nil)
	if err != nil {
		t.Error(err)
	}
	expect := ` ORDER BY created_at desc`

	if res != expect {
		t.Errorf("generateGetInstancesOrderings failed with nil opts: expected '%s', but got '%s'", expect, res)
	}

	res, err = generateGetInstancesOrderings(&instancestore.ListOpts{
		Orders: []instancestore.Order{},
	})
	if err != nil {
		t.Error(err)
	}

	if res != expect {
		t.Errorf("generateGetInstancesOrderings failed with zero orderings: expected '%s', but got '%s'", expect, res)
	}

	res, err = generateGetInstancesOrderings(&instancestore.ListOpts{
		Orders: []instancestore.Order{
			{Field: instancestore.FieldCreatedAt, Descending: false},
		},
	})
	if err != nil {
		t.Error(err)
	}
	expect = ` ORDER BY created_at`

	if res != expect {
		t.Errorf("generateGetInstancesOrderings failed with one valid orderings: expected '%s', but got '%s'", expect, res)
	}

	res, err = generateGetInstancesOrderings(&instancestore.ListOpts{
		Orders: []instancestore.Order{
			{Field: instancestore.FieldCreatedAt, Descending: true},
		},
	})
	if err != nil {
		t.Error(err)
	}
	expect = ` ORDER BY created_at desc`

	if res != expect {
		t.Errorf("generateGetInstancesOrderings failed with one valid orderings: expected '%s', but got '%s'", expect, res)
	}

	_, err = generateGetInstancesOrderings(&instancestore.ListOpts{
		Orders: []instancestore.Order{
			{Field: instancestore.FieldCreatedAt, Descending: false},
			{Field: instancestore.FieldCreatedAt, Descending: true},
		},
	})
	if !errors.Is(err, instancestore.ErrBadListOpts) {
		t.Errorf("generateGetInstancesOrderings returned unexpected error: expected error is '%v', but got '%v'", instancestore.ErrBadListOpts, err)
	}

	_, err = generateGetInstancesOrderings(&instancestore.ListOpts{
		Orders: []instancestore.Order{
			{Field: "STATUS", Descending: false},
		},
	})
	if !errors.Is(err, instancestore.ErrBadListOpts) {
		t.Errorf("generateGetInstancesOrderings returned unexpected error: expected error is '%v', but got '%v'", instancestore.ErrBadListOpts, err)
	}
}

type generateGetInstancesFiltersResults struct {
	clauses []string
	vals    []interface{}
}

func (expect *generateGetInstancesFiltersResults) compare(clauses []string, vals []interface{}) error {
	if len(clauses) != len(expect.clauses) {
		return fmt.Errorf("expected clauses '%v', but got '%v'", expect.clauses, clauses)
	}

	for idx := range clauses {
		if clauses[idx] != expect.clauses[idx] {
			return fmt.Errorf("expected clauses '%v', but got '%v'", expect.clauses, clauses)
		}
	}

	if len(vals) != len(expect.vals) {
		return fmt.Errorf("expected vals '%v', but got '%v'", expect.vals, vals)
	}

	for idx := range vals {
		if tv, ok := vals[idx].(time.Time); ok {
			te, _ := expect.vals[idx].(time.Time)
			if !te.Equal(tv) {
				return fmt.Errorf("expected vals '%v', but got '%v'", expect.vals, vals)
			}
		} else if vals[idx] != expect.vals[idx] {
			return fmt.Errorf("expected vals '%v', but got '%v'", expect.vals, vals)
		}
	}

	return nil
}

func Test_generateGetInstancesFilters(t *testing.T) {
	clauses, vals, err := generateGetInstancesFilters(nil)
	if err != nil {
		t.Error(err)
	}

	expect := &generateGetInstancesFiltersResults{
		clauses: []string{},
		vals:    []interface{}{},
	}

	err = expect.compare(clauses, vals)
	if err != nil {
		t.Errorf("generateGetInstancesFilters failed with nil opts: %v", err)
	}

	clauses, vals, err = generateGetInstancesFilters(&instancestore.ListOpts{})
	if err != nil {
		t.Error(err)
	}

	err = expect.compare(clauses, vals)
	if err != nil {
		t.Errorf("generateGetInstancesFilters failed with zero filters: %v", err)
	}

	clauses, vals, err = generateGetInstancesFilters(&instancestore.ListOpts{
		Filters: []instancestore.Filter{
			{Field: fieldNamespaceID, Kind: instancestore.FilterKindMatch, Value: "x"},
		},
	})
	if err != nil {
		t.Error(err)
	}

	expect = &generateGetInstancesFiltersResults{
		clauses: []string{`namespace_id = ?`},
		vals:    []interface{}{"x"},
	}

	err = expect.compare(clauses, vals)
	if err != nil {
		t.Errorf("generateGetInstancesFilters failed with one filter: %v", err)
	}

	t0 := time.Now().UTC()
	clauses, vals, err = generateGetInstancesFilters(&instancestore.ListOpts{
		Filters: []instancestore.Filter{
			{Field: fieldNamespaceID, Kind: instancestore.FilterKindMatch, Value: "x"},
			{Field: instancestore.FieldCreatedAt, Kind: instancestore.FilterKindBefore, Value: t0},
			{Field: instancestore.FieldCreatedAt, Kind: instancestore.FilterKindAfter, Value: t0},
			{Field: fieldDeadline, Kind: instancestore.FilterKindBefore, Value: t0},
			{Field: fieldDeadline, Kind: instancestore.FilterKindAfter, Value: t0},
			{Field: instancestore.FieldWorkflowPath, Kind: instancestore.FilterKindPrefix, Value: "x"},
			{Field: instancestore.FieldWorkflowPath, Kind: instancestore.FilterKindContains, Value: "x"},
			{Field: instancestore.FieldStatus, Kind: instancestore.FilterKindMatch, Value: instancestore.InstanceStatusComplete},
			{Field: instancestore.FieldStatus, Kind: "<", Value: instancestore.InstanceStatusComplete},
			{Field: instancestore.FieldInvoker, Kind: instancestore.FilterKindMatch, Value: "x"},
			{Field: instancestore.FieldInvoker, Kind: instancestore.FilterKindContains, Value: "x"},
		},
	})
	if err != nil {
		t.Error(err)
	}

	expect = &generateGetInstancesFiltersResults{
		clauses: []string{
			`namespace_id = ?`,
			`created_at < ?`,
			`created_at > ?`,
			`deadline < ?`,
			`deadline > ?`,
			`workflow_path LIKE ?`,
			`workflow_path LIKE ?`,
			`status = ?`,
			`status < ?`,
			`invoker = ?`,
			`invoker LIKE ?`,
		},
		vals: []interface{}{
			"x", t0, t0, t0, t0, "x%", "%x%", instancestore.InstanceStatusComplete,
			instancestore.InstanceStatusComplete, "x", "%x%",
		},
	}

	err = expect.compare(clauses, vals)
	if err != nil {
		t.Errorf("generateGetInstancesFilters failed with many filters: %v", err)
	}

	_, _, err = generateGetInstancesFilters(&instancestore.ListOpts{
		Filters: []instancestore.Filter{
			{Field: fieldNamespaceID, Kind: instancestore.FilterKindPrefix, Value: "x"},
		},
	})
	if !errors.Is(err, instancestore.ErrBadListOpts) {
		t.Errorf("generateGetInstancesFilters returned unexpected error: expected error is '%v', but got '%v'", instancestore.ErrBadListOpts, err)
	}

	// could do more tests here to check returned errors for every case
}

func Test_generateInsertQuery(t *testing.T) {
	res := generateInsertQuery(table, []string{fieldID})
	expect := `INSERT INTO instances_v2(id) VALUES (?)`

	if res != expect {
		t.Errorf("generateInsertQuery failed with one column: expected '%s', but got '%s'", expect, res)
	}

	res = generateInsertQuery(table, []string{fieldID, fieldNamespaceID, fieldWorkflowPath})
	expect = `INSERT INTO instances_v2(id, namespace_id, workflow_path) VALUES (?, ?, ?)`

	if res != expect {
		t.Errorf("generateInsertQuery failed with multiple columns: expected '%s', but got '%s'", expect, res)
	}
}

type generateGetInstancesQueriesResults struct {
	countQuery string
	query      string
	vals       []interface{}
}

func (expect *generateGetInstancesQueriesResults) compare(countQuery, query string, vals []interface{}) error {
	if countQuery != expect.countQuery {
		return fmt.Errorf("expected countQuery '%s', but got '%s'", expect.countQuery, countQuery)
	}

	if query != expect.query {
		return fmt.Errorf("expected query '%s', but got '%s'", expect.query, query)
	}

	if len(vals) != len(expect.vals) {
		return fmt.Errorf("expected vals '%v', but got '%v'", expect.vals, vals)
	}

	for idx := range vals {
		if vals[idx] != expect.vals[idx] {
			return fmt.Errorf("expected vals '%v', but got '%v'", expect.vals, vals)
		}
	}

	return nil
}

func Test_generateGetInstancesQueries(t *testing.T) {
	countQuery, query, vals, err := generateGetInstancesQueries([]string{fieldID}, nil)
	if err != nil {
		t.Error(err)
	}

	expect := &generateGetInstancesQueriesResults{
		countQuery: `SELECT COUNT(*) FROM instances_v2`,
		query:      `SELECT id FROM instances_v2 ORDER BY created_at desc`,
		vals:       nil,
	}

	err = expect.compare(countQuery, query, vals)
	if err != nil {
		t.Errorf("generateGetInstancesQueries failed with a single column: %v", err)
	}

	countQuery, query, vals, err = generateGetInstancesQueries([]string{fieldID, fieldNamespaceID}, &instancestore.ListOpts{
		Limit:  10,
		Offset: 10,
		Filters: []instancestore.Filter{
			{Field: fieldNamespaceID, Kind: instancestore.FilterKindMatch, Value: "x"},
			{Field: instancestore.FieldStatus, Kind: instancestore.FilterKindMatch, Value: instancestore.InstanceStatusFailed},
		},
	})
	if err != nil {
		t.Error(err)
	}

	expect = &generateGetInstancesQueriesResults{
		countQuery: `SELECT COUNT(*) FROM instances_v2 WHERE (namespace_id = ?) AND (status = ?)`,
		query:      `SELECT id, namespace_id FROM instances_v2 WHERE (namespace_id = ?) AND (status = ?) ORDER BY created_at desc LIMIT 10 OFFSET 10`,
		vals:       []interface{}{"x", instancestore.InstanceStatusFailed},
	}

	err = expect.compare(countQuery, query, vals)
	if err != nil {
		t.Errorf("generateGetInstancesQueries failed with a complex example: %v", err)
	}
}

package gateway2

import (
	"testing"
)

func Test_Valid_parseFetchSecretExpressionTwoArgs(t *testing.T) {
	tests := []struct {
		callExpression string
		namespace      string
		secretName     string
	}{
		{
			`fetchSecret("foo","bar")`, "foo", "bar",
		},
		{
			`fetchSecret(  "foo","bar")`, "foo", "bar",
		},
		{
			`fetchSecret   ("foo","bar")`, "foo", "bar",
		},
		{
			`fetchSecret("foo",   "bar")`, "foo", "bar",
		},
		{
			`fetchSecret("foo"  ,  "bar")`, "foo", "bar",
		},
		{
			`  fetchSecret  (  "foo"  , "bar" )  `, "foo", "bar",
		},
		{
			`fetchSecret(   "foo","bar"   )`, "foo", "bar",
		},
		{
			`fetchSecret("foo"    ,"bar")`, "foo", "bar",
		},
	}
	for _, tt := range tests {
		t.Run("case", func(t *testing.T) {
			got, err := parseFetchSecretExpressionTwoArgs(tt.callExpression)
			if err != nil {
				t.Errorf("parseFetchSecretExpression() error = %v", err)
				return
			}
			if tt.namespace != got.namespace {
				t.Errorf("parseFetchSecretExpression() namespace got = %v, want %v", got.namespace, tt.namespace)
			}
			if tt.secretName != got.secretName {
				t.Errorf("parseFetchSecretExpression() namespace got = %v, want %v", got.secretName, tt.secretName)
			}
		})
	}
}

func Test_InValid_parseFetchSecretExpressionTwoArgs(t *testing.T) {
	tests := []struct {
		callExpression string
	}{
		{
			`fetch Secret("foo","bar")`,
		},
		{
			`fetchSecret(foo","bar")`,
		},
		{
			`fetchSecret(foo,bar)`,
		},
		{
			`fetchSecret("foo" "bar")`,
		},
		{
			`fetchSecret("foo", "bar",)`,
		},
		{
			`fetchSecret(,"foo", "bar")`,
		},
	}
	for _, tt := range tests {
		t.Run("case", func(t *testing.T) {
			got, err := parseFetchSecretExpressionTwoArgs(tt.callExpression)
			if got != nil {
				t.Errorf("parseFetchSecretExpression() got = %v, want nil", got)
			}
			if err == nil {
				t.Errorf("parseFetchSecretExpression() error = %v", err)
				return
			}
		})
	}
}

func Test_Valid_parseFetchSecretExpressionSingleArgs(t *testing.T) {
	tests := []struct {
		callExpression string
		secretName     string
	}{
		{
			`fetchSecret("foo")`, "foo",
		},
		{
			`fetchSecret(  "foo")`, "foo",
		},
		{
			`fetchSecret   ("foo")`, "foo",
		},
		{
			`fetchSecret("foo")   `, "foo",
		},
		{
			`   fetchSecret(   "foo"   )`, "foo",
		},
		{
			`   fetchSecret  (   "foo"   ) `, "foo",
		},
	}
	for _, tt := range tests {
		t.Run("case", func(t *testing.T) {
			got, err := parseFetchSecretExpressionSingleArg(tt.callExpression)
			if err != nil {
				t.Errorf("parseFetchSecretExpression() error = %v", err)
				return
			}
			if got.namespace != "" {
				t.Errorf("parseFetchSecretExpression() namespace got = %v, want %v", got.secretName, "")
			}
			if tt.secretName != got.secretName {
				t.Errorf("parseFetchSecretExpression() namespace got = %v, want %v", got.secretName, tt.secretName)
			}
		})
	}
}

func Test_InValid_parseFetchSecretExpressionSingleArgs(t *testing.T) {
	tests := []struct {
		callExpression string
	}{
		{
			`fetch Secret("foo")`,
		},
		{
			`fetchSecret(foo)`,
		},
		{
			`fetchSecret('foo')`,
		},
		{
			`fetchSecret("foo",)`,
		},
		{
			`fetchSecret(,"foo")`,
		},
	}
	for _, tt := range tests {
		t.Run("case", func(t *testing.T) {
			got, err := parseFetchSecretExpressionSingleArg(tt.callExpression)
			if got != nil {
				t.Errorf("parseFetchSecretExpression() got = %v, want nil", got)
			}
			if err == nil {
				t.Errorf("parseFetchSecretExpression() error = %v", err)
				return
			}
		})
	}
}

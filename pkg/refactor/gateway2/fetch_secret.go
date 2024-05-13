package gateway2

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/database"
)

type fetchSecretArgs struct {
	namespace  string
	secretName string
}

func fetchSecret(db *database.SQLStore, namespace string, callExpression string) (string, error) {
	callExpression = strings.TrimSpace(callExpression)
	if !strings.HasPrefix(callExpression, "fetchSecret") {
		return callExpression, nil
	}

	fArgs, err := parseFetchSecretExpressionSingleArg(callExpression)
	if err != nil {
		fArgs, err = parseFetchSecretExpressionTwoArgs(callExpression)
	}
	if err != nil {
		return "", err
	}
	if fArgs.namespace == "" {
		fArgs.namespace = namespace
	}
	if fArgs.namespace != namespace && namespace != core.SystemNamespace {
		return "", fmt.Errorf("trying to fetch secret from different namespace")
	}

	s, err := db.DataStore().Secrets().Get(context.Background(), fArgs.namespace, fArgs.secretName)
	if err != nil {
		return "", fmt.Errorf("can not fetch secret: %w", err)
	}
	if !utf8.Valid(s.Data) {
		return "", fmt.Errorf("secret '%s' has none utf8 content", fArgs.secretName)
	}

	return string(s.Data), nil
}

func parseFetchSecretExpressionTwoArgs(callExpression string) (*fetchSecretArgs, error) {
	// parse fetchSecret( g1 ) pattern
	pattern := `^[ ]{0,}fetchSecret[ ]{0,}\((.*)\)[ ]{0,}$`
	regex := regexp.MustCompile(pattern)
	matches := regex.FindStringSubmatch(callExpression)
	if len(matches) != 2 {
		return nil, fmt.Errorf("syntax: invalid expression")
	}
	argsExpr := matches[1]

	// parse "g1", "g2" pattern
	pattern = `^[ ]{0,}[\"](.*)[\"][ ]{0,}[,][ ]{0,}[\"](.*)[\"][ ]{0,}$`
	regex = regexp.MustCompile(pattern)
	matches = regex.FindStringSubmatch(argsExpr)
	if len(matches) != 3 {
		return nil, fmt.Errorf("syntax: invalid arguments")
	}

	if strings.TrimSpace(matches[1]) != matches[1] {
		return nil, fmt.Errorf("syntax: extra spaces in arguments")
	}
	if strings.TrimSpace(matches[2]) != matches[2] {
		return nil, fmt.Errorf("syntax: extra spaces in arguments")
	}

	return &fetchSecretArgs{
		matches[1], matches[2],
	}, nil
}

func parseFetchSecretExpressionSingleArg(callExpression string) (*fetchSecretArgs, error) {
	// parse fetchSecret( g1 ) pattern
	pattern := `^[ ]{0,}fetchSecret[ ]{0,}\((.*)\)[ ]{0,}$`
	regex := regexp.MustCompile(pattern)
	matches := regex.FindStringSubmatch(callExpression)
	if len(matches) != 2 {
		return nil, fmt.Errorf("syntax: invalid expression")
	}
	argsExpr := matches[1]

	// parse "g1" pattern
	pattern = `^[ ]{0,}[\"](.*)[\"][ ]{0,}$`
	regex = regexp.MustCompile(pattern)
	matches = regex.FindStringSubmatch(argsExpr)
	if len(matches) != 2 {
		return nil, fmt.Errorf("syntax: invalid arguments")
	}

	if strings.TrimSpace(matches[1]) != matches[1] {
		return nil, fmt.Errorf("syntax: extra spaces in arguments")
	}

	return &fetchSecretArgs{
		"", matches[1],
	}, nil
}

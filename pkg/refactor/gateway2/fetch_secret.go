package gateway2

import (
	"fmt"
	"github.com/direktiv/direktiv/pkg/refactor/database"
	"regexp"
	"strings"
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

	return "", nil
}

func parseFetchSecretExpression(callExpression string) (*fetchSecretArgs, error) {
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
		strings.TrimSpace(matches[1]), matches[2],
	}, nil
}

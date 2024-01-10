package states

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"

	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/senseyeio/duration"
)

type actionRetryInfo struct {
	Children []*ChildInfo
	Idx      int
	Iterator int
}

type actionResultPayload struct {
	ActionID     string
	ErrorCode    string
	ErrorMessage string
	Output       []byte
}

func isRetryable(code string, patterns []string) bool {
	for _, pattern := range patterns {
		// NOTE: this error should be checked in model validation

		if pattern == "*" {
			pattern = ".*"
		}

		matched, _ := regexp.MatchString(pattern, code)
		if matched {
			return true
		}
	}

	return false
}

func retryDelay(attempt int, delay string, multiplier float64) time.Duration {
	d := time.Second * 5
	if x, err := duration.ParseISO8601(delay); err == nil {
		t0 := time.Now().UTC()
		t1 := x.Shift(t0)
		d = t1.Sub(t0)
	}

	if multiplier != 0 {
		for i := 0; i < attempt; i++ {
			d = time.Duration(float64(d) * multiplier)
		}
	}

	return d
}

func preprocessRetry(retry *model.RetryDefinition, attempt int, err error) (time.Duration, error) {
	var d time.Duration

	if retry == nil {
		return d, err
	}

	cerr := new(derrors.CatchableError)
	if !errors.As(err, &cerr) {
		return d, err
	}

	if !isRetryable(cerr.Code, retry.Codes) {
		return d, err
	}

	if attempt >= retry.MaxAttempts {
		return d, derrors.NewCatchableError("direktiv.retries.exceeded", "maximum retries exceeded")
	}

	d = retryDelay(attempt, retry.Delay, retry.Multiplier)

	return d, nil
}

func scheduleRetry(ctx context.Context, instance Instance, children []*ChildInfo, idx int, d time.Duration) error {
	var err error

	children[idx].Attempts++
	children[idx].ID = ""

	err = instance.SetMemory(ctx, children)
	if err != nil {
		return err
	}

	retry := &actionRetryInfo{
		Idx:      idx,
		Children: children,
		Iterator: idx,
	}

	err = instance.Sleep(ctx, d, retry)
	if err != nil {
		return err
	}

	return nil
}

type generateActionInputArgs struct {
	Instance Instance
	Source   interface{}
	Action   *model.ActionDefinition
	Files    []model.FunctionFileDefinition
	Iterator int
}

func generateActionInput(ctx context.Context, args *generateActionInputArgs) ([]byte, []model.FunctionFileDefinition, error) {
	var err error
	var input interface{}

	input, err = jqObject(args.Source, "jq(.)")
	if err != nil {
		return nil, nil, err
	}

	m, ok := input.(map[string]interface{})
	if !ok {
		err = derrors.NewInternalError(errors.New("invalid state data"))
		return nil, nil, err
	}

	m, err = addSecrets(ctx, args.Instance, m, args.Action.Secrets...)
	if err != nil {
		return nil, nil, err
	}

	if args.Action.Input == nil {
		input, err = jqOne(m, "jq(.)")
		if err != nil {
			return nil, nil, err
		}
	} else {
		input, err = jqOne(m, args.Action.Input)
		if err != nil {
			return nil, nil, err
		}
	}

	var inputData []byte

	inputData, err = json.Marshal(input)
	if err != nil {
		err = derrors.NewInternalError(err)
		return nil, nil, err
	}

	files := make([]model.FunctionFileDefinition, 0)

	for idx := range args.Files {
		file := args.Files[idx]

		s, err := jqString(m, file.As)
		if err != nil {
			return nil, nil, wrap(err, fmt.Sprintf("error evaluating jq in 'as' for function file %d: %%w", idx))
		}
		file.As = s

		s, err = jqString(m, file.Key)
		if err != nil {
			return nil, nil, wrap(err, fmt.Sprintf("error evaluating jq in 'key' for function file %d: %%w", idx))
		}
		file.Key = s

		if file.Key == "" {
			return nil, nil, derrors.NewCatchableError(ErrCodeInvalidVariableKey, "invalid 'key' for function file %d: got zero-length string", idx)
		}

		if file.Scope != util.VarScopeFileSystem && !util.VarNameRegex.MatchString(file.Key) {
			return nil, nil, derrors.NewCatchableError(ErrCodeInvalidVariableKey, "invalid 'key' for function file %d: must start with a letter and only contain letters, numbers and '_'", idx)
		}

		s, err = jqString(m, file.Scope)
		if err != nil {
			return nil, nil, wrap(err, fmt.Sprintf("error evaluating jq in 'scope' for function file %d: %%w", idx))
		}
		file.Scope = s

		switch file.Scope {
		case "":
		case util.VarScopeNamespace:
		case util.VarScopeWorkflow:
		case util.VarScopeInstance:
		case util.VarScopeThread:
		case util.VarScopeFileSystem:
		default:
			return nil, nil, derrors.NewCatchableError(ErrCodeInvalidVariableScope, "invalid 'scope' for function file %d: %s", idx, file.Scope)
		}

		s, err = jqString(m, file.Type)
		if err != nil {
			return nil, nil, wrap(err, fmt.Sprintf("error evaluating jq in 'type' for function file %d: %%w", idx))
		}
		file.Type = s

		if file.Permissions != "" {
			_, err := strconv.ParseUint(file.Permissions, 8, 32)
			if err != nil {
				return nil, nil, derrors.NewCatchableError(ErrCodeInvalidVariablePermissions, "invalid 'permissions' for function file %d: %s", idx, err.Error())
			}
		}

		files = append(files, file)
	}

	return inputData, files, nil
}

func addSecrets(ctx context.Context, instance Instance, m map[string]interface{}, secrets ...string) (map[string]interface{}, error) {
	if len(secrets) > 0 {
		s := make(map[string]string)

		for _, name := range secrets {
			dd, err := instance.RetrieveSecret(ctx, name)
			if err != nil {
				return nil, err
			}

			s[name] = dd
		}

		m["secrets"] = s
	}

	return m, nil
}

type invokeActionArgs struct {
	instance Instance
	async    bool
	fn       model.FunctionDefinition
	input    []byte
	attempt  int
	timeout  int
	files    []model.FunctionFileDefinition
	iterator int
}

func invokeAction(ctx context.Context, args invokeActionArgs) (*ChildInfo, error) {
	child, err := args.instance.CreateChild(ctx, CreateChildArgs{
		Definition: args.fn,
		Input:      args.input,
		Timeout:    args.timeout,
		Async:      args.async,
		Files:      args.files,
		Iterator:   args.iterator,
	})
	if err != nil {
		return nil, err
	}

	defer child.Run(ctx)

	ci := child.Info()

	if args.async {
		args.instance.Log(ctx, "info", "Running child '%s' in fire-and-forget mode (async).", ci.ID)
		//nolint:nilnil
		return nil, nil
	}

	return &ChildInfo{
		ID:          ci.ID,
		Type:        ci.Type,
		Attempts:    args.attempt,
		ServiceName: ci.ServiceName,
	}, nil
}

func ISO8601StringtoSecs(timeout string) (int, error) {
	// default 15 mins timeout
	wfto := 15 * 60

	if len(timeout) > 0 {
		to, err := duration.ParseISO8601(timeout)
		if err != nil {
			return wfto, err
		}

		dur := time.Until(to.Shift(time.Now().UTC()))
		wfto = int(dur.Seconds())
	}

	return wfto, nil
}

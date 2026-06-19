package invocation

import (
	"errors"
	"fmt"
	"strings"
)

type Result struct {
	BackendName   string
	NeedsSelector bool
	Model         string
	Prompt        string
}

func Parse(
	args []string,
	stdinText string,
	hasStdin bool,
	isSupportedBackend func(string) bool,
) (Result, error) {
	var backendName string
	var model string
	var promptArgs []string
	needsSelector := false

	for index := 0; index < len(args); index++ {
		arg := args[index]

		if value, ok := splitModelFlag(arg); ok {
			if value == nil {
				if index+1 >= len(args) {
					return Result{}, errors.New("missing model value")
				}
				model = args[index+1]
				if model == "" {
					return Result{}, errors.New("missing model value")
				}
				index++
				continue
			}
			if *value == "" {
				return Result{}, errors.New("missing model value")
			}
			model = *value
			continue
		}

		if value, ok := splitBackendFlag(arg); ok {
			if value != nil {
				parsedBackend, parsedPrompt, selector, err := parseBackendFlagValue(
					*value,
					len(args[index+1:]) > 0,
					len(promptArgs) > 0,
					hasStdin,
					isSupportedBackend,
				)
				if err != nil {
					return Result{}, err
				}
				if parsedBackend != "" {
					backendName = parsedBackend
				}
				if parsedPrompt != "" {
					promptArgs = append(promptArgs, parsedPrompt)
				}
				needsSelector = needsSelector || selector
				continue
			}

			if index+1 >= len(args) {
				needsSelector = true
				continue
			}

			next := args[index+1]
			parsedBackend, parsedPrompt, selector, err := parseBackendFlagValue(
				next,
				len(args[index+2:]) > 0,
				len(promptArgs) > 0,
				hasStdin,
				isSupportedBackend,
			)
			if err != nil {
				return Result{}, err
			}
			if parsedBackend != "" {
				backendName = parsedBackend
				index++
			}
			if parsedPrompt != "" {
				promptArgs = append(promptArgs, parsedPrompt)
				index++
			}
			needsSelector = needsSelector || selector
			continue
		}

		promptArgs = append(promptArgs, arg)
	}

	prompt := buildPrompt(strings.Join(promptArgs, " "), stdinText, hasStdin)
	if prompt == "" {
		return Result{}, errors.New(usageText())
	}

	return Result{
		BackendName:   backendName,
		NeedsSelector: needsSelector,
		Model:         model,
		Prompt:        prompt,
	}, nil
}

func splitModelFlag(arg string) (*string, bool) {
	const prefix = "--model="
	if strings.HasPrefix(arg, prefix) {
		value := strings.TrimPrefix(arg, prefix)
		return &value, true
	}

	switch arg {
	case "-m", "--model":
		return nil, true
	default:
		return nil, false
	}
}

func splitBackendFlag(arg string) (*string, bool) {
	for _, prefix := range []string{"--backend=", "--provider="} {
		if strings.HasPrefix(arg, prefix) {
			value := strings.TrimPrefix(arg, prefix)
			return &value, true
		}
	}

	switch arg {
	case "-b", "-p", "--backend", "--provider":
		return nil, true
	default:
		return nil, false
	}
}

func parseBackendFlagValue(
	value string,
	hasRemainingArgs bool,
	hasPriorPrompt bool,
	hasStdin bool,
	isSupportedBackend func(string) bool,
) (backendName string, prompt string, needsSelector bool, err error) {
	if value == "" {
		return "", "", true, nil
	}
	if isSupportedBackend != nil && isSupportedBackend(value) {
		return value, "", false, nil
	}
	if hasRemainingArgs || hasPriorPrompt || hasStdin {
		return "", "", false, fmt.Errorf("unsupported backend: %s", value)
	}
	return "", value, true, nil
}

func buildPrompt(argsPrompt string, stdinText string, hasStdin bool) string {
	if argsPrompt != "" && hasStdin && stdinText != "" {
		return argsPrompt + "\n\nContext:\n\n" + stdinText
	}
	if argsPrompt != "" {
		return argsPrompt
	}
	if hasStdin {
		return stdinText
	}
	return ""
}

func usageText() string {
	return strings.Join([]string{
		`Usage: qq [--backend BACKEND] "question"`,
		`       qq -b "question"`,
		`       echo "context" | qq "question"`,
	}, "\n")
}

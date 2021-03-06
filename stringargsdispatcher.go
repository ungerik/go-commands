package command

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode"
)

type constError string

func (e constError) Error() string {
	return string(e)
}

const (
	Default = ""

	ErrNotFound = constError("command not found")
)

type stringArgsCommand struct {
	command         string
	description     string
	args            Args
	commandFunc     interface{}
	stringArgsFunc  StringArgsFunc
	resultsHandlers []ResultsHandler
}

func checkCommandChars(command string) error {
	if strings.IndexFunc(command, unicode.IsSpace) >= 0 {
		return fmt.Errorf("command contains space characters: '%s'", command)
	}
	if strings.IndexFunc(command, unicode.IsGraphic) == -1 {
		return fmt.Errorf("command contains non graphc characters: '%s'", command)
	}
	if strings.ContainsAny(command, "|&;()<>") {
		return fmt.Errorf("command contains invalid characters: '%s'", command)
	}
	return nil
}

type StringArgsCommandLogger interface {
	LogStringArgsCommand(command string, args []string)
}

type StringArgsCommandLoggerFunc func(command string, args []string)

func (f StringArgsCommandLoggerFunc) LogStringArgsCommand(command string, args []string) {
	f(command, args)
}

type StringArgsDispatcher struct {
	comm    map[string]*stringArgsCommand
	loggers []StringArgsCommandLogger
}

func NewStringArgsDispatcher(loggers ...StringArgsCommandLogger) *StringArgsDispatcher {
	return &StringArgsDispatcher{
		comm:    make(map[string]*stringArgsCommand),
		loggers: loggers,
	}
}

func (disp *StringArgsDispatcher) AddCommand(command, description string, commandFunc interface{}, args Args, resultsHandlers ...ResultsHandler) error {
	if _, exists := disp.comm[command]; exists {
		return fmt.Errorf("Command '%s' already added", command)
	}
	if err := checkCommandChars(command); err != nil {
		return fmt.Errorf("Command '%s' returned: %w", command, err)
	}
	stringArgsFunc, err := GetStringArgsFunc(commandFunc, args, resultsHandlers...)
	if err != nil {
		return fmt.Errorf("Command '%s' returned: %w", command, err)
	}
	disp.comm[command] = &stringArgsCommand{
		command:         command,
		description:     description,
		args:            args,
		commandFunc:     commandFunc,
		stringArgsFunc:  stringArgsFunc,
		resultsHandlers: resultsHandlers,
	}
	return nil
}

func (disp *StringArgsDispatcher) MustAddCommand(command, description string, commandFunc interface{}, args Args, resultsHandlers ...ResultsHandler) {
	err := disp.AddCommand(command, description, commandFunc, args, resultsHandlers...)
	if err != nil {
		panic(err)
	}
}

func (disp *StringArgsDispatcher) AddDefaultCommand(description string, commandFunc interface{}, args Args, resultsHandlers ...ResultsHandler) error {
	stringArgsFunc, err := GetStringArgsFunc(commandFunc, args, resultsHandlers...)
	if err != nil {
		return fmt.Errorf("Default command: %w", err)
	}
	disp.comm[Default] = &stringArgsCommand{
		command:         Default,
		description:     description,
		args:            args,
		commandFunc:     commandFunc,
		stringArgsFunc:  stringArgsFunc,
		resultsHandlers: resultsHandlers,
	}
	return nil
}

func (disp *StringArgsDispatcher) MustAddDefaultCommand(description string, commandFunc interface{}, args Args, resultsHandlers ...ResultsHandler) {
	err := disp.AddDefaultCommand(description, commandFunc, args, resultsHandlers...)
	if err != nil {
		panic(err)
	}
}

func (disp *StringArgsDispatcher) HasCommnd(command string) bool {
	_, found := disp.comm[command]
	return found
}

func (disp *StringArgsDispatcher) HasDefaultCommnd() bool {
	_, found := disp.comm[Default]
	return found
}

func (disp *StringArgsDispatcher) Dispatch(ctx context.Context, command string, args ...string) error {
	cmd, found := disp.comm[command]
	if !found {
		return ErrNotFound
	}
	for _, logger := range disp.loggers {
		logger.LogStringArgsCommand(command, args)
	}
	return cmd.stringArgsFunc(ctx, args...)
}

func (disp *StringArgsDispatcher) MustDispatch(ctx context.Context, command string, args ...string) {
	err := disp.Dispatch(ctx, command, args...)
	if err != nil {
		panic(fmt.Errorf("Command '%s' returned: %w", command, err))
	}
}

func (disp *StringArgsDispatcher) DispatchDefaultCommand() error {
	return disp.Dispatch(context.Background(), Default)
}

func (disp *StringArgsDispatcher) MustDispatchDefaultCommand() {
	err := disp.DispatchDefaultCommand()
	if err != nil {
		panic(fmt.Errorf("Default command: %w", err))
	}
}

func (disp *StringArgsDispatcher) DispatchCombinedCommandAndArgs(ctx context.Context, commandAndArgs []string) (command string, err error) {
	if len(commandAndArgs) == 0 {
		return Default, disp.DispatchDefaultCommand()
	}
	command = commandAndArgs[0]
	args := commandAndArgs[1:]
	return command, disp.Dispatch(ctx, command, args...)
}

func (disp *StringArgsDispatcher) MustDispatchCombinedCommandAndArgs(ctx context.Context, commandAndArgs []string) (command string) {
	command, err := disp.DispatchCombinedCommandAndArgs(ctx, commandAndArgs)
	if err != nil {
		panic(fmt.Errorf("MustDispatchCombinedCommandAndArgs(%v): %w", commandAndArgs, err))
	}
	return command
}

func (disp *StringArgsDispatcher) PrintCommands(appName string) {
	list := make([]*stringArgsCommand, 0, len(disp.comm))
	for _, cmd := range disp.comm {
		list = append(list, cmd)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].command < list[j].command
	})

	for _, cmd := range list {
		CommandUsageColor.Printf("  %s %s %s\n", appName, cmd.command, cmd.args)
		if cmd.description != "" {
			CommandDescriptionColor.Printf("      %s\n", cmd.description)
		}
		hasAnyArgDesc := false
		for _, arg := range cmd.args.Args() {
			if arg.Description != "" {
				hasAnyArgDesc = true
			}
		}
		if hasAnyArgDesc {
			for _, arg := range cmd.args.Args() {
				CommandDescriptionColor.Printf("          <%s:%s> %s\n", arg.Name, arg.Type, arg.Description)
			}
		}
		CommandDescriptionColor.Println()
	}
}

func (disp *StringArgsDispatcher) PrintCommandsUsageIntro(appName string, output io.Writer) {
	if len(disp.comm) > 0 {
		fmt.Fprint(output, "Commands:\n")
		disp.PrintCommands(appName)
		fmt.Fprint(output, "Flags:\n")
	}
}

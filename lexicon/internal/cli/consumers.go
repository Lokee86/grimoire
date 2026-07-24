package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/Lokee86/lexicon/internal/config"
	"github.com/Lokee86/lexicon/internal/consumer"
	"github.com/Lokee86/lexicon/internal/objectstore"
)

type stringListFlag []string

func (values *stringListFlag) String() string {
	return strings.Join(*values, ",")
}

func (values *stringListFlag) Set(value string) error {
	*values = append(*values, value)
	return nil
}

func runConsumer(ctx context.Context, arguments []string, stdout, stderr io.Writer) error {
	if len(arguments) == 0 {
		return fmt.Errorf("consumer requires list, add, remove, or run")
	}
	switch arguments[0] {
	case "list":
		return runConsumerList(arguments[1:], stdout, stderr)
	case "add":
		return runConsumerAdd(arguments[1:], stdout, stderr)
	case "remove":
		return runConsumerRemove(arguments[1:], stdout, stderr)
	case "run":
		return runConsumerOne(ctx, arguments[1:], stdout, stderr)
	default:
		return fmt.Errorf("unknown consumer command %q", arguments[0])
	}
}

func runConsumerList(arguments []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("consumer list", flag.ContinueOnError)
	flags.SetOutput(stderr)
	repository := flags.String("repo", "", "repository to inspect")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	root, err := resolveRepository(*repository)
	if err != nil {
		return err
	}
	names, err := consumer.ListDefinitions(config.StateRoot(root))
	if err != nil {
		return err
	}
	for _, name := range names {
		fmt.Fprintln(stdout, strings.TrimSuffix(name, ".json"))
	}
	return nil
}

func runConsumerAdd(arguments []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("consumer add", flag.ContinueOnError)
	flags.SetOutput(stderr)
	repository := flags.String("repo", "", "repository to configure")
	name := flags.String("name", "", "consumer name")
	command := flags.String("command", "", "consumer executable")
	timeout := flags.Duration("timeout", 0, "consumer timeout")
	var args stringListFlag
	flags.Var(&args, "arg", "consumer argument; repeat as needed")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	if strings.TrimSpace(*name) == "" || strings.TrimSpace(*command) == "" {
		return fmt.Errorf("consumer add requires --name and --command")
	}
	root, err := resolveRepository(*repository)
	if err != nil {
		return err
	}
	fileName, err := consumerFileName(*name)
	if err != nil {
		return err
	}
	definition := consumer.Definition{Version: consumer.Version, Command: *command, Args: args, Timeout: *timeout}
	if err := consumer.AddDefinition(config.StateRoot(root), fileName, definition); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "registered consumer: %s\n", strings.TrimSuffix(fileName, ".json"))
	return nil
}

func runConsumerRemove(arguments []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("consumer remove", flag.ContinueOnError)
	flags.SetOutput(stderr)
	repository := flags.String("repo", "", "repository to configure")
	name := flags.String("name", "", "consumer name")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	if strings.TrimSpace(*name) == "" {
		return fmt.Errorf("consumer remove requires --name")
	}
	root, err := resolveRepository(*repository)
	if err != nil {
		return err
	}
	fileName, err := consumerFileName(*name)
	if err != nil {
		return err
	}
	if err := consumer.RemoveDefinition(config.StateRoot(root), fileName); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "removed consumer: %s\n", strings.TrimSuffix(fileName, ".json"))
	return nil
}

func runConsumerOne(ctx context.Context, arguments []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("consumer run", flag.ContinueOnError)
	flags.SetOutput(stderr)
	repository := flags.String("repo", "", "repository to use")
	name := flags.String("name", "", "consumer name")
	snapshot := flags.String("snapshot", "CURRENT", "snapshot ID or CURRENT")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	if strings.TrimSpace(*name) == "" {
		return fmt.Errorf("consumer run requires --name")
	}
	root, err := resolveRepository(*repository)
	if err != nil {
		return err
	}
	stateRoot := config.StateRoot(root)
	snapshotID := strings.TrimSpace(*snapshot)
	if snapshotID == "" || strings.EqualFold(snapshotID, "CURRENT") {
		snapshotID, _, err = (objectstore.Store{Root: stateRoot}).Current()
		if err != nil {
			return err
		}
	} else if _, err := (objectstore.Store{Root: stateRoot}).Load(snapshotID); err != nil {
		return err
	}
	fileName, err := consumerFileName(*name)
	if err != nil {
		return err
	}
	if err := consumer.RunOne(ctx, root, stateRoot, fileName, snapshotID, stdout); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "consumer %s processed %s\n", strings.TrimSuffix(fileName, ".json"), snapshotID)
	return nil
}

func consumerFileName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || filepath.Base(name) != name || strings.ContainsAny(name, `/\\`) {
		return "", fmt.Errorf("invalid consumer name %q", name)
	}
	if filepath.Ext(name) == "" {
		name += ".json"
	}
	return name, nil
}

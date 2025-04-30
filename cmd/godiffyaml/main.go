package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/rgeraskin/godiffyaml/internal/cmd"
)

type Runner interface {
	Init([]string) error
	Run() error
	Name() string
	Usage()
}

func usage(cmds *[]Runner) {
	fmt.Fprintf(
		os.Stderr,
		"%s\n\nShows human-readable diffs for yamls with multiple documents\n\n",
		os.Args[0],
	)

	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(
		os.Stderr,
		"  %s diff [options] <filename1> <filename2>\n\t%s",
		os.Args[0],
		"Split yaml contents of files to yaml docs and diff them by difftastic\n\n",
	)
	fmt.Fprintf(
		os.Stderr,
		"  %s k8s [options] <filename1> <filename2>\n\t%s",
		os.Args[0],
		fmt.Sprintf(
			"%s %s\n\n",
			"Same as 'diff' but with predefined",
			"'paths=apiVersion,kind,metadata.namespace,metadata.name'",
		),
	)
	fmt.Fprintf(
		os.Stderr,
		"  %s sort [options] <filename>\n\t%s %s",
		os.Args[0],
		"Read yaml file and dump it to stdout.",
		"If 'sort' arg is provided, docs inside will be sorted by values in paths\n\n",
	)

	for _, cmd := range *cmds {
		cmd.Usage()
		fmt.Fprintf(os.Stderr, "\n")
	}

	fmt.Fprintf(
		os.Stderr,
		"%s %s\n",
		"All unrecognized flags are passed directly to difftastic (except `--display`).",
		"Use --flag=value' notation, e.g. '--tab-width=10'",
	)
}

func wrapError(err error) string {
	errors := strings.Split(err.Error(), ":")
	for n, error := range errors[1:] {
		errors[n+1] = strings.Repeat(" ", n+1) + "because" + error
	}
	return strings.Join(errors, "\n")
}

func main() {
	// Disable logging
	log.SetOutput(io.Discard)

	// Supported subcommands
	subcommands := []Runner{
		cmd.NewDiffCommand(),
		cmd.NewK8SCommand(),
		cmd.NewSortCommand(),
	}

	// If no subcommand is provided, show usage
	if len(os.Args) < 2 {
		usage(&subcommands)
		os.Exit(1)
	}

	subcommandName := os.Args[1]

	for _, subcommand := range subcommands {
		if subcommand.Name() == subcommandName {
			if err := subcommand.Init(os.Args[2:]); err != nil {
				fmt.Fprintf(os.Stderr, "\n%s\n", err)
				os.Exit(1)
			}

			if err := subcommand.Run(); err != nil {
				if !strings.HasSuffix(err.Error(), "planned") {
					message := wrapError(err)
					fmt.Fprintf(os.Stderr, "%s\n", message)
				}
				os.Exit(1)
			}
			return
		}
	}

	// Show usage if other subcommand is provided
	usage(&subcommands)

	// Not Help subcommand
	if os.Args[1] != "help" && os.Args[1] != "-help" && os.Args[1] != "--help" &&
		os.Args[1] != "-h" {
		fmt.Fprintf(os.Stderr, "\nUnknown subcommand: %s\n", subcommandName)
		os.Exit(1)
	}
}

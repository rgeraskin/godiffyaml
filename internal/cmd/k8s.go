package cmd

import (
	"flag"
	"fmt"
	"strings"
)

func NewK8SCommand() *DiffCommand {
	cmd := &DiffCommand{
		flagSet: flag.NewFlagSet("k8s", flag.ExitOnError),
		paths:   "apiVersion,kind,metadata.namespace,metadata.name",
	}

	cmd.flagSet.StringVar(
		&cmd.display,
		"display",
		"side-by-side-show-both",
		fmt.Sprintf("Display format: %s", strings.Join(DISPLAYFORMATS, "|")),
	)

	return cmd
}

package cmd

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/rgeraskin/godiffyaml/internal/docs"

	"gopkg.in/yaml.v3"
)

func NewSortCommand() *SortCommand {
	cmd := &SortCommand{
		flagSet: flag.NewFlagSet("sort", flag.ExitOnError),
	}

	cmd.flagSet.StringVar(&cmd.order, "order", "", "Comma-separated list of fields to sort by")

	return cmd
}

type SortCommand struct {
	flagSet *flag.FlagSet

	filename string
	order    string
}

func (cmd *SortCommand) Name() string {
	return cmd.flagSet.Name()
}

func (cmd *SortCommand) Usage() {
	cmd.flagSet.Usage()
}

func (cmd *SortCommand) Init(args []string) error {
	if err := cmd.flagSet.Parse(args); err != nil {
		return fmt.Errorf("error parsing flags: %v", err)
	}

	if len(cmd.flagSet.Args()) != 1 {
		cmd.Usage()
		return fmt.Errorf("expected 1 argument, got %d", len(cmd.flagSet.Args()))
	}

	cmd.filename = cmd.flagSet.Args()[0]

	return nil
}

func (cmd *SortCommand) Run() error {
	return sortYAMLDocuments(cmd.filename, cmd.order)
}

func printYAMLDocuments(docs []docs.Doc) error {
	isFirst := true
	for _, doc := range docs {
		if !isFirst {
			fmt.Println("---")
		}
		isFirst = false

		encoder := yaml.NewEncoder(os.Stdout)
		encoder.SetIndent(2)
		if err := encoder.Encode(doc); err != nil {
			return fmt.Errorf("error encoding document: %v", err)
		}
		encoder.Close()
	}
	return nil
}

func readYAMLDocuments(filename string) ([]docs.Doc, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file %s: %v", filename, err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	var yamlDocs []docs.Doc

	// Iterate through all YAML documents in the file
	for {
		// Create an empty map to store the raw YAML data
		var yamlDoc docs.Doc
		// Attempt to decode the next YAML document from the file
		err := decoder.Decode(&yamlDoc)
		// Break the loop if we've reached the end of the file
		if err == io.EOF {
			break
		}
		// Return an error if decoding fails for any other reason
		if err != nil {
			return nil, fmt.Errorf("error decoding document %s: %v", filename, err)
		}

		// Append the document to our collection of documents
		yamlDocs = append(yamlDocs, yamlDoc)
	}

	return yamlDocs, nil
}

func sortYAMLDocuments(filename string, orderFlag string) error {
	documents, err := readYAMLDocuments(filename)
	if err != nil {
		return fmt.Errorf("error reading YAML document %s: %v", filename, err)
	}

	yamlDocs := docs.Docs{Docs: documents}
	if orderFlag != "" {
		order := strings.Split(orderFlag, ",")
		yamlDocs.Order = order
		sort.Sort(yamlDocs)
	}

	if err := printYAMLDocuments(yamlDocs.Docs); err != nil {
		return fmt.Errorf("error printing YAML document %s: %v", filename, err)
	}
	return nil
}

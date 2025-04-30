package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rgeraskin/godiffyaml/internal/docs"

	"gopkg.in/yaml.v3"
)

var DISPLAYFORMATS = []string{
	"side-by-side",
	"side-by-side-show-both",
	"inline",
	"json",
}

func NewDiffCommand() *DiffCommand {
	cmd := &DiffCommand{
		flagSet: flag.NewFlagSet("diff", flag.ExitOnError),
	}

	cmd.flagSet.StringVar(
		&cmd.paths,
		"paths",
		"",
		"Comma-separated list of paths to compose yaml doc filename",
	)
	cmd.flagSet.StringVar(
		&cmd.display,
		"display",
		"side-by-side-show-both",
		fmt.Sprintf("Display format: %s", strings.Join(DISPLAYFORMATS, "|")),
	)

	return cmd
}

type DiffCommand struct {
	flagSet *flag.FlagSet

	dir            string
	filename1      string
	filename2      string
	paths          string
	display        string
	argsDifftastic []string
}

func (cmd *DiffCommand) Name() string {
	return cmd.flagSet.Name()
}

func (cmd *DiffCommand) Usage() {
	cmd.flagSet.Usage()
}

// parseArgs parses the known flags and stores unknown args as difftastic args
func (cmd *DiffCommand) parseArgs(args []string) error {
	// Get names of known flags
	argsKnown := []string{}
	cmd.flagSet.VisitAll(func(f *flag.Flag) {
		argsKnown = append(argsKnown, f.Name)
	})

	argsParse := []string{} // This args will be parsed
	for _, arg := range args {
		// loop through known flags
		knownFound := false
		for _, known := range argsKnown {
			if strings.HasPrefix(arg, fmt.Sprintf("-%s", known)) ||
				strings.HasPrefix(arg, fmt.Sprintf("--%s", known)) {
				argsParse = append(argsParse, arg)
				knownFound = true
				break
			}
		}
		if knownFound {
			continue
		}

		// Unknown flags are for difftastic
		if strings.HasPrefix(arg, "-") {
			cmd.argsDifftastic = append(cmd.argsDifftastic, arg)
			continue
		}

		// Others should parse too
		argsParse = append(argsParse, arg)
	}

	// Parse only known flags
	return cmd.flagSet.Parse(argsParse)
}

func (cmd *DiffCommand) Init(args []string) error {
	if err := cmd.parseArgs(args); err != nil {
		return fmt.Errorf("error parsing flags: %v", err)
	}

	if len(cmd.flagSet.Args()) != 2 {
		cmd.Usage()
		return fmt.Errorf("expected 2 arguments, got %d", len(cmd.flagSet.Args()))
	}

	if cmd.paths == "" {
		cmd.Usage()
		return fmt.Errorf("paths flag is required")
	}

	if !slices.Contains(DISPLAYFORMATS, cmd.display) {
		cmd.Usage()
		return fmt.Errorf("display must be one of: %s", strings.Join(DISPLAYFORMATS, "|"))
	}

	cmd.filename1 = cmd.flagSet.Args()[0]
	cmd.filename2 = cmd.flagSet.Args()[1]

	return nil
}

func (cmd *DiffCommand) getDocFilePath(
	doc docs.Doc,
	writeDir string,
	inputFile string,
) (string, error) {
	var filenameParts []string
	for _, path := range strings.Split(cmd.paths, ",") {
		value := doc.GetValueByPath(path)
		if value == "" {
			fmt.Fprintf(
				os.Stderr,
				"# WARNING: no value found from path '%s' "+
					"in one of the documents in file '%s'\n",
				path,
				inputFile,
			)
		} else {
			value = strings.ReplaceAll(value, "/", "_")
			filenameParts = append(filenameParts, value)
		}
	}
	if len(filenameParts) == 0 {
		return "", fmt.Errorf(
			"no values found from yaml paths '%s' in file '%s'",
			cmd.paths,
			inputFile,
		)
	}
	fileName := strings.Join(filenameParts, "_") + ".yaml"
	return filepath.Join(writeDir, fileName), nil
}

func (cmd *DiffCommand) getFileToWrite(filePath string, inputFile string) (*os.File, error) {
	// check if file exists
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		return nil, fmt.Errorf(
			"file '%s' already exists: yaml paths '%s' are not unique for docs in file '%s'",
			filepath.Base(filePath),
			cmd.paths,
			inputFile,
		)
	}

	// create file
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file '%s': %v", filePath, err)
	}
	return file, nil
}

func (cmd *DiffCommand) writeYamlDocs(writeDir string, docs []docs.Doc, inputYaml string) error {
	// get paths values from each document to create a unique filename
	for _, doc := range docs {
		filePath, err := cmd.getDocFilePath(doc, writeDir, inputYaml)
		if err != nil {
			return fmt.Errorf("failed to get doc file path: %v", err)
		}

		log.Printf("prepare file %s to write yaml doc", filePath)

		file, err := cmd.getFileToWrite(filePath, inputYaml)
		if err != nil {
			return fmt.Errorf("failed to get file to write: %v", err)
		}
		defer file.Close()

		// write yaml doc to file
		encoder := yaml.NewEncoder(file)
		encoder.SetIndent(2)
		defer encoder.Close()
		if err := encoder.Encode(doc); err != nil {
			return fmt.Errorf("error encoding document: %v", err)
		}
	}

	return nil
}

// prepareDifftasticInput prepares the input dirs for difftastic
func (cmd *DiffCommand) prepareDifftasticInput() ([]string, error) {
	var dirs []string
	for n, filename := range []string{cmd.filename1, cmd.filename2} {
		log.Printf("processing yaml docs for %s", filename)
		subdir := fmt.Sprintf("%d", n)
		if err := cmd.processYamls(subdir, filename); err != nil {
			return nil, fmt.Errorf("failed to process yaml docs: %v", err)
		}
		dirs = append(dirs, subdir)
	}

	return dirs, nil
}

// check if difftastic is installed
func checkDifftastic() error {
	_, err := exec.Command("difft", "--version").Output()
	if err != nil {
		return fmt.Errorf("difft --version run failed: %v", err)
	}

	return nil
}

// runDifftastic prepares the args and runs difftastic
func (cmd *DiffCommand) runDifftastic(dirs []string) error {
	args := []string{"--display", cmd.display}
	args = append(args, dirs...)
	args = append(args, cmd.argsDifftastic...)

	// run difftastic
	log.Printf("running difftastic with args: %s", strings.Join(args, " "))
	c := exec.Command("difft", args...)
	c.Dir = cmd.dir
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return c.Run()
}

// difftastic compares the two dirs with yaml docs
func (cmd *DiffCommand) difftastic(dirs []string) error {
	// check if difftastic is installed
	if err := checkDifftastic(); err != nil {
		return err
	}

	// run difftastic
	if err := cmd.runDifftastic(dirs); err != nil {
		if slices.Contains(cmd.argsDifftastic, "--exit-code") {
			return fmt.Errorf("planned")
		}
		return fmt.Errorf("difftastic run failed: %v", err)
	}

	return nil
}

// processYamls reads the yaml file with docs and creates a subdir with a doc per file
func (cmd *DiffCommand) processYamls(subdir string, inputYamlFileName string) error {
	// check if file exists
	if _, err := os.Stat(inputYamlFileName); os.IsNotExist(err) {
		return fmt.Errorf("file '%s' does not exist", inputYamlFileName)
	}

	// read yaml documents
	docs, err := readYAMLDocuments(inputYamlFileName)
	if err != nil {
		return fmt.Errorf("failed to read file '%s': %v", inputYamlFileName, err)
	}

	// create sub directory
	writeDir := filepath.Join(cmd.dir, subdir)
	if err := os.MkdirAll(writeDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory to write yamls '%s': %v", writeDir, err)
	}

	// write yaml docs
	if err := cmd.writeYamlDocs(writeDir, docs, inputYamlFileName); err != nil {
		return fmt.Errorf("failed to write yamls: %v", err)
	}

	return nil
}

// Run runs the diff subcommand
func (cmd *DiffCommand) Run() error {
	log.Printf("running diff")
	log.Printf("filename1: %s", cmd.filename1)
	log.Printf("filename2: %s", cmd.filename2)
	log.Printf("paths: %s", cmd.paths)
	log.Printf("display: %s", cmd.display)

	// create temporary directory
	var err error
	cmd.dir, err = os.MkdirTemp("", "godiffyaml-")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(cmd.dir)
	log.Printf("temporary directory: %s", cmd.dir)

	// create dirs with yaml docs to compare with difftastic
	dirs, err := cmd.prepareDifftasticInput()
	if err != nil {
		return fmt.Errorf("failed to prepare difftastic input: %v", err)
	}

	// run difftastic and show output
	if err := cmd.difftastic(dirs); err != nil {
		return fmt.Errorf("failed to run difftastic: %v", err)
	}

	return nil
}

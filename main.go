package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/sliceutil"
	"github.com/bitrise-io/stepman/models"
	"gopkg.in/yaml.v2"
)

func main() {
	goFilePath, stepYMLPath := getFilepaths(os.Args)

	log.Infof("Analyzing %s", goFilePath)

	goinputs, err := analyzeGoFile(goFilePath)
	if err != nil {
		log.Errorf("%s", err)
		os.Exit(1)
	}

	fmt.Println("Found", len(*goinputs.envs), "inputs")

	log.Infof("Analyzing %s", stepYMLPath)

	step, err := analyzeStepYML(stepYMLPath)
	if err != nil {
		log.Errorf("%s", err)
		os.Exit(1)
	}

	fmt.Println("Found", len(step.Inputs), "inputs")
	fmt.Println()

	stepYMLInputs, err := extractStepInputKeys(step)
	if err != nil {
		log.Errorf("%s", err)
		os.Exit(1)
	}

	log.Infof("Result")

	var fail bool

	r := func(a1, a2 []string) {
		for _, i := range a1 {
			if sliceutil.IsStringInSlice(i, a2) {
				fmt.Print(" - ")
				log.Donef(i)
			} else {
				fmt.Print(" - ")
				log.Errorf(i)
				fail = true
			}
		}
	}

	fmt.Printf("- %s:\n", goFilePath)
	r(*goinputs.envs, stepYMLInputs)

	fmt.Printf("- %s:\n", stepYMLPath)
	r(stepYMLInputs, *goinputs.envs)

	if fail {
		os.Exit(1)
	}
}

func extractStepInputKeys(step models.StepModel) ([]string, error) {
	var stepYMLInputs []string
	for _, input := range step.Inputs {
		key, _, err := input.GetKeyValuePair()
		if err != nil {
			return nil, fmt.Errorf("could not get key value pairs: %s", err)
		}
		stepYMLInputs = append(stepYMLInputs, key)
	}
	return stepYMLInputs, nil
}

func analyzeStepYML(stepYMLPath string) (models.StepModel, error) {
	stepYMLFile, err := os.Open(stepYMLPath)
	if err != nil {
		return models.StepModel{}, fmt.Errorf("could not open %s: %v", stepYMLPath, err)
	}

	var step models.StepModel
	if err := yaml.NewDecoder(stepYMLFile).Decode(&step); err != nil {
		return models.StepModel{}, fmt.Errorf("could not parse %s: %v", stepYMLPath, err)
	}
	return step, nil
}

func analyzeGoFile(goFilePath string) (visitor, error) {
	f, err := parser.ParseFile(token.NewFileSet(), goFilePath, nil, 0)
	if err != nil {
		return visitor{}, fmt.Errorf("could not parse %s: %v", goFilePath, err)
	}

	v := visitor{&[]string{}}

	ast.Walk(v, f)
	return v, nil
}

func getFilepaths(osArgs []string) (goFilePath, stepYMLPath string) {
	var root string
	if len(osArgs) == 2 {
		root = osArgs[1]
	}

	stepYMLPath = filepath.Join(root, "step.yml")
	goFilePath = filepath.Join(root, "main.go")
	return
}

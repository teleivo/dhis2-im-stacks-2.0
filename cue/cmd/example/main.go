package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"cuelang.org/go/cue"
	"cuelang.org/go/tools/flow"
)

func main() {
	var r cue.Runtime
	f, err := os.Open("example.cue")
	if err != nil {
		fmt.Printf("failed to open file\n")
		os.Exit(1)
	}
	inst, err := r.Compile("example.cue", f)
	if err != nil {
		log.Fatal(err)
	}
	controller := flow.New(nil, inst, ioTaskFunc)
	if err := controller.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func ioTaskFunc(v cue.Value) (flow.Runner, error) {
	inputPath := cue.ParsePath("input")

	input := v.LookupPath(inputPath)
	if !input.Exists() {
		return nil, nil
	}

	return flow.RunnerFunc(func(t *flow.Task) error {
		inputVal, err := t.Value().LookupPath(inputPath).String()
		if err != nil {
			return fmt.Errorf("input not of type string")
		}

		outputVal := fmt.Sprintf("hello %s", inputVal)
		fmt.Printf("setting %s.output to %q\n", t.Path(), outputVal)

		return t.Fill(map[string]string{
			"output": outputVal,
		})
	}), nil
}

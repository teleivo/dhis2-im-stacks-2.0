package main

import (
	"fmt"
	"os"

	"github.com/teleivo/providers/stack"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("exit due to error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	err := deploy()
	if err != nil {
		return err
	}

	err = chain()
	if err != nil {
		return err
	}

	return nil
}

// deploy is a sketch of how it could look like when deploying dhis2-core linked to dhis2-db
// it shows consumed parameters and multiple variables/patterns previously only hostname pattern.
func deploy() error {
	// right now users provide the linked instance from which to consume
	// so imagine a user deploying an instance of dhis2-core linking to dhis2DBInstance
	source := stack.Instance{
		Name:  "mydb",
		Group: "whoami",
		Stack: stack.DHIS2DBStack,
		Parameters: map[string]stack.Parameter{
			"DATABASE_ID": {
				Value: "1",
			},
			"DATABASE_USERNAME": {
				Value: "foo",
			},
		},
	}

	// resolve parameters using target stack and source instance
	dhis2CoreInstanceParams := make(map[string]string, len(stack.DHIS2CoreStack.Parameters))
	for k, p := range stack.DHIS2CoreStack.Parameters {
		if !p.Consumed {
			dhis2CoreInstanceParams[k] = p.Value
			continue
		}

		// find parameter on the source instance parameters first
		if param, ok := source.Parameters[k]; ok {
			dhis2CoreInstanceParams[k] = param.Value
			continue
		}

		// find parameter on the source parameter providers next
		p, ok := source.Stack.Providers[k]
		if !ok {
			return fmt.Errorf("source stack %q cannot provide parameter %q", source.Stack.Name, k)
		}

		v, err := p.Provide(source)
		if err != nil {
			return fmt.Errorf("failed to evaluate parameter %q using source stack %q", k, source.Stack.Name)
		}
		dhis2CoreInstanceParams[k] = v
	}

	fmt.Println(dhis2CoreInstanceParams)

	return nil
}

// chain is a sketch of chained deployments.
func chain() error {
	// TODO create more stacks with some interesting requires
	// TODO create a d2 graph from the data :)
	// TODO can I create a CLI that allows me to show how we resolve a user selecting dhis2-core or
	// so?
	// TODO should
	return nil
}

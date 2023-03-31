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
		Stack: stack.DHIS2DB,
		Parameters: map[string]stack.Parameter{
			"DATABASE_ID": {
				Value: "1",
			},
			"DATABASE_USERNAME": {
				Value: "foo",
			},
			"DATABASE_PASSWORD": {
				Value: "faa",
			},
			"DATABASE_NAME": {
				Value: "mono",
			},
		},
	}

	// resolve parameters using source instance and target stack
	targetParams := make(map[string]stack.Parameter, len(stack.DHIS2Core.Parameters))
	for k, p := range stack.DHIS2Core.Parameters {
		if !p.Consumed {
			targetParams[k] = stack.Parameter{
				Value: p.Value,
			}
			continue
		}

		// find parameter on the source instance parameters first
		if param, ok := source.Parameters[k]; ok {
			targetParams[k] = stack.Parameter{
				Value:    param.Value,
				Consumed: true,
			}
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
		targetParams[k] = stack.Parameter{
			Value:    v,
			Consumed: true,
		}
	}

	fmt.Printf("deploying %q linked to %q(%s) with parameters %#v\n", "dhis-core", source.Name, source.Stack.Name, targetParams)

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

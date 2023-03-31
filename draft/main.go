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
	err := deployDHIS2Core()
	if err != nil {
		return fmt.Errorf("failed deploying dhis2-core: %v", err)
	}

	stacks, err := stack.New(
		stack.DHIS2DB,
		stack.DHIS2Core,
		stack.PgAdmin,
		stack.DHIS2,
		stack.WhoamiGo,
	)
	if err != nil {
		return fmt.Errorf("failed creating IM stacks: %v", err)
	}

	fmt.Println()
	err = drawStacks(stacks)
	if err != nil {
		return fmt.Errorf("failed drawing IM stack diagram: %v", err)
	}

	fmt.Println()
	err = chain()
	if err != nil {
		return err
	}

	return nil
}

// deployDHIS2Core is a sketch of how it could look like when deploying dhis2-core linked to dhis2-db
// it shows consumed parameters and multiple variables/patterns previously only hostname pattern.
func deployDHIS2Core() error {
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

func drawStacks(stacks stack.Stacks) error {
	f, err := os.Create("stacks.d2")
	if err != nil {
		return err
	}
	defer f.Close()

	required := make(map[string]struct{})
	for src, v := range stacks {
		for _, dest := range v.Requires {
			fmt.Fprintf(f, "%s -> %s\n", src, dest.Name)
			required[dest.Name] = struct{}{}
		}
	}
	for k := range stacks {
		if _, ok := required[k]; !ok {
			fmt.Fprintf(f, "%s\n", k)
		}
	}
	fmt.Printf("created https://d2lang.com diagram of IM stacks in %q\n", f.Name())

	return nil
}

// TODO analyze that a stack has its consumed parameters provided by its required Stacks
// making sure that there is only one provider per consumed parameter thus preventing conflicts
// TODO analyze that our stacks do not have a cycle
func analyze() error {
	return nil
}

// chain is a sketch of chained deployments.
func chain() error {
	// TODO create a simple CLI to show how we could guide a user through a chain
	return nil
}

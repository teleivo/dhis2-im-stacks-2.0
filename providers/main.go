package main

import (
	"fmt"
	"os"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("exit due to error: %v\n", err)
		os.Exit(1)
	}
}

type Stack struct {
	Name       string
	Parameters map[string]Parameter
	Providers  map[string]Provider
}

// Parameter is a stack parameter.
type Parameter struct {
	Value    string
	Consumed bool
}

// Provides a stack parameters value.
type Provider interface {
	Provide(instance Instance) (value string, err error)
}

type ProviderFunc func(instance Instance) (string, error)

func (p ProviderFunc) Provide(instance Instance) (string, error) {
	return p(instance)
}

type Instance struct {
	Name       string
	Group      string
	Stack      Stack
	Parameters map[string]Parameter
}

func run() error {
	// This is a sketch of how it could look like when deploying dhis2-core linked to dhis2-db
	// it shows consumed parameters and hostname variables

	dhis2DBStack := Stack{
		Name: "dhis2-db",
		Parameters: map[string]Parameter{
			"DATABASE_ID":       {},
			"DATABASE_USERNAME": {},
		},
		Providers: map[string]Provider{
			"DATABASE_HOSTNAME": ProviderFunc(func(instance Instance) (string, error) {
				return fmt.Sprintf("%s-database-postgresql.%s.svc", instance.Name, instance.Group), nil
			}),
		},
	}
	dhis2CoreStack := Stack{
		Name: "dhis2-core",
		Parameters: map[string]Parameter{
			"DHIS2_HOME": {
				Value: "/opt/dhis2",
			},
			"DATABASE_USERNAME": {
				Consumed: true,
			},
			"DATABASE_HOSTNAME": {
				Consumed: true,
			},
		},
	}

	// right now users provide the linked instance from which to consume
	// so imagine a user deploying an instance of dhis2-core linking to dhis2DBInstance
	source := Instance{
		Name:  "mydb",
		Group: "whoami",
		Stack: dhis2DBStack,
		Parameters: map[string]Parameter{
			"DATABASE_ID": {
				Value: "1",
			},
			"DATABASE_USERNAME": {
				Value: "foo",
			},
		},
	}

	// resolve parameters using target stack and source instance
	dhis2CoreInstanceParams := make(map[string]string, len(dhis2CoreStack.Parameters))
	for k, p := range dhis2CoreStack.Parameters {
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

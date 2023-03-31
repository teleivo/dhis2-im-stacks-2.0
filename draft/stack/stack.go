package stack

import "fmt"

type Stack struct {
	Name string
	// File is the path to the helmfile.
	File string
	// Parameters used by the stacks helmfile template.
	Parameters map[string]Parameter
	// Providers provide parameters to other stacks.
	Providers map[string]Provider
	// Requires these stacks to deploy an instance of this stack.
	Requires []Stack
}

// Parameter is a stack parameter.
type Parameter struct {
	Value string
	// Consumed signals that this parameter is provided by another i.e. one of the stacks required stacks.
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

// Instance of a stack which has all the parameters needed to deploy the instance.
type Instance struct {
	Name       string
	Group      string
	Stack      Stack
	Parameters map[string]Parameter
}

// Stack representing https://github.com/dhis2-sre/im-manager/blob/df95b498828ec7e2bb85245bf0e6a051f14f61fd/stacks/dhis2-db/helmfile.yaml
// Note: parameters are incomplete and might differ.
var DHIS2DB = Stack{
	Name: "dhis2-db",
	Parameters: map[string]Parameter{
		"DATABASE_ID":       {},
		"DATABASE_USERNAME": {},
		"DATABASE_PASSWORD": {},
		"DATABASE_NAME":     {},
	},
	Providers: map[string]Provider{
		"DATABASE_HOSTNAME": postgresHostNameProvider,
		"DATABASE_GREETING": ProviderFunc(func(instance Instance) (string, error) {
			return fmt.Sprintf("hello from stack %q instance %q", instance.Stack.Name, instance.Name), nil
		}),
	},
}

// Stack representing https://github.com/dhis2-sre/im-manager/blob/df95b498828ec7e2bb85245bf0e6a051f14f61fd/stacks/dhis2-core/helmfile.yaml
// Note: parameters are incomplete and might differ.
var DHIS2Core = Stack{
	Name: "dhis2-core",
	Parameters: map[string]Parameter{
		"DHIS2_HOME": {
			Value: "/opt/dhis2",
		},
		"DATABASE_USERNAME": {
			Consumed: true,
		},
		"DATABASE_PASSWORD": {
			Consumed: true,
		},
		"DATABASE_NAME": {
			Consumed: true,
		},
		"DATABASE_HOSTNAME": {
			Consumed: true,
		},
		"DATABASE_GREETING": { // just an example to show multiple "hostname variables" are possible
			Consumed: true,
		},
	},
	Requires: []Stack{
		DHIS2DB,
	},
}

// Stack representing https://github.com/dhis2-sre/im-manager/blob/df95b498828ec7e2bb85245bf0e6a051f14f61fd/stacks/dhis2/helmfile.yaml
// Note: parameters are incomplete and might differ.
var DHIS2 = Stack{
	Name: "dhis2-core",
	Parameters: map[string]Parameter{
		"DHIS2_HOME": {
			Value: "/opt/dhis2",
		},
		"DATABASE_USERNAME": {},
		"DATABASE_PASSWORD": {},
		"DATABASE_NAME":     {},
	},
	Providers: map[string]Provider{
		"DATABASE_HOSTNAME": postgresHostNameProvider,
	},
}

// Provides the PostgreSQL hostname as previously done by the hostname pattern.
var postgresHostNameProvider = ProviderFunc(func(instance Instance) (string, error) {
	return fmt.Sprintf("%s-database-postgresql.%s.svc", instance.Name, instance.Group), nil
})

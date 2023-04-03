// Package stack is a rough example of how we could structure stacks. Its important that we prevent
// any conflicts in instance parameters as such an instance cannot be deployed. It is important that
// we prevent any cycles in for example chained deployments as such chains cannot be deployed.
package stack

import (
	"errors"
	"fmt"

	"github.com/dominikbraun/graph"
)

type Stacks map[string]Stack

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

// Chain of stacks to be deployed in order.
type Chain struct {
	stacks  []Stack
	visited map[string]struct{}
	idx     map[string]int
	Chain   []Stack
}

// New creates stacks ensuring consumed parameters are provided by required stacks.
func New(stacks ...Stack) (Stacks, error) {
	err := validateConsumedParams(stacks)
	if err != nil {
		return nil, err
	}

	err = validateNoCycles(stacks)
	if err != nil {
		return nil, err
	}

	result := make(Stacks, len(stacks))
	for _, s := range stacks {
		result[s.Name] = s
	}
	return result, nil
}

func validateConsumedParams(stacks []Stack) error {
	var errs []error
	for _, s := range stacks { // validate each stacks consumed parameters are provided by its required stacks
		// collect all consumed parameters
		freq := make(map[string]int)
		for k, p := range s.Parameters {
			if !p.Consumed {
				continue
			}
			freq[k] = 0
		}

		// generate frequency map of provided parameters
		for _, dest := range s.Requires {
			// TODO does it matter if a consumed parameter is itself a consumed parameter on the required stack?
			// as long as we have no cycles its not a problem.
			for n := range dest.Parameters {
				_, ok := freq[n]
				if ok {
					freq[n]++
				}
			}
			for n := range dest.Providers {
				_, ok := freq[n]
				if ok {
					freq[n]++
				}
			}
		}
		for p, cnt := range freq {
			if cnt == 0 {
				errs = append(errs, fmt.Errorf("no provider for stack %q parameter %q", s.Name, p))
			}
			if cnt > 1 {
				errs = append(errs, fmt.Errorf("every consumed parameter must have exactly one provider. %d provider(s) for stack %q parameter %q", cnt, s.Name, p))
			}
		}
	}

	return errors.Join(errs...)
}

func validateNoCycles(stacks []Stack) error {
	g := graph.New(graph.StringHash, graph.Directed(), graph.PreventCycles())
	for _, s := range stacks {
		err := g.AddVertex(s.Name)
		if err != nil {
			return fmt.Errorf("failed adding vertex %q: %v", s.Name, err)
		}
	}
	for _, src := range stacks {
		for _, dest := range src.Requires {
			err := g.AddEdge(dest.Name, src.Name)
			if err != nil {
				if errors.Is(err, graph.ErrEdgeCreatesCycle) {
					return fmt.Errorf("edge %q -> %q creates cycle", dest.Name, src.Name)
				}
				return fmt.Errorf("failed adding edge %q -> %q: %v", dest.Name, src.Name, err)
			}
		}
	}

	return nil
}

// NewChain creates a stack chain of the given stacks. All stacks and their required stacks will be
// added to the chain in topological order. Any duplicate stacks will be ignored. Returns an error
// if given stacks contain a cycle.
// TODO some validation of stacks: cycles
func NewChain(stacks ...Stack) (*Chain, error) {
	c := Chain{
		visited: make(map[string]struct{}, len(stacks)),
		stacks:  stacks,
		Chain:   make([]Stack, 0, len(stacks)),
	}

	for _, s := range stacks {
		if _, ok := c.visited[s.Name]; !ok {
			c.dfs(s)
		}
	}

	return &c, nil
}

// Collect stacks in depth-first search order. We collect stacks that have no required stack i.e.
// vertices with no outgoing edges. This way required stacks will already be deployed before the
// stacks depending on them.
func (c *Chain) dfs(stack Stack) {
	c.visited[stack.Name] = struct{}{}
	for _, s := range stack.Requires {
		if _, ok := c.visited[s.Name]; !ok {
			c.dfs(s)
		}
	}
	c.Chain = append(c.Chain, stack)
}

// Add stack to the chain. Stack will be ignored if its already part of the chain. The chain is
// kept in topological order. Returns an error if adding the stack would cause a cycle.
func (c *Chain) Add(stack Stack) (*Chain, error) {
	return c, nil
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
	Name: "dhis2",
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

// Stack representing https://github.com/dhis2-sre/im-manager/blob/df95b498828ec7e2bb85245bf0e6a051f14f61fd/stacks/pgadmin/helmfile.yaml
// Note: parameters are incomplete and might differ.
var PgAdmin = Stack{
	Name: "pgadmin",
	Parameters: map[string]Parameter{
		"PGADMIN_USERNAME": {},
		"PGADMIN_PASSWORD": {},
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
	},
	Requires: []Stack{
		DHIS2DB,
	},
}

// Stack representing https://github.com/dhis2-sre/im-manager/blob/df95b498828ec7e2bb85245bf0e6a051f14f61fd/stacks/whoami-go/helmfile.yaml
// Note: parameters are incomplete and might differ.
var WhoamiGo = Stack{
	Name: "whoami-go",
	Parameters: map[string]Parameter{
		"REPLICA_COUNT": {
			Value: "1",
		},
	},
}

// im-job-runner: does not have any interesting consumed parameters right now
// https://github.com/dhis2-sre/im-manager/blob/df95b498828ec7e2bb85245bf0e6a051f14f61fd/stacks/im-job-runner/helmfile.yaml

// Provides the PostgreSQL hostname as previously done by the hostname pattern.
// Leveraging code as data and the Provider interface we can create reusable providers using any
// data an instance or its stack has. A Provider could in theory also reach out over the network to
// fetch some information. In this case I would suggest we add https://pkg.go.dev/context to the
// signature to enable timing out.
var postgresHostNameProvider = ProviderFunc(func(instance Instance) (string, error) {
	return fmt.Sprintf("%s-database-postgresql.%s.svc", instance.Name, instance.Group), nil
})

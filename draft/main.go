package main

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/dominikbraun/graph"
	"github.com/teleivo/providers/stack"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("exit due to error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	stacks, err := stack.New(
		stack.DHIS2Core,
		stack.DHIS2DB,
		stack.PgAdmin,
		stack.DHIS2,
		stack.WhoamiGo,
	)
	if err != nil {
		return fmt.Errorf("failed creating IM stacks: %v", err)
	}

	err = drawStacks(stacks)
	if err != nil {
		return fmt.Errorf("failed drawing IM stack diagram: %v", err)
	}

	fmt.Println()
	chain, err := pickChain(stacks)
	if err != nil {
		return err
	}

	fmt.Println()
	err = deploy(chain)
	if err != nil {
		return fmt.Errorf("failed deploying chain %v: %v", chain, err)
	}

	fmt.Println()
	err = deployDHIS2Core()
	if err != nil {
		return fmt.Errorf("failed deploying dhis2-core: %v", err)
	}

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

// pickChain is a sketch of chained deployments guiding users in selecting stacks.
// On every selection we automatically pick the required stacks and topologically sort them.
func pickChain(stacks stack.Stacks) ([]stack.Stack, error) {
	g := graph.New(graph.StringHash, graph.Directed(), graph.PreventCycles())

	opts := make([]stack.Stack, 0, len(stacks))
	for _, s := range stacks {
		opts = append(opts, s)
	}
	sort.Slice(opts, func(i, j int) bool {
		return opts[i].Name < opts[j].Name
	})
	var chainOrder []string

	fmt.Println("Pick a stack chain to deploy")
	for len(opts) > 0 {
		fmt.Print("Pick one of ")
		fmt.Print(renderOptions(opts))
		fmt.Print(": ")

		n, err := readNumber()
		if err != nil || n >= len(opts) {
			fmt.Println("Please choose one of the stacks")
			continue
		}

		opt := opts[n]
		opts = removeByIdx(opts, n)
		err = g.AddVertex(opt.Name)
		if err != nil && !errors.Is(err, graph.ErrVertexAlreadyExists) {
			return nil, fmt.Errorf("failed adding vertex %q: %v", opt.Name, err)
		}
		for _, dest := range opt.Requires {
			opts = removeByName(opts, dest.Name)
			err := g.AddVertex(dest.Name)
			if err != nil && !errors.Is(err, graph.ErrVertexAlreadyExists) {
				return nil, fmt.Errorf("failed adding vertex %q: %v", dest.Name, err)
			}
			err = g.AddEdge(dest.Name, opt.Name)
			if err != nil && !errors.Is(err, graph.ErrEdgeAlreadyExists) {
				return nil, fmt.Errorf("failed adding edge %q -> %q to chain: %v", dest.Name, opt.Name, err)
			}
		}

		chainOrder, err = graph.TopologicalSort(g)
		if err != nil {
			return nil, fmt.Errorf("failed topological sort stack chain: %v", err)
		}
		fmt.Printf("Current stack chain in deployment order: %v\n", chainOrder)

		if len(opts) > 0 {
			fmt.Print("Enter 0 to deploy stack|any other number to continue picking stacks: ")
			n, err = readNumber()
			if err != nil {
				fmt.Println("Please choose one of the stacks")
				continue
			}
			if n == 0 {
				break
			}
		}
	}

	chain := make([]stack.Stack, 0, len(chainOrder))
	for _, s := range chainOrder {
		chain = append(chain, stacks[s])
	}
	return chain, nil
}

func renderOptions(stacks []stack.Stack) string {
	var opts strings.Builder
	for i, s := range stacks {
		opts.WriteString(fmt.Sprintf("%d) %q", i, s.Name))
		if i < len(stacks)-1 {
			opts.WriteString(" ")
		}
	}
	return opts.String()
}

func readNumber() (int, error) {
	var choice int
	_, err := fmt.Scanf("%d", &choice)
	if err != nil {
		return 0, fmt.Errorf("failed to parse number: %v", err)
	}
	return choice, nil
}

func removeByIdx(stacks []stack.Stack, i int) []stack.Stack {
	return append(stacks[:i], stacks[i+1:]...)
}

func removeByName(stacks []stack.Stack, name string) []stack.Stack {
	for i, s := range stacks {
		if s.Name == name {
			return removeByIdx(stacks, i)
		}
	}

	return stacks
}

func deploy(chain []stack.Stack) error {
	stacks := make([]string, 0, len(chain))
	for _, s := range chain {
		stacks = append(stacks, s.Name)
	}
	fmt.Printf("deploying stack chain %v\n", stacks)

	// here we would iterate over the chain, deploy each instance which would resolve its parameters
	// so the subsequent stack instance can consume it
	// we can stop if a deployment fails and run any destroy hooks which are just functions defined
	// on the stack. They could have a similar signature as the Provider
	// Destroy(instance Instance) err error
	// maybe with a context
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

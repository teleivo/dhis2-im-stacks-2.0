package stack_test

import (
	"strings"
	"testing"

	"github.com/teleivo/providers/stack"
)

func TestNew(t *testing.T) {
	provider := stack.ProviderFunc(func(instance stack.Instance) (string, error) {
		return "1", nil
	})

	t.Run("Success", func(t *testing.T) {
		a := stack.Stack{
			Name: "a",
			Parameters: map[string]stack.Parameter{
				"a_param": {},
			},
			Providers: map[string]stack.Provider{
				"a_param_provided": provider,
			},
		}
		b := stack.Stack{
			Name: "b",
			Parameters: map[string]stack.Parameter{
				"a_param": {
					Consumed: true,
				},
				"a_param_provided": {
					Consumed: true,
				},
			},
			Requires: []stack.Stack{a},
		}

		stacks, err := stack.New(a, b)
		if err != nil {
			t.Fatalf("unexpected error %v", err)
		}

		if _, ok := stacks[a.Name]; !ok {
			t.Fatalf("want %q to be part of stacks, instead got %v", a.Name, stacks)
		}
		if _, ok := stacks[b.Name]; !ok {
			t.Fatalf("want %q to be part of stacks, instead got %v", b.Name, stacks)
		}
	})

	t.Run("FailGivenStackWithUnmetConsumedParameter", func(t *testing.T) {
		a := stack.Stack{
			Name: "a",
			Providers: map[string]stack.Provider{
				"a_param_provided": provider,
			},
		}
		b := stack.Stack{
			Name: "b",
			Parameters: map[string]stack.Parameter{
				"a_param": {
					Consumed: true,
				},
				"a_param_provided": {
					Consumed: true,
				},
			},
			Requires: []stack.Stack{a},
		}

		_, err := stack.New(a, b)
		if err == nil {
			t.Fatalf("expected error got none")
		}
		if want := `stack "b" parameter "a_param"`; !strings.Contains(err.Error(), want) {
			t.Fatalf("want error to contain '%s', instead got '%s'", want, err.Error())
		}
	})

	t.Run("FailGivenStackWithUnmetConsumedParameterDueToMissingProvider", func(t *testing.T) {
		a := stack.Stack{
			Name: "a",
			Parameters: map[string]stack.Parameter{
				"a_param": {},
			},
		}
		b := stack.Stack{
			Name: "b",
			Parameters: map[string]stack.Parameter{
				"a_param": {
					Consumed: true,
				},
				"a_param_provided": {
					Consumed: true,
				},
			},
			Requires: []stack.Stack{a},
		}

		_, err := stack.New(a, b)
		if err == nil {
			t.Fatalf("expected error got none")
		}
		if want := `stack "b" parameter "a_param_provided"`; !strings.Contains(err.Error(), want) {
			t.Fatalf("want error to contain '%s', instead got '%s'", want, err.Error())
		}
	})

	t.Run("FailGivenStackWithMultipleProvidersForOneParameter", func(t *testing.T) {
		a := stack.Stack{
			Name: "a",
			Parameters: map[string]stack.Parameter{
				"a_param": {},
			},
		}
		b := stack.Stack{
			Name: "b",
			Parameters: map[string]stack.Parameter{
				"a_param": {},
			},
		}
		c := stack.Stack{
			Name: "c",
			Parameters: map[string]stack.Parameter{
				"a_param": {
					Consumed: true,
				},
			},
			Requires: []stack.Stack{a, b},
		}

		_, err := stack.New(a, b, c)
		if err == nil {
			t.Fatalf("expected error got none")
		}
		if want := `stack "c" parameter "a_param"`; !strings.Contains(err.Error(), want) {
			t.Fatalf("want error to contain '%s', instead got '%s'", want, err.Error())
		}
	})

	t.Run("FailGivenStackWithMissingRequiredStack", func(t *testing.T) {
		t.Skip("TODO this is possible. Not sure if we could prevent this using a different API.")
		a := stack.Stack{
			Name: "a",
			Parameters: map[string]stack.Parameter{
				"a_param": {},
			},
		}
		b := stack.Stack{
			Name: "b",
			Parameters: map[string]stack.Parameter{
				"a_param": {
					Consumed: true,
				},
			},
			Requires: []stack.Stack{a},
		}

		_, err := stack.New(b)
		if err == nil {
			t.Fatalf("expected error got none")
		}
		if want := `TODO`; !strings.Contains(err.Error(), want) {
			t.Fatalf("want error to contain '%s', instead got '%s'", want, err.Error())
		}
	})
}

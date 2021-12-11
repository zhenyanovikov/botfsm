package main

import (
	"fmt"
	fsm "github.com/zhenyanovikov/botfsm"
	"os"
)

func main() {
	f := fsm.New()

	// When we start, we don't have state,
	// so any call should point to start, to
	// show start message
	f.RegisterMenu("",
		fsm.NewTransition(
			// this helper will return empty handler with arguments
			// passed to arguments of this functions
			fsm.StaticHandlerWithArgs(),
			fsm.StaticText(""),
		),
		fsm.Transitions{fsm.WildcardToken: "first"},
	)

	f.RegisterMenu("first",
		fsm.NewTransition(
			PrintHandlerFunc(),
			fsm.StaticText("first"),
		),
		fsm.Transitions{"second": "second"},
	)

	f.RegisterMenu("second",
		fsm.NewTransition(
			PrintHandlerFunc(),
			fsm.StaticText("second (GoForward). Jumping to third with handling"),
			// this flag will make Handle() function execute next
			// transition right after current
			fsm.GoForward(true),
		),
		fsm.Transitions{fsm.WildcardToken: "third"},
	)

	f.RegisterMenu("third",
		fsm.NewTransition(
			PrintHandlerFunc(),
			fsm.StaticText("third (GoForwardSilent). Jumping to third WITHOUT handling"),
			fsm.GoForwardSilent(true),
		),
		fsm.Transitions{fsm.WildcardToken: "fourth"},
	)

	f.RegisterMenu("fourth",
		fsm.NewTransition(
			PrintHandlerFunc(),
			fsm.StaticText("fourth"),
		),
		fsm.Transitions{fsm.WildcardToken: "end"},
	)

	f.RegisterMenu("end",
		fsm.NewTransition(
			ExitHandlerFunc(),
			fsm.StaticText(""),
		),
		fsm.Transitions{},
	)

	events := []string{"first", "second", "end"}

	var state string
	var args fsm.Arguments

	for _, event := range events {
		printInput(event)

		newState, newArgs, err := f.Handle(event, state, args, nil)
		if err != nil {
			panic(err)
		}

		printState(state, newState)

		state = newState
		args = newArgs
	}
}

func printState(state string, newState string) {
	fmt.Printf("[STATE] '%s' -> '%s'\n", state, newState)
}

func ExitHandlerFunc() fsm.HandlerFunc {
	return func(ctx *fsm.Context) (fsm.Arguments, error) {
		fmt.Println("Ending after fourth state. Bye!")
		os.Exit(1)
		return nil, nil
	}
}

func PrintHandlerFunc() fsm.HandlerFunc {
	return func(ctx *fsm.Context) (fsm.Arguments, error) {
		text, _, _ := ctx.Text()
		printOutput(text)
		return ctx.Arguments, nil
	}
}

func printOutput(str string) {
	fmt.Println("[OUT]", str)
}

func printInput(str string) {
	fmt.Println("[IN]", str)
}

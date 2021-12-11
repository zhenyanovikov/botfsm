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
	f.RegisterMenu("", fsm.NewTransition(
		// this helper will return empty handler with arguments
		// passed to arguments of this functions
		fsm.StaticHandlerWithArgs(),
		fsm.StaticText(""),
	), fsm.Transitions{fsm.WildcardToken: "start"})

	// Declare/Register start state
	f.RegisterMenu("start",
		fsm.NewTransition(
			// Setting up HandlerFunc, that will execute
			// after fsm.Handle(...) call
			func(ctx *fsm.Context) (fsm.Arguments, error) {
				if len(ctx.Arguments) == 1 {
					if ctx.Arg(0) == "exit" {
						fmt.Println("Goodbye!")
						os.Exit(1)
					}
				}
				text, _, _ := ctx.Text()
				printOutput(text)
				return fsm.Arguments{}, nil
			},

			// TextFunc is read-only func that generates text
			// for state. The reason for it ot exists is decoupling
			// work (HandlerFunc) function and printOutput data (text). This
			// allows you to get text for any state without thinking about
			// side effects of executing.
			// Also, see StaticText(str) which is alias for TextFunc
			// returning a str.
			func(data fsm.StateData) (string, interface{}, error) {
				if len(data.Arguments) == 1 {
					return fmt.Sprintf("Hey, %s, i know you!", data.Arg(0)), nil, nil
				}

				return "Enter your name: ", nil, nil
			},
		),
		fsm.Transitions{
			"exit":            "exit",
			fsm.WildcardToken: "enter_name",
		})

	f.RegisterMenu("exit",
		fsm.NewTransition(
			fsm.StaticHandlerWithArgs("exit"),
			fsm.StaticText(""),
			// this flag will make Handle() function execute next
			// transition right after current
			fsm.GoForward(true),
		),
		fsm.Transitions{
			fsm.WildcardToken: "start",
		},
	)

	f.RegisterMenu("enter_name",
		fsm.NewTransition(
			func(ctx *fsm.Context) (fsm.Arguments, error) {
				return fsm.Arguments{ctx.Event}, nil
			},
			fsm.StaticText(""),
			fsm.GoForward(true),
		),
		fsm.Transitions{
			fsm.WildcardToken: "start",
		},
	)

	events := []string{"", "Yevhenii", "exit"}

	var state string
	var args fsm.Arguments

	for _, event := range events {
		printInput(event)

		s, a, err := f.Handle(event, state, args, nil)
		if err != nil {
			panic(err)
		}

		fmt.Printf("'%s' -> '%s'\n", state, s)

		state = s
		args = a

	}
}

func printOutput(str string) {
	fmt.Println("[OUT]", str)
}

func printInput(str string) {
	fmt.Println("[IN]", str)
}

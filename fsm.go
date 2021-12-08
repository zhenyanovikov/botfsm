package botfsm

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

const (
	BackToken     = "\aback"
	WildcardToken = "*"
)

type (
	Arguments []string
	Keyboard  [][]string

	HandlerFunc     func(ctx *Context) (Arguments, error)
	BackMenu        *Menu
	BackHandler     HandlerFunc
	GoForward       bool
	GoForwardSilent bool

	//TextFunc returns text, attachment and error
	TextFunc     func(data StateData) (string, interface{}, error)
	KeyboardFunc func(data StateData) (Keyboard, error)

	Transitions map[string]string
)

//Transition is a struct with all transition data
type Transition struct {
	//Handler will be executed on Handle() method, on error state and arguments won't be returned
	Handler HandlerFunc
	//BackMenu pointer to menu, that will be associated with BackToken event
	BackMenu BackMenu
	// called before Handler, if previus state was BackToken
	// needed for context check (validate args len)
	BackHandler BackHandler

	//Text function to generate text (and attachments) for state, must be a read-only function
	Text TextFunc
	//Text function to generate keyboard for state, must be a read-only function
	Keyboard KeyboardFunc

	//GoForward will handle state of the only transition, right after Handle method
	GoForward bool
	//GoForwardSilent will return (without calling Handle) state of the only transition after Handle method
	GoForwardSilent bool
}

type BotFSM struct {
	mu    *sync.Mutex
	Menus map[string]*Menu
}

func New() *BotFSM {
	return &BotFSM{
		Menus: make(map[string]*Menu),
		mu:    &sync.Mutex{},
	}
}

func NewTransition(handler HandlerFunc, text TextFunc, opt ...interface{}) Transition {
	obj, err := ShouldNewTransition(handler, text, opt...)
	if err != nil {
		panic(err)
	}
	return obj
}

func ShouldNewTransition(handler HandlerFunc, text TextFunc, opts ...interface{}) (Transition, error) {
	if handler == nil {
		return Transition{}, errors.New("handler is nil")
	}

	transition := Transition{
		Handler: handler,
		Text:    text,
	}

	for _, opt := range opts {
		switch field := opt.(type) {
		case KeyboardFunc:
			transition.Keyboard = field
		case BackMenu:
			transition.BackMenu = field
		case BackHandler:
			transition.BackHandler = field
		case GoForward:
			transition.GoForward = bool(field)
		case GoForwardSilent:
			transition.GoForwardSilent = bool(field)
		default:
			return Transition{}, errors.New("unknown data type passed to optionals: " + reflect.TypeOf(field).String())
		}
	}

	if transition.Handler == nil {
		return Transition{}, errors.New("handler is nil")
	}

	if !transition.GoForward && !transition.GoForwardSilent {
		if text == nil {
			return Transition{}, errors.New("text func is nil")
		}
	}

	return transition, nil
}

func (fsm *BotFSM) ShouldRegisterMenu(state string, transition Transition, transitions Transitions) (*Menu, error) {
	fsm.mu.Lock()
	defer fsm.mu.Unlock()

	if menu, ok := fsm.Menus[state]; ok {
		return menu, nil
	}

	if err := validateMenu(transition, transitions); err != nil {
		return nil, err
	}

	menu := Menu{fsm, state, transition, transitions}

	fsm.Menus[state] = &menu
	return &menu, nil
}

func (fsm *BotFSM) RegisterMenu(state string, transition Transition, transitions Transitions) *Menu {
	menu, err := fsm.ShouldRegisterMenu(state, transition, transitions)
	if err != nil {
		panic(err)
	}
	return menu
}

//Handle returns newState and newArguments or error
func (fsm *BotFSM) Handle(event string, state string, arguments []string, vars map[string]interface{}) (string, []string, error) {
	menu := fsm.Menus[state]
	nextState, ok := menu.Transitions[event]
	if !ok {
		transition, ok := menu.Transitions[WildcardToken]
		if !ok {
			return "", nil, &NoTransitionError{text: fmt.Sprintf("no transitions found for this state (%s) and event (%s)", state, event)}
		}
		nextState = transition
	}
	nextMenu, ok := fsm.Menus[nextState]
	if !ok {
		if nextState != BackToken {
			return "", nil, fmt.Errorf("not found next menu for state '%s'", nextState)
		}
		nextMenu = menu.Transition.BackMenu
		nextState = menu.Transition.BackMenu.State
	}

	fsmContext := NewContext(vars, event, arguments, nextMenu.Transition.Text, nextMenu.Transition.Keyboard)

	if nextMenu.Transition.BackHandler != nil {
		args, err := nextMenu.Transition.BackMenu.Transition.BackHandler(fsmContext)
		if err != nil {
			return "", nil, err
		}
		arguments = args
	}

	newArguments, err := nextMenu.Transition.Handler(fsmContext)
	if err != nil {
		return "", nil, err
	}

	if nextMenu.Transition.GoForward {
		return fsm.Handle(WildcardToken, nextState, newArguments, vars)
	}

	if nextMenu.Transition.GoForwardSilent {
		return nextMenu.Transitions[WildcardToken], newArguments, nil
	}
	return nextState, newArguments, nil
}

//Validate will check fsm for transitions to state, that does not exist and more
func (fsm *BotFSM) Validate() error {
	var errorsStack []string
	for _, menu := range fsm.Menus {
		if menu.Transition.GoForward && menu.Transitions[WildcardToken] == menu.State {
			errorsStack = append(errorsStack, fmt.Sprintf("\ton state '%s' GoForward transition is recursive",
				menu.State))
		}

		for _, transition := range menu.Transitions {
			if transition == BackToken {
				continue
			}
			if _, ok := fsm.Menus[transition]; !ok {
				errorsStack = append(errorsStack, fmt.Sprintf("\ton state '%s' transition '%s' have no menu",
					menu.State, transition))
			}
		}
	}

	if len(errorsStack) != 0 {
		return errors.New("Validate failed. Not found menus\n" + strings.Join(errorsStack, "\n"))
	}

	return nil
}

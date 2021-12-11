package botfsm

import (
	"errors"
	"fmt"
	"strings"
)

type Menu struct {
	fsm *BotFSM

	State      string
	Transition *Transition
	//Transitions is a map[event]state, that is mapping events with state
	Transitions Transitions
}

func (m *Menu) ShouldConfirmAfter(handler HandlerFunc, backMenu *Menu, successState string, text TextFunc,
	yesButton, noButton string) (*Menu, error) {
	confirmState := m.State + "_confirm"

	for event, transition := range backMenu.Transitions {
		if transition == successState {
			backMenu.Transitions[event] = confirmState
			break
		}
	}

	return m.fsm.ShouldRegisterMenu(m.State+"_confirm", NewTransition(
		handler, text, StaticKeyboard(yesButton, noButton), BackMenu(backMenu),
	), Transitions{
		yesButton: successState,
		noButton:  BackToken,
	})
}

func (m *Menu) ConfirmAfter(handler HandlerFunc, backMenu *Menu, successState string, text TextFunc,
	yesButton, noButton string) *Menu {
	menu, err := m.ShouldConfirmAfter(handler, backMenu, successState, text, yesButton, noButton)
	if err != nil {
		panic(err)
	}
	return menu
}

func validateMenu(transition *Transition, transitions Transitions) error {
	if transition == nil {
		return errors.New("transition is nil")
	}

	if transition.GoForward {
		if len(transitions) != 1 {
			return errors.New("transition must have only 1 transition if it is marked as GoForward")
		}

		if _, ok := transitions[WildcardToken]; !ok {
			return errors.New("transition from transitions must be WildcardToken if it is marked as GoForward")
		}
	}

	//backTokenUsed := false
	for event, state := range transitions {
		//if state == BackToken {
		//	backTokenUsed = true
		//}
		if state == BackToken && transition.BackMenu == nil {
			return errors.New("BackToken is used, but BackMenu is not assigned")
		}
		if strings.TrimSpace(state) == "" {
			return fmt.Errorf("transition from event '%s' is empty", event)
		}
		if strings.TrimSpace(event) == "" {
			return fmt.Errorf("event for transition '%s' is empty", transitions)
		}
	}

	//if transition.BackMenu != nil && !backTokenUsed {
	//	return errors.New("BackMenu is specified, but transitions have no BackToken")
	//}

	return nil
}

func validateTransition(transition Transition) error {
	if transition.Text == nil {
		return errors.New("text func is nil")
	}
	if transition.Keyboard == nil {
		return errors.New("keyboard func is nil")
	}
	return nil
}

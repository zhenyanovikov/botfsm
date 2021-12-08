package botfsm

func StaticText(text string) TextFunc {
	return func(data StateData) (string, interface{}, error) {
		return text, nil, nil
	}
}

func StaticKeyboard(keyboardButtons ...string) KeyboardFunc {
	keyboard := make(Keyboard, len(keyboardButtons))
	for i, button := range keyboardButtons {
		keyboard[i] = []string{button}
	}

	return func(data StateData) (Keyboard, error) {
		return keyboard, nil
	}
}

package botfsm

type NoTransitionError struct {
	text string
}

func (t *NoTransitionError) Error() string {
	return t.text
}

package botfsm

import (
	"strconv"
	"sync"
)

type Context struct {
	StateData
	textFunc     TextFunc
	keyboardFunc KeyboardFunc
}

type StateData struct {
	Event     string
	Arguments Arguments
	mu        *sync.RWMutex
	values    map[string]interface{}
}

func NewStateData(event string, arguments Arguments, vars map[string]interface{}) StateData {
	if vars == nil {
		vars = make(map[string]interface{})
	}
	return StateData{
		Event:     event,
		Arguments: arguments,
		mu:        &sync.RWMutex{},
		values:    vars,
	}
}

func NewContext(vars map[string]interface{}, event string, arguments Arguments, textFunc TextFunc, keyboardFunc KeyboardFunc) *Context {
	return &Context{
		StateData:    NewStateData(event, arguments, vars),
		textFunc:     textFunc,
		keyboardFunc: keyboardFunc,
	}
}

func (c *Context) Text() (string, interface{}, error) {
	return c.textFunc(c.StateData)
}

func (c *Context) Keyboard() (Keyboard, error) {
	return c.keyboardFunc(c.StateData)
}

func (c *StateData) Arg(index int) string {
	return c.Arguments[index]
}

func (c *StateData) ArgInt(index int) int {
	a, _ := strconv.Atoi(c.Arguments[index])
	return a
}

func (c *StateData) SetValue(key string, value interface{}) {
	c.mu.Lock()
	c.values[key] = value
	c.mu.Unlock()
}

func (c *StateData) Value(key string) interface{} {
	c.mu.RLock()
	v := c.values[key]
	c.mu.RUnlock()
	return v
}

func (c *StateData) GetValue(key string) (interface{}, bool) {
	c.mu.RLock()
	v, ok := c.values[key]
	c.mu.RUnlock()
	return v, ok
}

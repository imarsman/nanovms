package stack

import (
	"container/list"
	"fmt"
)

// NewStack get a new instance of a Stack struct
func NewStack() *Stack {
	stack := &Stack{
		stack: list.New(),
	}

	return stack
}

// Stack a simple stack
type Stack struct {
	stack *list.List
}

// Push to the stack
func (c *Stack) Push(value string) {
	c.stack.PushFront(value)
}

// Pop from stack
func (c *Stack) Pop() error {
	if c.stack.Len() > 0 {
		ele := c.stack.Front()
		c.stack.Remove(ele)
	}
	return fmt.Errorf("Pop Error: Queue is empty")
}

// Front get front item in stack if it is a string
func (c *Stack) Front() (string, error) {
	if c.stack.Len() > 0 {
		if val, ok := c.stack.Front().Value.(string); ok {
			return val, nil
		}
		return "", fmt.Errorf("Peep Error: Queue Datatype is incorrect")
	}
	return "", fmt.Errorf("Peep Error: Queue is empty")
}

// Size get size of stack
func (c *Stack) Size() int {
	return c.stack.Len()
}

// Empty is stack empty
func (c *Stack) Empty() bool {
	return c.stack.Len() == 0
}

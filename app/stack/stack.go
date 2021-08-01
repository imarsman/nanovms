package stack

import (
	"container/list"
	"fmt"
)

/*
	A simple stack using a Go container/list, which is an ordered list of
	interface values. This stack implementation assumes strings are stored.
*/

// From https://golangbyexample.com/stack-in-golang/

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
func (s *Stack) Push(value string) {
	s.stack.PushFront(value)
}

// Pop from stack
func (s *Stack) Pop() error {
	if s.stack.Len() > 0 {
		ele := s.stack.Front()
		s.stack.Remove(ele)
	}
	return fmt.Errorf("Pop Error: Queue is empty")
}

// Front get front item in stack if it is a string
func (s *Stack) Front() (interface{}, error) {
	if s.stack.Len() > 0 {
		val := s.stack.Front()
		return val.Value, nil
	}
	// return "", fmt.Errorf("Peep Error: Queue Datatype is incorrect")
	return nil, fmt.Errorf("Peep Error: Queue is empty")
}

// Size get size of stack
func (s *Stack) Size() int {
	return s.stack.Len()
}

// Empty is stack empty
func (s *Stack) Empty() bool {
	return s.stack.Len() == 0
}

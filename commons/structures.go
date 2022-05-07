package commons

import "log"

type Stack[T any] struct {
	array []T
}

func NewStack[T any](elements ...T) *Stack[T] {
	res := Stack[T]{
		array: make([]T, 0, len(elements)),
	}
	res.array = append(res.array, elements...)
	return &res
}

func (stack *Stack[T]) ReplaceTop(value T) {
	if len(stack.array) > 0 {
		stack.array = stack.array[len(stack.array)-2:]
	}
	stack.array = append(stack.array, value)
}

func (stack *Stack[T]) Pop() T {
	length := len(stack.array)
	if length < 1 {
		log.Fatal("Can't pop: no elements in stack")
	}

	result := stack.array[length-1]
	stack.array = stack.array[:length-1]
	return result
}

func (stack *Stack[T]) Top() T {
	length := len(stack.array)
	if length < 1 {
		log.Fatal("Can't pop: no elements in stack")
	}

	return stack.array[length-1]
}

func (stack *Stack[T]) Push(value T) {
	stack.array = append(stack.array, value)
}
func (stack *Stack[T]) Length() int {
	return len(stack.array)
}

type Pair[f any, s any] struct {
	First  f
	Second s
}

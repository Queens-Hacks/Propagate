package sandbox

import (
	"github.com/Shopify/go-lua"
)

type WorldState struct {
	i int
}

type NewState struct {
	i int
}

type Node struct {
	resume  chan<- WorldState
	respond <-chan NewState
}

func (n *Node) Update(state WorldState) <-chan NewState {
	n.resume <- state
	return n.respond
}

func AddNode(program string) *Node {
	// Make the communication channels
	resume := make(chan WorldState)
	respond := make(chan NewState)

	n := Node{
		resume:  resume,
		respond: respond,
	}

	go runNode(internalNode{
		program: program,
		resume:  resume,
		respond: respond,
	})

	return &n
}

type internalNode struct {
	program string
	resume  <-chan WorldState
	respond chan<- NewState
}

type Direction int

const (
	Left Direction = iota
	Right
	Up
	Down
)

func addDirFunc(l *lua.State, name string, fn func(Direction) int) {
	l.PushGoFunction(func(l *lua.State) int {
		argCount := l.Top()
		if argCount != 1 {
			l.PushString("incorrect number of arguments") // XXX Include name of function
			l.Error()
			return 0
		}
		s, err := l.ToString(1) // toLowerCAse
		if err != nil {
			l.PushString("incorrect type of argument") // XXX Include name of function
			l.Error()
			return 0
		}

		if s == "left" {

		} else if s == "right" {
		} else if s == "up" {
		} else if s == "down" {
		}

	})
}

func runNode(node internalNode) {
	l := lua.NewState()
	l.PushGoFunction(func(l *lua.State) int {

	})
	lua.LoadString(l, node.program)
}

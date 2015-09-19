package sandbox

import (
	"github.com/Shopify/go-lua"
)

type WorldState struct {
	lighting map[Direction]float64
}

type NewState struct {
	moveDir Direction
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

func addIntFunc(l *lua.State, name string, fn func(*lua.State, int) int) {
	l.PushGoFunction(func(l *lua.State) int {
		if l.Top() != 1 {
			l.PushString("Wrong number of arguments")
			l.Error()
			return 0
		}

		i, ok := l.ToInteger(1)
		if !ok {
			l.PushString("Wrong argument type")
			l.Error()
			return 0
		}

		return fn(l, i)
	})

	l.SetGlobal(name)
}

func addVoidFunc(l *lua.State, name string, fn func(*lua.State) int) {
	l.PushGoFunction(func(l *lua.State) int {
		if l.Top() != 0 {
			l.PushString("Too many arguments to void function")
			l.Error()
			return 0
		}

		return fn(l)
	})

	l.SetGlobal(name)
}

func addDirFunc(l *lua.State, name string, fn func(*lua.State, Direction) int) {
	l.PushGoFunction(func(l *lua.State) int {
		argCount := l.Top()
		if argCount != 1 {
			l.PushString("incorrect number of arguments") // XXX Include name of function
			l.Error()
			return 0
		}

		s, ok := l.ToString(1)
		if !ok {
			l.PushString("incorrect type of argument") // XXX Include name of function
			l.Error()
			return 0
		}

		var d Direction
		if s == "left" {
			d = Left
		} else if s == "right" {
			d = Right
		} else if s == "up" {
			d = Up
		} else if s == "down" {
			d = Down
		}

		return fn(l, d)
	})

	l.SetGlobal(name)
}

func runNode(node internalNode) {
	l := lua.NewState()

	world := <-node.resume

	addDirFunc(l, "grow", func(l *lua.State, d Direction) int {
		var state NewState
		state.moveDir = d

		// Send a response and wait
		node.respond <- state
		world = <-node.resume

		return 0
	})

	addDirFunc(l, "lighting", func(l *lua.State, d Direction) int {
		l.PushNumber(world.lighting[d])
		return 1
	})

	lua.LoadString(l, node.program)
}

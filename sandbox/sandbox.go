package sandbox

import (
	"github.com/Shopify/go-lua"
	"time"
)

type WorldState struct {
	Lighting map[Direction]float64
}

type StateChange int

const (
	Move StateChange = iota
	Split
	Wait
)

type NewState struct {
	Operation StateChange
	Dir Direction
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

	in := internalNode{
		program: program,
		resume:  resume,
		respond: respond,
	}

	go runNode(in)

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
	Undef
)

func watchLuaThread(l *lua.State, d time.Duration) {
	end_time := time.Now().Add(d)
	setLuaTimeoutHook(l, 500, func() {
		if time.Now().After(end_time) {
			panic("AAAAHHHH!!!")
		}
	})
	l.ProtectedCall(0, lua.MultipleReturns, 0)
}

func setLuaTimeoutHook(l *lua.State, count int, callback func()) {
	lua.SetDebugHook(l, func(l *lua.State, ar lua.Debug) {
		callback()
	}, lua.MaskCount, count)
}

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
	defer func() {
		if r := recover(); r != nil {
			close(node.respond)
		}
	}()

	l := lua.NewState()

	world := <-node.resume

	addDirFunc(l, "grow", func(l *lua.State, d Direction) int {
		var state NewState
		state.Dir = d
		state.Operation = Move

		// Send a response and wait
		node.respond <- state
		world = <-node.resume

		return 0
	})

	addVoidFunc(l, "wait", func(l *lua.State) int {
		var state NewState
		state.Dir = Undef
		state.Operation = Wait

		node.respond <- state
		world = <-node.resume

		return 0
	})

	addDirFunc(l, "split", func(l *lua.State, d Direction) int {
		var state NewState
		state.Dir = d
		state.Operation = Split

		// Send a response and wait
		node.respond <- state
		world = <-node.resume

		return 0
	})

	addDirFunc(l, "lighting", func(l *lua.State, d Direction) int {
		l.PushNumber(world.Lighting[d])
		return 1
	})

	lua.LoadString(l, node.program)
	watchLuaThread(l, time.Duration(500)*time.Millisecond)
}

package sandbox

import (
	"fmt"
	"github.com/Shopify/go-lua"
	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
	"math/rand"
	"time"
)

type WorldState struct {
	Lighting map[Direction]float64
	Energy   int
	Age      int
}

type StateChange int

const (
	Move StateChange = iota
	Split
	Wait
	Spawn
)

type NewState struct {
	Operation StateChange
	Dir       Direction
	Meta      string
}

type Node struct {
	resume  chan<- WorldState
	respond <-chan NewState
	ctx     context.Context
}

func (n *Node) Update(state WorldState) (<-chan NewState, bool) {
	ok := true
	select {
	case n.resume <- state:
	case <-n.ctx.Done():
		ok = false
	}

	return n.respond, ok
}

// Kill the thread - returns a channel which will be fulfilled when the thread
// is dead
func (n *Node) Halt() <-chan struct{} {
	close(n.resume)
	return n.ctx.Done()
}

func AddNode(program string, meta string) *Node {
	// Make the communication channels
	resume := make(chan WorldState)
	respond := make(chan NewState)
	ctx, cancel := context.WithCancel(context.Background())

	n := Node{
		resume:  resume,
		respond: respond,
		ctx:     ctx,
	}

	in := internalNode{
		program: program,
		meta:    meta,
		cancel:  cancel,
		resume:  resume,
		respond: respond,
	}

	go runNode(in)
	return &n
}

type internalNode struct {
	program string
	meta    string
	cancel  func()
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

const duration time.Duration = 500 * time.Millisecond

func updateEndTime(t *time.Time) {
	*t = time.Now().Add(duration)
}

func watchLuaThread(l *lua.State, end_time *time.Time) {
	setLuaTimeoutHook(l, func() {
		if time.Now().After(*end_time) {
			panic("AAAAHHHH!!!")
		}
	})

	l.ProtectedCall(0, lua.MaskCall, 0)
}

func setLuaTimeoutHook(l *lua.State, callback func()) {
	lua.SetDebugHook(l, func(l *lua.State, ar lua.Debug) {
		callback()
	}, lua.MaskCount, 500)
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

func addDirStrFunc(l *lua.State, name string, fn func(*lua.State, Direction, string) int) {
	l.PushGoFunction(func(l *lua.State) int {
		argCount := l.Top()
		if argCount != 2 {
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

		s2, ok := l.ToString(2)
		if !ok {
			l.PushString("incorrect type of argument")
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
		} else {
			l.PushString("incorrect type of argument")
			l.Error()
			return 0
		}

		return fn(l, d, s2)
	})

	l.SetGlobal(name)
}

func addStrFunc(l *lua.State, name string, fn func(*lua.State, string) int) {
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

		return fn(l, s)
	})

	l.SetGlobal(name)
}

// blantently stolen from shopify's lua-go/libs.go
func openSafeLibs(l *lua.State, preloaded ...lua.RegistryFunction) {
	libs := []lua.RegistryFunction{
		{"_G", lua.BaseOpen},
		// {"package", PackageOpen},
		// {"coroutine", CoroutineOpen},
		{"table", lua.TableOpen},
		// {"io", IOOpen},
		// {"os", OSOpen},
		{"string", lua.StringOpen},
		// {"bit32", Bit32Open},
		{"math", lua.MathOpen},
		// {"debug", DebugOpen},
	}
	for _, lib := range libs {
		lua.Require(l, lib.Name, lib.Function, true)
		l.Pop(1)
	}
	lua.SubTable(l, lua.RegistryIndex, "_PRELOAD")
	for _, lib := range preloaded {
		l.PushGoFunction(lib.Function)
		l.SetField(-2, lib.Name)
	}
	l.Pop(1)
}

func respond(node *internalNode, state NewState) WorldState {
	// This could panic if the respond thread is closed. In this case, we want to
	// be killing the thread, so we're good!
	// Panicking here will unwind to the recover in runNode, which is what we want

	node.respond <- state
	st, ok := <-node.resume
	if !ok {
		panic("")
	}

	return st
}

func runNode(node internalNode) {
	defer func() {
		if r := recover(); r != nil {
			logrus.Info("Dead thread")
			node.cancel()
			close(node.respond)
		}
	}()

	l := lua.NewState()
	var end_time time.Time

	world := <-node.resume
	updateEndTime(&end_time)

	addDirFunc(l, "grow", func(l *lua.State, d Direction) int {
		var state NewState
		state.Dir = d
		state.Operation = Move

		world = respond(&node, state)
		updateEndTime(&end_time)
		return 0
	})

	addVoidFunc(l, "wait", func(l *lua.State) int {
		var state NewState
		state.Dir = Undef
		state.Operation = Wait

		world = respond(&node, state)
		updateEndTime(&end_time)
		return 0
	})

	addVoidFunc(l, "terminate", func(l *lua.State) int {
		updateEndTime(&end_time)
		panic("this is normal, apparently")
	})

	addVoidFunc(l, "spawn", func(l *lua.State) int {
		var state NewState
		state.Dir = Undef
		state.Operation = Spawn

		world = respond(&node, state)
		updateEndTime(&end_time)
		panic("Goodbye, thread")

		return 0
	})

	addVoidFunc(l, "get_energy", func(l *lua.State) int {
		l.PushInteger(world.Energy)
		return 1
	})

	addVoidFunc(l, "get_age", func(l *lua.State) int {
		l.PushInteger(world.Age)
		return 1
	})

	addDirStrFunc(l, "split", func(l *lua.State, d Direction, s string) int {
		var state NewState
		state.Dir = d
		state.Meta = s
		state.Operation = Split

		world = respond(&node, state)
		updateEndTime(&end_time)
		return 0
	})

	addDirFunc(l, "lighting", func(l *lua.State, d Direction) int {
		l.PushNumber(world.Lighting[d])
		return 1
	})

	addVoidFunc(l, "ran_dir", func(l *lua.State) int {
		num := rand.Float64()
		var dir string
		if num < 0.25 {
			dir = "left"
		} else if num < 0.5 {
			dir = "right"
		} else if num < 0.75 {
			dir = "up"
		} else {
			dir = "down"
		}

		l.PushString(dir)
		return 1
	})

	// Include the meta string property from the other context
	addVoidFunc(l, "meta", func(l *lua.State) int {
		l.PushString(node.meta)
		return 1
	})

	addStrFunc(l, "debug", func(l *lua.State, s string) int {
		fmt.Println(s)
		return 0
	})

	lua.LoadString(l, node.program)
	openSafeLibs(l)
	watchLuaThread(l, &end_time)
}

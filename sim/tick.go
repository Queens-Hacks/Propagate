package sim

import (
	"encoding/json"
	"github.com/Queens-Hacks/Propagate/sandbox"
	"github.com/Sirupsen/logrus"
	"time"
)

type MarshalledState struct {
	State []byte
	Diff  []byte
}

// After calling this function it is no longer safe to do anything with s from
// outside of the simulation
func (s *State) StartSimulate() <-chan MarshalledState {
	ch := make(chan MarshalledState)

	go func() {
		// We want to run the thing every 500 milliseconds
		tick := time.NewTicker(time.Millisecond * 500)

		for {
			s.simulateTick()

			md, err := json.Marshal(s.diff)
			if err != nil {
				logrus.Fatal(err)
			}
			ms, err := json.Marshal(s.state)
			if err != nil {
				logrus.Fatal(err)
			}

			ch <- MarshalledState{md, ms}
			<-tick.C
		}
	}()

	return ch
}

func (s *State) mkWorldState(_ *growthRoot) sandbox.WorldState {
	var ws sandbox.WorldState

	ws.Lighting[sandbox.Left] = 0
	ws.Lighting[sandbox.Right] = 0
	ws.Lighting[sandbox.Up] = 0
	ws.Lighting[sandbox.Down] = 0

	return ws
}

func (s *State) applyChanges(root *growthRoot, in sandbox.NewState) {
	new := root.Loc

	if in.MoveDir == sandbox.Left {
		new.X -= 1
	} else if in.MoveDir == sandbox.Right {
		new.X += 1
	} else if in.MoveDir == sandbox.Up {
		new.Y -= 1
	} else if in.MoveDir == sandbox.Down {
		new.Y += 1
	} else {
		// Super sketchy way to represent do nothing?
		return
	}

	// Can't move there, it's out of bounds!
	if new.Y < 0 || new.Y > s.Height() {
		logrus.Info("newY out of bounds")
		return
	}
	if new.X < 0 || new.X > s.Width() {
		logrus.Info("newY out of bounds")
		return
	}

	s.SetTile(new, Tile{PlantTile, &plantInfo{
		PlantId: root.PlantId,
		Parent:  root.Loc,
		Age:     0,
	}})

	// XXX Should this go through a method rather than direct mutation?
	// Move the growth root to the new location
	root.Loc = new
}

type newStateInfo struct {
	ch   <-chan sandbox.NewState
	root *growthRoot
}

// This is called by a timer every n time units
func (s *State) simulateTick() {
	responses := make([]newStateInfo, len(s.state.roots))

	// Tell each root to run until the next move operation
	for i := range s.state.roots {
		root := s.state.roots[i]
		ch := root.node.Update(s.mkWorldState(root))
		responses[i] = newStateInfo{ch, root}
	}

	for _, response := range responses {
		newState := <-response.ch
		s.applyChanges(response.root, newState)
	}
}

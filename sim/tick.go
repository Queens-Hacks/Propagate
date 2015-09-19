package sim

import (
	"encoding/json"
	"time"

	"github.com/Queens-Hacks/Propagate/sandbox"
	"github.com/Sirupsen/logrus"
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

			if !s.diff.isEmpty() {
				md, err := json.Marshal(s.diff)
				if err != nil {
					logrus.Fatal(err)
				}
				ms, err := json.Marshal(s.state)
				if err != nil {
					logrus.Fatal(err)
				}

				spores := []spore{}
				if s.diff.Spores != nil {
					spores = s.diff.Spores
				}

				s.diff = diff{[]tileDiff{}, map[string]*Plant{}, []string{}, spores}
				ch <- MarshalledState{ms, md}
			}

			<-tick.C
		}
	}()

	return ch
}

func (s *State) mkWorldState(_ *growthRoot) sandbox.WorldState {
	var ws sandbox.WorldState
	ws.Lighting = map[sandbox.Direction]float64{}

	ws.Lighting[sandbox.Left] = 0
	ws.Lighting[sandbox.Right] = 0
	ws.Lighting[sandbox.Up] = 0
	ws.Lighting[sandbox.Down] = 0

	return ws
}

func MoveGrowthNodeToLocation(loc Location, in sandbox.NewState, mh int, mw int) Location {
	new := loc

	if in.Dir == sandbox.Left {
		new.X -= 1
	} else if in.Dir == sandbox.Right {
		new.X += 1
	} else if in.Dir == sandbox.Up {
		new.Y -= 1
	} else if in.Dir == sandbox.Down {
		new.Y += 1
	} else {
		return loc
	}

	// Can't move there, it's out of bounds!
	if new.Y < 0 || new.Y > mh {
		logrus.Info("newY out of bounds", new.Y)
		return loc
	}
	if new.X < 0 || new.X > mw {
		logrus.Info("newY out of bounds")
		return loc
	}

	return new // new looks good!
}

func (s *State) applyChanges(root *growthRoot, in sandbox.NewState) {
	new := root.Loc

	if in.Operation == sandbox.Move {
		new = MoveGrowthNodeToLocation(root.Loc, in, s.Height(), s.Width())
		s.SetTile(new, Tile{PlantTile, &plantInfo{
			PlantId: root.PlantId,
			Parent:  root.Loc,
		}})
	} else if in.Operation == sandbox.Split {
		tmp := MoveGrowthNodeToLocation(root.Loc, in, s.Height(), s.Width())
		s.SetTile(tmp, Tile{PlantTile, &plantInfo{
			PlantId: root.PlantId,
			Parent:  root.Loc,
		}})
		s.SetTile(root.Loc, Tile{PlantTile, &plantInfo{
			PlantId: root.PlantId,
			Parent:  root.Loc,
		}})
	} else if in.Operation == sandbox.Wait {
		// that was easy
		return
	} else {
		return
	}

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

	spores := []spore{}
	for _, p := range s.diff.Spores {
		// UpdateSpore returns true if it has planted the spore
		spawned := s.UpdateSpore(&p)
		// Unplanted spores are kept for next tick
		if !spawned {
			spores = append(spores, p)
		}
	}
	s.diff.Spores = spores
}

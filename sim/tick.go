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

				s.diff = diff{[]tileDiff{}, map[string]*Species{}, []string{}, spores}
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

func boundsCheck(loc Location, mh int, mw int) Location {
	// Can't move there, it's out of bounds!
	if loc.Y < 0 {
		logrus.Info("newY out of bounds", loc.Y)
		loc.Y = 0
	}

	if loc.Y >= mh {
		logrus.Info("newY out of bounds", loc.Y)
		loc.Y = mh - 1
	}

	// loop around loop
	if loc.X < 0 {
		loc.X = mw + loc.X
	} else if loc.X >= mw {
		loc.X = loc.X - mw
	}

	return loc
}

func (s *State) DirectionToLocation(loc Location, dir sandbox.Direction) Location {
	new := loc

	if dir == sandbox.Left {
		new.X -= 1
	} else if dir == sandbox.Right {
		new.X += 1
	} else if dir == sandbox.Up {
		new.Y -= 1
	} else if dir == sandbox.Down {
		new.Y += 1
	}

	new = boundsCheck(new, s.Height(), s.Width())

	return new // new looks good!
}

func (s *State) applyChanges(root *growthRoot, in sandbox.NewState) {
	new := root.Loc

	if in.Operation == sandbox.Move {
		new = s.DirectionToLocation(root.Loc, in.Dir)

		if s.GetTile(new).Type == PlantTile || s.GetTile(new).Type == DirtTile {
			return
		}

		s.SetTile(new, Tile{PlantTile, &extraTileInfo{
			root.SpeciesId,
			root.Loc,
			false,
			root.Plant,
		}})
	} else if in.Operation == sandbox.Split {
		tmp := s.DirectionToLocation(root.Loc, in.Dir)

		if s.GetTile(tmp).Type == PlantTile || s.GetTile(new).Type == DirtTile {
			return
		}

		s.SetTile(tmp, Tile{PlantTile, &extraTileInfo{
			root.SpeciesId,
			root.Loc,
			false,
			root.Plant,
		}})

		// Add the new root for the new plant
		s.AddGrowth(tmp, root.Plant, in.Meta)
	} else if in.Operation == sandbox.Wait {
		// that was easy
		return
	} else {
		logrus.Warn("Unrecognized Operation")
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
	responses := []newStateInfo{}
	deadNodes := []*growthRoot{}

	// Tell each root to run until the next move operation
	for root := range s.state.roots {
		ch, ok := root.node.Update(s.mkWorldState(root))
		if !ok {
			// We have to remove the node later
			deadNodes = append(deadNodes, root)
			continue
		}

		responses = append(responses, newStateInfo{ch, root})
	}

	for _, dead := range deadNodes {
		delete(s.state.roots, dead)
	}

	for _, response := range responses {
		newState, ok := <-response.ch
		if !ok {
			continue
		}
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

	surviving := make([]*Plant, 0, len(s.state.plants))
	for _, p := range s.state.plants {
		deltaEnergy := 0
		for _ = range p.tiles {
			deltaEnergy += 10
		}

		p.Energy += deltaEnergy
		p.Energy -= (p.Age * p.Age)

		if p.Energy < 0 {
			for t := range p.tiles {
				s.SetTile(t, Tile{AirTile, nil})
			}
			toKill := []*growthRoot{}
			for r := range s.state.roots {
				if r.Plant == p {
					toKill = append(toKill, r)
				}
			}
			for _, k := range toKill {
				s.HaltGrowth(k)
			}

		} else {
			surviving = append(surviving, p)
		}

		p.Age++
	}
	s.state.plants = surviving

}

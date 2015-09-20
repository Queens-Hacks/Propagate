package sim

import (
	"encoding/json"
	"math/rand"
	"time"

	"github.com/Queens-Hacks/Propagate/sandbox"
	"github.com/Sirupsen/logrus"
)

type Action struct {
	Kind    string `json:"kind"`
	Code    string `json:"code"`
	Color   int    `json:"color"`
	Species string `json:"species"`
	X       int    `json:"x"`
	Y       int    `json:"y"`
}

type MarshalledState struct {
	State []byte
	Diff  []byte
}

func (s *State) StopSimulate() {
	// XXX TODO IMPLEMENT
}

// After calling this function it is no longer safe to do anything with s from
// outside of the simulation
func (s *State) StartSimulate(actions <-chan Action) <-chan MarshalledState {
	ch := make(chan MarshalledState)

	go func() {
		// We want to run the thing every 500 milliseconds
		tick := time.NewTicker(time.Millisecond * 200)
		defer tick.Stop()

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

			for {
				select {
				case <-tick.C:
				case action := <-actions:
					s.handleAction(&action)
					continue
				}
				break
			}
		}
	}()

	return ch
}

func (s *State) handleAction(a *Action) {
	if a.Kind == "+species" {
		color := a.Color
		if color < 0 {
			logrus.Warn("Color too small")
			color = 0
		}
		if color > 360 {
			logrus.Warn("Color too big")
			color = 360
		}
		code := a.Code
		if len(code) == 0 {
			logrus.Warn("Ignored empty code property")
			return
		}
		s.AddSpecies(color, code, "")
		// XXX Maybe send the species back to the client somehow? who knows...
	} else if a.Kind == "+spawn" {
		_, ok := s.GetSpecies(a.Species)
		if !ok {
			logrus.Warn("Non-existant species")
		}
		loc := s.Clamp(&Location{a.X, a.Y})
		s.AddSpore(loc, a.Species)
	} else if a.Kind == "+species+spawn" {
		color := a.Color
		if color < 0 {
			logrus.Warn("Color too small")
			color = 0
		}
		if color > 360 {
			logrus.Warn("Color too big")
			color = 360
		}
		code := a.Code
		if len(code) == 0 {
			logrus.Warn("Ignored empty code property")
			return
		}
		species := s.AddSpecies(color, code, "")

		for i := 0; i < 10; i++ {
			loc := Location{rand.Intn(s.Width()), 0}
			logrus.Info(loc)
			s.AddSpore(loc, species)
			// XXX Instantly plant them sometimes
		}
	} else {
		logrus.Warnf("Unrecognized kind %s", a.Kind)
	}
}

func (s *State) mkWorldState(rt *growthRoot) sandbox.WorldState {
	var ws sandbox.WorldState
	ws.Lighting = map[sandbox.Direction]float64{}

	ws.Lighting[sandbox.Left] = 0
	ws.Lighting[sandbox.Right] = 0
	ws.Lighting[sandbox.Up] = 0
	ws.Lighting[sandbox.Down] = 0

	ws.Energy = rt.Plant.Energy

	ws.Age = rt.Plant.Age

	return ws
}

func (s *State) Clamp(loc *Location) Location {
	// Can't move there, it's out of bounds!
	if loc.Y < 0 {
		// logrus.Info("newY out of bounds", loc.Y)
		loc.Y = 0
	}

	if loc.Y >= s.Height() {
		// logrus.Info("newY out of bounds", loc.Y)
		loc.Y = s.Height() - 1
	}

	// loop around loop
	if loc.X < 0 {
		loc.X = s.Width() + loc.X
	} else if loc.X >= s.Width() {
		loc.X = loc.X - s.Width()
	}

	return *loc
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

	s.Clamp(&new)

	return new // new looks good!
}

func (s *State) applyChanges(root *growthRoot, in sandbox.NewState) {
	new := root.Loc

	energy := root.Plant.Energy

	if in.Operation == sandbox.Move && energy > 200 {
		root.cache = nil
		root.Plant.Energy -= 50
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
	} else if in.Operation == sandbox.Split && energy > 200 {
		root.cache = nil
		root.Plant.Energy -= 75
		tmp := s.DirectionToLocation(root.Loc, in.Dir)

		if s.GetTile(tmp).Type == PlantTile || s.GetTile(tmp).Type == DirtTile {
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
	} else if in.Operation == sandbox.Spawn {
		s.AddSpore(root.Loc, root.SpeciesId)
	} else if in.Operation == sandbox.Wait {
		// that was easy
		return
	} else {
		// logrus.Warn("Unrecognized Operation")
		root.cache = &in
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
		if root.cache != nil {
			continue
		}
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

	for r := range s.state.roots {
		if r.cache != nil {
			s.applyChanges(r, *r.cache)
		}
	}

	spores := []spore{}
	for _, p := range s.diff.Spores {
		// UpdateSpore returns true if it has planted the spore
		spawned := s.UpdateSpore(&p)
		// Unplanted spores are kept for next tick
		if spawned {
			s.plantRelease(p.SpeciesId)
		} else {
			spores = append(spores, p)
		}
	}
	s.diff.Spores = spores

	surviving := make([]*Plant, 0, len(s.state.plants))
	// logrus.Infof("len of plants: %d", len(s.state.plants))
	for _, p := range s.state.plants {
		deltaEnergy := 0
		factor := 1000.0
		for _ = range p.tiles {
			deltaEnergy += int(float64(2+p.Luck) * (factor / 1000))
			factor = (factor * (.5))
			// logrus.Infof("factor %f")
		}

		deltaEnergy -= (p.Age * p.Age) / 10000
		p.Energy += deltaEnergy

		// logrus.Infof("delta energy: %d", deltaEnergy)
		if p.Energy < 0 {
			for r := range p.roots {
				s.HaltGrowth(r)
			}
			s.plantRelease(p.SpeciesId)
			for _, t := range p.tiles {
				s.SetTile(t, Tile{AirTile, nil})
			}
			for i := 0; i < 2; i++ {
				s.AddSpore(Location{rand.Intn(500), 75}, p.SpeciesId)
			}
		} else {
			surviving = append(surviving, p)
		}
		p.Age++
	}

	s.state.plants = surviving
}

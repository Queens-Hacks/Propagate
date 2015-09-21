package main

import (
	"math/rand"

	"github.com/Queens-Hacks/Propagate/sim"
	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

var maxSpiral string = `
function foo (n)
	  i = 0
	  while i<n do
          i= i+1
	  grow(getdir(n))
	  end

      if n > 0 then return foo(n - 1) end
end

function getdir(n)
   v = n % 4
   if v == 0 then return "up" end
   if v == 1 then return "right" end
   if v == 2 then return "down" end
   if v == 3 then return "left" end

end

grow("up")
grow("up")
grow("up")

foo(8);
`
var maxMeander string = `
while 1 do

	while math.random(10)<8 do
		grow("up")
	end
	while math.random(10)<8 do
		grow("left")
	end
	while math.random(10)<8 do
		grow("up")
	end
	while math.random(10)<8 do
		grow("right")
	end

end

`
var jakeRand string = `
local i = 10
while i > 0 do
  grow("up")
  grow("up")
  if math.random(2) == 1 then
     grow("left")
     grow("left")
     split("left", "left")
  else
     grow("right")
     grow("right")
     split("right", "right")
  end
end
`

var maxMemory string = `
while 1 do
	grow("up")
	grow("up")
	split("up","up")
	grow("left")
	grow("left")
	split("left","left")
	grow("right")
	grow("right")
	split("right","right")
end
`

var crystal string = `
while 1 do

  grow("up")
  grow("up")
  grow("up")
  grow("up")

  split("up", "right")

  split("left", "left")
  split("right", "right")
end
`
var coral string = `

while 1 do
n=0
 while math.random(10)>n do
  grow("left")
  grow("up")
  n= n+1
end
n = 0
while math.random(10)>n do
  grow("right")
  grow("up")
n= n+1
end
end

`

var twistyLeft string = `
while 1 do
  grow("left")
  grow("up")
  grow("up")
end
`

var twistyUp string = `
while 1 do
  grow("left")
  grow("up")
  grow("up")
  grow("left")
  grow("up")
  grow("up")
  grow("right")
  grow("up")
  grow("up")
  grow("right")
  grow("up")
  grow("up")
end
`

var upUp string = `
while 1 do
  grow("up")
end
`

var fearnLeft string = `
while 1 do
  if meta() == "" then
    grow("up")
    grow("up")
	split("left", "left")
  else
  	grow("left")
  end
end

`

func main() {

	logrus.SetLevel(logrus.WarnLevel)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	total := make(chan []byte)
	diff := make(chan []byte)
	actions := make(chan sim.Action)

	s := sim.NewState(500, 125)

	spawnHeight := 50

	species := s.AddSpecies(256, maxMeander, "Me")
	for i := 0; i < 10; i++ {
		s.AddSpore(sim.Location{rand.Intn(500), spawnHeight}, species)
	}

	species = s.AddSpecies(244, jakeRand, "Me")
	for i := 0; i < 10; i++ {
		s.AddSpore(sim.Location{rand.Intn(500), spawnHeight}, species)
	}

	species = s.AddSpecies(128, upUp, "Me")
	for i := 0; i < 10; i++ {
		s.AddSpore(sim.Location{rand.Intn(500), spawnHeight}, species)
	}

	species = s.AddSpecies(63, crystal, "Me")
	for i := 0; i < 10; i++ {
		s.AddSpore(sim.Location{rand.Intn(500), spawnHeight}, species)
	}

	species = s.AddSpecies(44, maxMemory, "Me")
	for i := 0; i < 10; i++ {
		s.AddSpore(sim.Location{rand.Intn(500), spawnHeight}, species)
	}

	species = s.AddSpecies(14, fearnLeft, "Me")
	for i := 0; i < 10; i++ {
		s.AddSpore(sim.Location{rand.Intn(500), spawnHeight}, species)
	}

	species = s.AddSpecies(1, coral, "Me")
	for i := 0; i < 10; i++ {
		s.AddSpore(sim.Location{rand.Intn(500), spawnHeight}, species)
	}

	species = s.AddSpecies(275, twistyUp, "Me")
	for i := 0; i < 10; i++ {
		s.AddSpore(sim.Location{rand.Intn(500), spawnHeight}, species)
	}

	species = s.AddSpecies(300, twistyLeft, "Me")
	for i := 0; i < 10; i++ {
		s.AddSpore(sim.Location{rand.Intn(500), spawnHeight}, species)
	}

	port := ":4444"

	logrus.Infof("Listening on port %s", port)
	go New(ctx, total, diff, actions, port)

	ss := s.StartSimulate(actions)
	for {
		ms := <-ss
		total <- ms.State
		diff <- ms.Diff
	}
}

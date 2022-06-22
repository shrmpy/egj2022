package polarity

import (
	"encoding/json"
)

func (m *Maze) eval(t Ticket) Delta {
	if dead(t.owner.inventory) {
		return deltaCorpse(t.owner)
	}
	// The ticket.move field is user input and must be checked.
	var move Move
	if err := json.Unmarshal([]byte(t.move), &move); err != nil {
		return unresolved(t.owner)
	}
	switch move.Command {
	case "walk":
		m.walk(t.owner, move.Direction)
	}

	return unresolved(t.owner)
}
func (m *Maze) walk(j Job, dir string) Delta {
	// TODO
	// - extract xy position
	// - is direction blocked?
	// - ifnot then modify position in job's state
	//   and set grid mask at xy position
	//   and unset grid mask from old xy position
	return Delta{}
}
func unresolved(j Job) Delta {
	// TODO capture error
	// TODO indicate unsuccessful request
	return Delta{
		inv: j.inventory,
		fow: j.fow,
		x:   j.x,
		y:   j.y,
	}
}
func dead(inv map[Kit]int) bool {
	return inv[Battery] <= 0
}

// move request data from the ticket JSON
type Move struct {
	Command   string `json:"command"`
	Direction string `json:"direction"`
}

// job fields outcome from transition
type Delta struct {
	inv  map[Kit]int
	fow  [][]Mask
	x, y int
}

func deltaCorpse(j Job) Delta {
	i := make(map[Kit]int)
	i[Battery] = -8
	return Delta{
		x: j.x,
		y: j.y,
	}
}

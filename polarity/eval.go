package polarity

import (
	"encoding/json"
)

func (m *Maze) eval(t Ticket) (map[Kit]int, [][]Mask) {
	// The ticket.move field is user input and must be checked.
	var move Move
	if err := json.Unmarshal([]byte(t.move), &move); err != nil {
		return unresolved(t.owner)
	}
	switch move.Command {
	case "walk":
		walk(m, t.owner, move.Direction)
	}

	return unresolved(t.owner)
}
func walk(m *Maze, j Job, dir string) {
	// TODO
	// - extract xy position
	// - is direction blocked?
	// - ifnot then modify position in job's state
	//   and set grid mask at xy position
	//   and unset grid mask from old xy position
}
func unresolved(j Job) (map[Kit]int, [][]Mask) {
	// TODO capture error
	// TODO indicate unsuccessful request
	return j.inventory, j.fow
}

// move request data from the ticket JSON
type Move struct {
	Command   string
	Direction string
}

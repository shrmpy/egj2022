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
		return m.walk(t.owner, move.Direction)
	}

	return unresolved(t.owner)
}
func (m *Maze) walk(j Job, dir string) Delta {
	// TODO
	// - ifnot then modify position in job's state
	//   and set grid mask at xy position
	//   and unset grid mask from old xy position
	row, col := j.row, j.col
	switch dir {
	case "n", "north":
		if row > 0 {
			row--
		}
	case "s", "south":
		if row < (m.width - 1) {
			row++
		}
	case "e", "east":
		if col < (m.width - 1) {
			col++
		}
	case "w", "west":
		if col > 0 {
			col--
		}
	}
	if blocked(row, col, m.grid) {
		return Delta{
			inv: j.inventory,
			fog: j.fog,
			row: j.row,
			col: j.col,
		}
	}

	//TODO sync fog mask
	return Delta{
		inv: j.inventory,
		fog: j.fog,
		row: row,
		col: col,
	}
}
func unresolved(j Job) Delta {
	// TODO capture error
	// TODO indicate unsuccessful request
	return Delta{
		inv: j.inventory,
		fog: j.fog,
		row: j.row,
		col: j.col,
	}
}
func blocked(row, col int, grid [][]Mask) bool {
	cell := grid[row][col]
	if cell.Has(Barrier) || cell.Has(Robot) {
		return true
	}
	return false
}

// empty battery means death
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
	inv      map[Kit]int
	fog      [][]Mask
	row, col int
}

// delta for corpse (dead robot)
func deltaCorpse(j Job) Delta {
	inv := make(map[Kit]int)
	inv[Battery] = -8
	return Delta{
		inv: inv,
		row: j.row,
		col: j.col,
	}
}

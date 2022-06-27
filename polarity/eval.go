package polarity

import (
	//"fmt"
	"encoding/json"
)

// evaluate (jaeger) script and construct the outcome (inventory,fog)
func (m *Maze) eval(t Ticket) Delta {
	if grounded(t.owner.inventory) {
		return m.junk(t.owner)
	}
	// The ticket.move field is user input and must be checked.
	var move Move
	if err := json.Unmarshal([]byte(t.move), &move); err != nil {
		return unresolved(t.owner)
	}
	switch move.Command {
	case "walk":
		return m.walk(t.owner, move.Direction)
	default:
		m.hist("DEBUG exceptional move, " + move.Command)
	}

	return unresolved(t.owner)
}

// change position in the specified direction
func (m *Maze) walk(j Job, dir string) Delta {
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
	if blocked(row, col, m.mini) {
		////m.hist(fmt.Sprintf("DEBUG blocked, %s (r%d, c%d) ", j.name, row, col))
		// position unchanged
		return Delta{
			inv: j.inventory,
			fog: j.fog,
			row: j.row,
			col: j.col,
		}
	}

	d := Delta{
		inv: j.inventory,
		row: row,
		col: col,
	}
	// mask position in maze
	m.mini.Walk(j, d)
	// sync fog cells with position
	var fog Minimap
	fog = j.fog //todo copy
	fog.WalkMe(j, d)
	d.fog = fog

	return d
}

// jaeger is junk
func (m *Maze) junk(j Job) Delta {
	inv := make(map[Kit]int)
	inv[Battery] = -8
	delta := Delta{
		inv: inv,
		row: j.row,
		col: j.col,
	}
	// mask mini pos as junk
	m.mini.Junk(j, delta)
	return delta
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

// check if the destination is blocked
func blocked(row, col int, mm Minimap) bool {
	// TODO prevented walking into walls by calculation, but double-check
	cell := mm[row][col]
	if cell.Has(Barrier) || cell.Has(Jaeger) {
		return true
	}
	return false
}

// empty battery means incapacitated
func grounded(inv map[Kit]int) bool {
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
	fog      Minimap
	row, col int
}

package polarity

import (
	"encoding/json"
	"fmt"
)

// evaluate (jaeger) script and construct the outcome (inventory,fog)
// this is called by the listen goroutine and should be the only
// thread in order to be safe when modifying the maps
func (m *Maze) eval(t Ticket) Delta {
	if grounded(t.owner.inv) {
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
	case "ping":
		return m.ping(t.owner, move.Direction)
	default:
		m.hist("DEBUG exceptional move, %s", move.Command)
	}

	return unresolved(t.owner)
}

// change position in the specified direction
func (m *Maze) walk(j Job, dir string) Delta {
	kv := make(map[string]string)
	row, col, err := posFrom(j.row, j.col, dir, m.width)
	if err != nil {
		m.hist("DEBUG %s", err.Error())
		return unresolved(j)
	}
	if blocked(row, col, m.mini) {
		m.hist("DEBUG blocked, %s (r%d, c%d) ", j.name, row, col)
		return unresolved(j)
	}
	kv["last"] = "walk-" + dir
	d := Delta{
		row: row,
		col: col,
		kv:  kv,
	}
	// mask position in maze
	m.mini.Walk(j, d)

	d.inv = j.inv.Copy()
	// sync fog cells with position
	d.fog = j.fog.Copy()
	d.fog.WalkMe(j, d)

	return d
}

// scan in the specified direction
func (m *Maze) ping(j Job, dir string) Delta {
	// TODO Clone
	fog := j.fog.Copy()
	kv := make(map[string]string)
	lr, lc := j.row, j.col
	switch dir {
	case "w", "west":
		if j.col <= 0 {
			return unresolved(j)
		}
		kv["last"] = "ping-west"
		lc = 0
		for i := j.col; i > 0; i-- {
			cell := m.mini[j.row][i-1]
			if cell.Has(Barrier | Junk | Jaeger | Fusion | ToggleSwitch) {
				fog[j.row][i-1].Set(cell)
				break
			}
		}
	}
	d := Delta{
		fog: fog,
		row: j.row,
		col: j.col,
		inv: j.inv.Copy(),
		kv:  kv,
	}
	// add animation to maze
	////m.mini.Ping(j, d)
	m.loop(Ping, j.row, j.col, lr, lc)
	return d
}

// jaeger is junk
func (m *Maze) junk(j Job) Delta {
	inv := newInv()
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
	return Delta{
		inv: j.inv.Copy(),
		fog: j.fog.Copy(),
		row: j.row,
		col: j.col,
	}
}

// check if the destination is blocked
func blocked(row, col int, mm Minimap) bool {
	cell := mm[row][col]
	if cell.Has(Barrier) || cell.Has(Jaeger) {
		return true
	}
	return false
}

// empty battery means incapacitated
func grounded(i Inventory) bool {
	return i[Battery] <= 0
}
func posFrom(row, col int, dir string, wd int) (int, int, error) {
	switch dir {
	case "n", "north":
		return northFrom(row, col)
	case "s", "south":
		return southFrom(row, col, wd)
	case "e", "east":
		return eastFrom(row, col, wd)
	case "w", "west":
		return westFrom(row, col)
	}
	return 0, 0, fmt.Errorf("FAIL direction, unknown")
}
func northFrom(row, col int) (int, int, error) {
	r := row - 1
	if r < 0 {
		return row, col, fmt.Errorf("north maze boundary")
	}
	return r, col, nil
}
func southFrom(row, col, wd int) (int, int, error) {
	r := row + 1
	if r >= wd {
		return row, col, fmt.Errorf("south maze boundary")
	}
	return r, col, nil
}
func eastFrom(row, col, wd int) (int, int, error) {
	c := col + 1
	if c >= wd {
		return row, col, fmt.Errorf("east maze boundary")
	}
	return row, c, nil
}
func westFrom(row, col int) (int, int, error) {
	c := col - 1
	if c < 0 {
		return row, col, fmt.Errorf("west maze boundary")
	}
	return row, c, nil
}

// move request data from the ticket JSON
type Move struct {
	Command   string `json:"command"`
	Direction string `json:"direction"`
}

// job fields outcome from transition
type Delta struct {
	inv      Inventory
	fog      Minimap
	row, col int
	kv       map[string]string
}

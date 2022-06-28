package polarity

import (
	"fmt"
)

type Job struct {
	script   Script
	inv      Inventory
	fog      Minimap
	name     string
	row, col int
}

func (j *Job) state() string {
	// marshal/serialize for passing across interfaces
	return ""
}

// values for debug print
func (j *Job) String() string {
	return fmt.Sprintf("%s: b%d (c%d, r%d)",
		j.name,
		j.inv[Battery],
		j.col, j.row,
	)
}

// new job instance (to be in next round)
func newJob(t Ticket, d Delta) Job {
	return Job{
		script: t.owner.script,
		name:   t.owner.name,
		inv:    d.inv.Copy(),
		fog:    d.fog.Copy(),
		row:    d.row,
		col:    d.col,
	}
}

// initial jaeger instance (todo parameterize/setter script)
func newJaeger(row, col, wd int, name string) Job {
	j := Job{
		script: TestScript{},
		name:   name,
		inv:    newInv(),
		row:    row,
		col:    col,
	}
	fog := NewMinimap(wd)
	fog.JaegerMe(j)
	j.fog = fog
	return j
}
func newInv() Inventory {
	m := make([]int, 4)
	m[Battery] = 10
	m[Build] = 5000
	m[Cannon] = 5000
	m[Scanner] = 5000
	return m
}

// TODO load wasm script
type TestScript struct{}

func (s TestScript) Next(state string) string {
	// TODO json.Unmarshal(state, &job)
	return `{"command":"walk", "direction": "north"}`
}

type Inventory []int

func (i Inventory) Copy() Inventory {
	ni := make([]int, 4)
	copy(ni, i)
	return ni
}

type Kit uint8

const (
	Battery Kit = iota
	Cannon
	Build
	Scanner
)

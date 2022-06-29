package polarity

import (
	"encoding/json"
	"fmt"
)

type Job struct {
	script   Script
	inv      Inventory
	fog      Minimap
	name     string
	row, col int
	kv       map[string]string
}

// marshal for Next
func (j *Job) state() string {
	se := ScriptEnv{
		Name: j.name,
		Row:  j.row,
		Col:  j.col,
		Mini: j.fog,
		Inv:  j.inv,
		KV:   j.kv,
	}
	//TODO minimize data necessary
	js, err := json.Marshal(se)
	if err != nil {
		//TODO
		return ""
	}
	return string(js)
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
		kv:     d.kv,
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

type ScriptEnv struct {
	Name string            `json:"name"`
	Row  int               `json:"row"`
	Col  int               `json:"column"`
	Mini Minimap           `json:"mini"`
	Inv  Inventory         `json:"inventory"`
	KV   map[string]string `json:"kv"`
}

// TODO load wasm script
type TestScript struct{}

func (s TestScript) Next(state string) string {
	//TODO kv that is preserved to next cycle
	var se ScriptEnv
	if err := json.Unmarshal([]byte(state), &se); err != nil {
		//TODO
		return `{"command":"walk", "direction": "east"}`
	}
	// - check KV for notes from prev cycle
	// - when none, prioritize options (evade, heal, snipe)

	// check battery level
	if se.Inv[Battery] < 10 {
		// todo being attacked?
		// KV[priority] = "evade"
		return `{"command":"walk", "direction": "south"}`
	}
	// anything west?
	if se.Col > 0 {
		for i := 0; i < se.Col; i++ {
			cell := se.Mini[se.Row][i]
			if cell.Has(Fusion | Junk | Jaeger | ToggleSwitch) {
				return `{"command":"walk", "direction": "west"}`
			}
		}
		last := ""
		if se.KV != nil {
			if val, ok := se.KV["last"]; ok {
				last = val
			}
		}
		if last != "ping-west" {
			// did we ping-west last round?
			return `{"command":"ping", "direction": "west"}`
		}
	}

	return `{"command":"walk", "direction": "north"}`
}

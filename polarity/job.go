package polarity

import "fmt"

type Job struct {
	script    Script
	inventory map[Kit]int
	fog       Minimap
	name      string
	row, col  int
}

func (j *Job) state() string {
	// marshal/serialize for passing across interfaces
	return ""
}

// values for debug print
func (j *Job) String() string {
	return fmt.Sprintf("%s: b%d (c%d, r%d)",
		j.name,
		j.inventory[Battery],
		j.col, j.row,
	)
}

// reset to empty job queue
func cleanJobs() map[string]Job {
	return make(map[string]Job)
}

// new job instance
func newJob(t Ticket, d Delta) Job {
	return Job{
		script:    t.owner.script,
		name:      t.owner.name,
		inventory: d.inv,
		fog:       d.fog,
		row:       t.owner.row,
		col:       t.owner.col,
	}
}

// initial jaeger instance (todo parameterize/setter script)
func newJaeger(row, col, wd int, name string) Job {
	j := Job{
		script:    TestScript{},
		name:      name,
		inventory: newInventory(),
		row:       row,
		col:       col,
	}
	fog := NewMinimap(wd)
	fog.JaegerMe(j)
	j.fog = fog
	return j
}
func newInventory() map[Kit]int {
	m := make(map[Kit]int)
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

// job queue post-processing step
func nextJobs(m *Maze, next map[string]Job) {
	// TODO _Lock_ maze map/slice, atm we are called in Update (not child thread)
	// only maze.Update modifies the jobs map here
	// enqueue jobs to be processed next-cycle

	for k := range m.jobs {
		delete(m.jobs, k)
	}
	for k, v := range next {
		m.jobs[k] = v
		delete(next, k)
	}
	next = nil
}

type Kit uint8

const (
	Battery Kit = iota
	Cannon
	Build
	Scanner
)

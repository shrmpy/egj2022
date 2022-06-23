package polarity

//import "fmt"

type Job struct {
	script    Script
	inventory map[Kit]int
	fog       [][]Mask
	name      string
	row, col  int
}

func (j *Job) state() string {
	// marshal/serialize for passing across interfaces
	return ""
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

// initial robot instance
func newRobot(row, col, wd int, name string) Job {
	j := Job{
		script:    TestScript{},
		name:      name,
		inventory: newInventory(),
		row:       row,
		col:       col,
	}
	fog := NewGrid(wd)
	fog.RobotMe(j)
	j.fog = fog
	return j
}
func newInventory() map[Kit]int {
	m := make(map[Kit]int)
	m[Battery] = 10
	m[Build] = 5000
	m[Phaser] = 5000
	m[Scanner] = 5000
	return m
}

// TODO load wasm script
type TestScript struct{}

func (s TestScript) Next(state string) string {
	return `{"walk": "north"}`
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
	Phaser
	Build
	Scanner
)

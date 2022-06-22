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

// post-process job queue
func nextJobs(m *Maze, next map[string]Job) {
	// only maze.Update modifies the jobs map here
	// enqueue jobs to be processed next-cycle
	// (m.jobs = m.jobs[:0])
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

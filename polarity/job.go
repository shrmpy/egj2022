package polarity

//import "fmt"

type Job struct {
	script    Script
	inventory map[Kit]int
	fow       [][]Mask
	name      string
	x, y      int
}

func (j *Job) state() string {
	// serialize for passing across interfaces
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
		fow:       d.fow,
		x:         t.owner.x,
		y:         t.owner.y,
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

package polarity

//import "fmt"

type Job struct {
	script    Script
	inventory map[Kit]int
	fow       [][]Mask
	name      string
}

func (j *Job) state() string {
	// serialize for passing across interfaces
	return ""
}

// reset to empty job queue
func newJobs(sz int) map[string]Job {
	return make(map[string]Job, sz)
}

// new job instance
func newJob(t Ticket, i map[Kit]int, fow [][]Mask) Job {
	return Job{
		script:    t.owner.script,
		name:      t.owner.name,
		inventory: i,
		fow:       fow,
	}
}

// post-process job queue
func nextJobs(m *Maze, next map[string]Job) {
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

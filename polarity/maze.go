package polarity

import (
	"context"
	"fmt"
	"sync"
	"time"
)
import "golang.org/x/sync/errgroup"

// maze acts as world state
type Maze struct {
	jobs map[string]Job
	grid [][]Mask
}

// *wasm* a user defined script
type Script interface {
	Next(state string) string
}
type Ticket struct {
	move  string
	owner Job
}

func NewMaze(wd int) *Maze {
	// TODO populate the maze (and load robot scripts)

	rows := make([][]Mask, wd)
	for i := range rows {
		rows[i] = make([]Mask, wd)
	}

	return &Maze{
		grid: rows,
	}
}

// to be called by the game lifecycle
func (m *Maze) Update() error {
	// TODO shuffle order of workers
	if m.Done() {
		return fmt.Errorf("Simulation done.")
	}
	acc := make(chan Ticket)
	next := cleanJobs()
	go m.listen(next, acc)

	cx, cancel := context.WithTimeout(context.Background(), groupTurnMs)
	defer cancel()
	grp, ctx := errgroup.WithContext(cx)
	for _, j := range m.jobs {
		spawn(j, acc, grp, ctx, timeSliceMs)
	}
	if err := grp.Wait(); err != nil {
		// TODO not all errors are equal
		close(acc)
		return err
	}
	// end the for-range
	close(acc)
	nextJobs(m, next)
	return nil
}

func (m *Maze) Done() bool {
	// TODO more halt conditions
	if len(m.jobs) <= 1 {
		// work queue depleted
		return true
	}
	return false
}

// rcv (action) tickets and convert into item on next-job queue
func (m *Maze) listen(next map[string]Job, inch <-chan Ticket) {
	// only this thread tries to access 'next' map
	for t := range inch {
		delta := m.eval(t)
		if dead(delta.inv) {
			// mask grid pos as corpse, no job next-cycle
			m.grid[delta.y][delta.x].Set(Corpse)
		} else {
			next[t.owner.name] = newJob(t, delta)
		}
	}
}

// exec script and returns ticket for next (robot) action
func spawn(j Job, outch chan<- Ticket, g *errgroup.Group, ctx context.Context, ms time.Duration) {
	var (
		once    sync.Once
		running = false
		ticker  = time.NewTicker(1 * time.Millisecond)
	)
	g.Go(func() error {
		defer ticker.Stop()
		select {
		case <-ctx.Done():
			// group cancelled
			return ctx.Err()
		case <-ticker.C:
			ticker.Stop()
			if running {
				return fmt.Errorf("DEBUG job timeout exceeded, %v", j)
			}
			once.Do(func() {
				running = true
				// attempt wake on timeout
				ticker.Reset(ms)
				// calc next move according to scripted logic
				req := j.script.Next(j.state())
				// DEBUG unreachable, for long-running process!
				outch <- Ticket{owner: j, move: req}
			})
		}
		return nil
	})
}

const (
	// theoritical group size of 5
	groupTurnMs = 5000 * time.Millisecond
	timeSliceMs = 1000 * time.Millisecond
)

type Mask uint16

const (
	None Mask = 1 << iota
	Self
	Barrier
	Robot
	FirstAid
	Corpse
	Clone
	ToggleSwitch
	NorthPole
	SouthPole
	Demagnetized
)

func (m Mask) Set(flag Mask) Mask { return m | flag }
func (m Mask) Has(flag Mask) bool { return m&flag != 0 }
func (m Mask) Not(flag Mask) bool { return m&flag == 0 }

package polarity

import (
	"context"
	"fmt"
	//"encoding/json"
	"sync"
	"time"
)
import "golang.org/x/sync/errgroup"

// maze acts as world state
type Maze struct {
	jobs []Job
	grid [][]Mask
}

// *wasm* a user defined script
type Script interface {
	Next(state string) string
}
type Job struct {
	script Script
	state  string
}
type Ticket struct {
	move  string
	owner Job
}

func NewMaze(width int) *Maze {
	m := &Maze{
		grid: [][]Mask{},
	}
	return m
}

// to be called by the game lifecycle
func (m *Maze) Update() error {
	// TODO shuffle order of workers
	if m.Done() {
		return fmt.Errorf("Simulation done.")
	}

	acc := make(chan Ticket)
	defer close(acc)
	cx, cancel := context.WithTimeout(context.Background(), 8000*time.Millisecond)
	defer cancel()
	grp, ctx := errgroup.WithContext(cx)
	for _, j := range m.jobs {
		spawn(j, acc, grp, ctx, 2000*time.Millisecond)
	}

	if err := grp.Wait(); err != nil {
		// TODO on timeout, there may be tickets in the acc channel
		return err
	}
	//TODO goroutine? to sync world state and enqueue tickets
	// all jobs completed, next moves are waiting in the acc channel
	// so hypothetically need to range acc updating each job state,
	// requeuing for jobs to be processed next-cycle
	m.jobs = m.jobs[:0]

	return nil
}

func (m *Maze) Done() bool {
	// TODO more halt conditions
	if len(m.jobs) <= 1 {
		// work queue drained
		return true
	}
	return false
}

// spawn the script runner
func spawn(j Job, out chan<- Ticket, g *errgroup.Group, ctx context.Context, ms time.Duration) {
	var (
		once    sync.Once
		running = false
		ticker  = time.NewTicker(5 * time.Millisecond)
	)
	g.Go(func() error {
		defer ticker.Stop()
		select {
		case <-ctx.Done():
			// group cancelled
			return ctx.Err()
		case <-ticker.C:
			if running {
				return fmt.Errorf("DEBUG job timeout exceeded, %v", j)
			}
			once.Do(func() {
				running = true
				// attempt wake on timeout
				ticker = time.NewTicker(ms)
				// calc next move according to scripted logic
				req := j.script.Next(j.state)
				// DEBUG unreachable, for long-running process!
				out <- Ticket{owner: j, move: req}
			})
		}
		return nil
	})
}

type Mask uint16

const (
	None Mask = 1 << iota
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

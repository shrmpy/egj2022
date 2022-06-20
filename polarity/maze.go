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

func NewMaze(wd int) *Maze {
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
	next := make([]Job, len(m.jobs))
	go func() {
		var i = 0
		for t := range acc {
			nxst := m.eval(t)
			next[i] = Job{script: t.owner.script, state: nxst}
			i += 1
		}
	}()

	cx, cancel := context.WithTimeout(context.Background(), groupTurnMs)
	defer cancel()
	grp, ctx := errgroup.WithContext(cx)
	for _, j := range m.jobs {
		spawn(j, acc, grp, ctx, timeSliceMs)
	}
	if err := grp.Wait(); err != nil {
		close(acc)
		return err
	}
	// end the for-range
	close(acc)
	// enqueue jobs to be processed next-cycle
	m.jobs = m.jobs[:0]
	copy(m.jobs, next)

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

func (m *Maze) eval(t Ticket) string {
	// TODO resolve move and transition world state
	return t.owner.state
}

// spawn the script runner
func spawn(j Job, out chan<- Ticket, g *errgroup.Group, ctx context.Context, ms time.Duration) {
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
				req := j.script.Next(j.state)
				// DEBUG unreachable, for long-running process!
				out <- Ticket{owner: j, move: req}
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

package polarity

import (
	"context"
	"fmt"
	"encoding/json"
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
type Job struct {
	script Script
	state  string
	name string
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
	next := make(map[string]Job, len(m.jobs))
	go func() {
		for t := range acc {
			nxst := m.eval(t)
			next[t.owner.name] = Job{
				script: t.owner.script,
				name: t.owner.name,
				state: nxst,
			}
		}
	}()

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
	// enqueue jobs to be processed next-cycle
	//m.jobs = m.jobs[:0]
	for k := range m.jobs {
		delete(m.jobs, k)
	}
	for k, v :=  range next {
		m.jobs[k] = v
		delete(next, k)
	}

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
	// TODO resolve move and sync world state
	var move Move
	if err := json.Unmarshal([]byte(t.move), &move); err != nil {
		return unresolved(t.owner)
	}
	switch move.Command {
	case "walk": walk(m, t.owner, move.Direction)
	}

	return unresolved(t.owner)
}
func walk(m *Maze, j Job, dir string) {
	// TODO
	// - extract xy position
	// - is direction blocked?
	// - ifnot then modify position in job's state
	//   and set grid mask at xy position
	//   and unset grid mask from old xy position
}
func unresolved(j Job) string {
	// TODO capture error
	// TODO indicate unsuccessful request
	return j.state
}

// spawn a script runner
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

type Move struct {
	Command string
	Direction string
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

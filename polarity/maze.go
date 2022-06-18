package polarity

import (
	"context"
	"fmt"
	"encoding/json"
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
	state string
}
type Ticket struct {
	move string
	owner Job
}
// to be called by the game lifecycle
func (m *Maze) Update() error {
	// TODO
	// - shuffle order of workers

	if m.Done() return fmt.Errorf("Simulation done.")
	acc := make(chan Ticket)
	cx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	grp, ctx := errgroup.WithContext(cx)
	for _, j := range m.jobs {
		spawn(j, acc, grp, ctx)
	}

	defer close(acc)
	if err := grp.Wait(); err != nil {
		return err
	}
	//TODO goroutine? to sync world state and enqueue tickets

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
func spawn(j Job, out chan<- Ticket, g *errgroup.Group) {
	g.Go(func() error {
		select {
		case <-ctx.Done(): // timeout exceeded
			return ctx.Err()

		//case out<- Ticket{owner: j, move: j.script.Next(j.state)}:
		case req := j.script.Next(j.state):
			out<- Ticket{owner: j, move: req}
		}
		return nil
	})
}

type Mask uint16
const (
	None Mask = 1 << iota
	Barrier
	Agent
	FirstAid
	Corpse
	Friend
	ToggleSwitch
	NorthPole
	SouthPole
	Demagnetized
)
func (m Mask) Set(flag Mask) Mask { return m | flag }
func (m Mask) Has(flag Mask) bool { return m&flag != 0 }
func (m Mask) Not(flag Mask) bool { return m&flag == 0 }


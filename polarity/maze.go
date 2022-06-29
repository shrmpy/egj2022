package polarity

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
)
import "golang.org/x/sync/errgroup"

// maze acts as world state
type Maze struct {
	jobs  []Job
	mini  Minimap
	width int
	turn  int
	hist  func(string, ...any)
	loop  func(i Mask, or, oc, r, c int)
}

// *wasm* a user defined script
type Script interface {
	Next(state string) string
}
type Ticket struct {
	move  string
	owner Job
}

func NewMaze(wd int, fn func(t string, v ...any), al func(i Mask, or, oc, r, c int)) *Maze {
	// TODO populate the maze (and load jaeger scripts)
	rand.Seed(time.Now().UnixNano())
	var row, col int
	row, col = rollRowCol(wd)
	mechA := newJaeger(row, col, wd, "Gypsy")
	row, col = rollRowCol(wd)
	mechB := newJaeger(row, col, wd, "Cherno")
	mm := NewMinimap(wd)
	mm.Jaeger(mechA)
	mm.Jaeger(mechB)
	wq := []Job{
		mechA,
		mechB,
	}

	return &Maze{
		width: wd,
		mini:  mm,
		jobs:  wq,
		hist:  fn,
		loop:  al,
	}
}

// to be called by the game lifecycle
func (m *Maze) Update() error {
	m.turn++
	sz := len(m.jobs)
	// TODO shuffle order of workers
	if err := m.Done(); err != nil {
		return err
	}
	acc := make(chan Ticket, sz)
	defer close(acc)
	cx, cancel := context.WithTimeout(context.Background(), groupTurnMs)
	defer cancel()
	grp, ctx := errgroup.WithContext(cx)
	for _, j := range m.jobs {
		spawn(j, acc, grp, ctx)
	}
	if err := grp.Wait(); err != nil {
		// TODO not all errors are equal
		m.hist("DEBUG wait, %s", err.Error())
		return err
	}
	next := m.listen(sz, acc)
	copy(m.jobs, next)
	return nil
}

func (m *Maze) Done() error {
	// TODO more halt conditions
	if len(m.jobs) > 1 {
		return nil
	}
	return fmt.Errorf("INFO survivor, %s", m.String())
}

// used by game draw step
func (m *Maze) Mini() Minimap {
	// read-only "cache"
	return m.mini.Copy()
}

// used by game draw step
func (m *Maze) Width() float32 {
	return float32(m.width)
}

// debug print
func (m *Maze) String() string {
	var bld strings.Builder
	for _, v := range m.jobs {
		bld.WriteString(v.String())
	}
	return bld.String()
}

// rcv (action) tickets and prepare as next-queue
func (m *Maze) listen(expect int, inch <-chan Ticket) []Job {
	next := make([]Job, 0, expect)

	for i := 0; i < expect; i++ {
		t := <-inch
		delta := m.eval(t)

		if grounded(delta.inv) {
			m.hist("DEBUG junk, %s", t.owner.name)
			continue
		}

		next = append(next, newJob(t, delta))
	}
	return next
}

// exec script and returns ticket for next (jaeger) action
func spawn(j Job, outch chan<- Ticket, g *errgroup.Group, ctx context.Context) {
	var (
		once    sync.Once
		running = false
		ticker  = time.NewTicker(timeSliceMs)
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
			// attempt wake on timeout
			ticker.Reset(timeSliceMs)
		default:
			once.Do(func() {
				running = true
				// calc next move according to scripted logic
				req := j.script.Next(j.state())
				// DEBUG unreachable, for long-running process!
				outch <- Ticket{owner: j, move: req}
			})
		}
		return nil
	})
}

func rollRowCol(wd int) (int, int) {
	// TODO version that doesn't overlap
	row := rand.Intn(wd)
	col := rand.Intn(wd)
	return row, col
}

const (
	// theoritical group size of 5
	groupTurnMs = 5000 * time.Millisecond
	timeSliceMs = 1000 * time.Millisecond
)

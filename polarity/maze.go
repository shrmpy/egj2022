package polarity

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)
import "golang.org/x/sync/errgroup"

// maze acts as world state
type Maze struct {
	jobs  map[string]Job
	mini  Minimap
	width int
	turn  int
	log   func(string)
}

// *wasm* a user defined script
type Script interface {
	Next(state string) string
}
type Ticket struct {
	move  string
	owner Job
}

func NewMaze(wd int, f func(string)) *Maze {
	// TODO populate the maze (and load robot scripts)
	rand.Seed(time.Now().UnixNano())
	var row, col int
	row, col = rollRowCol(wd)
	botA := newRobot(row, col, wd, "Avocado")
	row, col = rollRowCol(wd)
	botB := newRobot(row, col, wd, "Bacon")
	mm := NewMinimap(wd)
	mm.Robot(botA)
	mm.Robot(botB)
	wq := map[string]Job{
		botA.name: botA,
		botB.name: botB,
	}

	return &Maze{
		width: wd,
		mini:  mm,
		jobs:  wq,
		log:   f,
	}
}

// to be called by the game lifecycle
func (m *Maze) Update() error {
	m.turn++
	// TODO shuffle order of workers
	if err := m.Done(); err != nil {
		m.log(fmt.Sprintf("DEBUG turn %d, %s", m.turn, err.Error()))
		return err
	}
	acc := make(chan Ticket)
	next := cleanJobs()
	go m.listen(next, acc)

	cx, cancel := context.WithTimeout(context.Background(), groupTurnMs)
	defer cancel()
	grp, ctx := errgroup.WithContext(cx)
	for _, j := range m.jobs {
		spawn(j, acc, grp, ctx)
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

func (m *Maze) Done() error {
	// TODO more halt conditions
	if len(m.jobs) > 1 {
		return nil
	}
	// 1 survivor
	return fmt.Errorf("DEBUG last survivor, %d", len(m.jobs))
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

// rcv (action) tickets and prepare as next-queue job
func (m *Maze) listen(next map[string]Job, inch <-chan Ticket) {
	// TODO _Lock_ maze map/slice, atm only this thread commits writes
	// only this thread tries to access 'next' map
	for t := range inch {
		delta := m.eval(t)
		if dead(delta.inv) {
			m.log(fmt.Sprintf("DEBUG %s deceased, (%d, %d) %v", t.owner.name, t.owner.row, t.owner.col, t.move))
			continue
		}
		m.log(fmt.Sprintf("DEBUG next added, %s", t.owner.name))
		next[t.owner.name] = newJob(t, delta)
	}
}

// exec script and returns ticket for next (robot) action
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

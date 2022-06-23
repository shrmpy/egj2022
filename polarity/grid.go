package polarity

// grid the row x col set of cells
type Grid [][]Mask

func NewGrid(wd int) Grid {
	rows := make([][]Mask, wd)
	for r := range rows {
		rows[r] = make([]Mask, wd)
	}

	return rows
}

func (g Grid) Corpse(j Job, d Delta) {
	g.Change(Robot, Corpse, j.row, j.col, d.row, d.col)
}

// walk from robot (/self) pov
func (g Grid) WalkMe(j Job, d Delta) {
	g.Change(Self, Self, j.row, j.col, d.row, d.col)
}

// walk changes position in maze
func (g Grid) Walk(j Job, d Delta) {
	g.Change(Robot, Robot, j.row, j.col, d.row, d.col)
}

func (g Grid) Change(oldm, newm Mask, oldrow, oldcol, row, col int) {
	// apply new mask
	newc := g[row][col].Set(newm)
	g[row][col] = newc
	// clear old mask
	oldc := g[oldrow][oldcol].Del(oldm)
	g[oldrow][oldcol] = oldc
}

// initial placement
func (g Grid) Robot(j Job) {
	newc := g[j.row][j.col].Set(Robot)
	g[j.row][j.col] = newc
}

// initial placement from robot pov
func (g Grid) RobotMe(j Job) {
	newc := g[j.row][j.col].Set(Self)
	g[j.row][j.col] = newc
}

// TODO slice copy and nil to avoid losing orphans from pass by reference

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
func (m Mask) Del(flag Mask) Mask { return m &^ flag }
func (m Mask) Has(flag Mask) bool { return m&flag != 0 }
func (m Mask) Not(flag Mask) bool { return m&flag == 0 }

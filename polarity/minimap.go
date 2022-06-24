package polarity

// minimap of row x col maze
// (2 scopes - maze and robot because bots have fog-of-war)
type Minimap [][]Mask

func NewMinimap(wd int) Minimap {
	rows := make([][]Mask, wd)
	for r := range rows {
		rows[r] = make([]Mask, wd)
	}

	return rows
}

func (m Minimap) Corpse(j Job, d Delta) {
	m.Change(Robot, Corpse, j.row, j.col, d.row, d.col)
}

// walk from robot (/self) pov
func (m Minimap) WalkMe(j Job, d Delta) {
	m.Change(Self, Self, j.row, j.col, d.row, d.col)
}

// walk changes position in maze
func (m Minimap) Walk(j Job, d Delta) {
	m.Change(Robot, Robot, j.row, j.col, d.row, d.col)
}

func (m Minimap) Change(oldm, newm Mask, oldrow, oldcol, row, col int) {
	// apply new mask
	newc := m[row][col].Set(newm)
	m[row][col] = newc
	// clear old mask
	oldc := m[oldrow][oldcol].Del(oldm)
	m[oldrow][oldcol] = oldc
}

// initial placement
func (m Minimap) Robot(j Job) {
	newc := m[j.row][j.col].Set(Robot)
	m[j.row][j.col] = newc
}

// initial placement from robot pov
func (m Minimap) RobotMe(j Job) {
	newc := m[j.row][j.col].Set(Self)
	m[j.row][j.col] = newc
}

// TODO slice copy and nil to avoid losing orphans from pass by reference
func (m Minimap) Copy() Minimap {
	// assume square (wd = ht)
	wd := len(m)
	mm := make([][]Mask, wd)
	for y, row := range m {
		mm[y] = make([]Mask, wd)
		copy(mm[y], row)
	}
	return mm
}

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

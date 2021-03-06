package main

import (
	"fmt"
	"math"
	"math/rand"

	"sort"

	"io"

	svg "github.com/ajstarks/svgo"
	"github.com/gin-gonic/gin"
)

const CANVAS_PADDING = 32
const CANVAS_WIDTH = 1024
const CANVAS_HEIGHT = 1024
const CELLS_SIDE = 10
const CELL_PADDING = 5
const MAX_USERS = 15

var WIDTH_TO_SIDE = 1 / math.Sqrt(3)

type TownType int

const (
	TownCode TownType = iota
	TownDocs
	TownTests
)

type ByZIndex []Tower

func (a ByZIndex) Len() int      { return len(a) }
func (a ByZIndex) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByZIndex) Less(i, j int) bool {
	return a[i].ZIndex < a[j].ZIndex
}

func random(from, to int) int {
	if from >= to {
		return from
	}
	return rand.Intn(to-from) + from
}

func Round(f float64) int {
	return int(math.Floor(f + .5))
}

func randColorGrey() (int, int, int) {
	r := random(190, 210)
	g := r + random(-10, 10)
	b := r + random(-10, 10)
	return r, g, b
}

func randColorPastelle() (int, int, int) {
	return rand.Intn(128) + 127, rand.Intn(128) + 127, rand.Intn(128) + 127
}

func lightColor(r, g, b int, k float64) (int, int, int) {
	k = math.Max(float64(k), 0)
	k = math.Min(float64(k), 1)

	r = int(math.Min(float64(r)+float64(r)*k, 255))
	g = int(math.Min(float64(g)+float64(g)*k, 255))
	b = int(math.Min(float64(b)+float64(b)*k, 255))
	fmt.Printf("r:%v g:%v b:%v\n", r, g, b)
	return r, g, b
}

func darkColor(r, g, b int, k float64) (int, int, int) {
	k = math.Max(float64(k), 0)
	k = math.Min(float64(k), 1)
	k = 1 - k
	return Round(float64(r) * k), Round(float64(g) * k), Round(float64(b) * k)
}

func fill(r, g, b int) string {
	return fmt.Sprintf("fill:rgb(%d,%d,%d)", r, g, b)
}

/*func drawCanvas(c *gin.Context) {
	c.Writer.Header().Set("Content-type", "image/svg+xml")
	canvas := svg.New(c.Writer)
	canvas.Start(CANVAS_WIDTH, CANVAS_HEIGHT)
	canvas.Title("Cubes")
	drawFloor(canvas)
	x, _ := strconv.Atoi(c.Query("x"))
	y, _ := strconv.Atoi(c.Query("y"))
	w, _ := strconv.Atoi(c.Query("w"))
	h, _ := strconv.Atoi(c.Query("h"))
	z, _ := strconv.Atoi(c.Query("z"))

	drawCube(canvas, x, y, w, h, z)

	drawCube(canvas, x+w, y, w, h*2, z)
	drawCube(canvas, x, y+z, w, h*3, z)
	drawCube(canvas, x, y+z*2, w, h*2, z)
	drawCube(canvas, x+w*3, y, w, h*5, z*5)
	drawCube(canvas, x+w*3, y+y*2, w, h*4, z)

	canvas.End()
}*/
func drawRepoHandler(c *gin.Context) {
	c.Writer.Header().Set("Content-type", "image/svg+xml")

	rc := RepoConfig{URL: c.Query("repo")}

	rc.Repo().Draw(c.Writer)
}

func (repo *Repo) Draw(w io.Writer) {
	canvas := svg.New(w)
	canvas.Start(CANVAS_WIDTH, CANVAS_HEIGHT)
	canvas.Title(repo.Owner + "/" + repo.Name)

	drawFloor(canvas)

	rs := repo.getCachedStat()

	sort.Sort(ByCodeLines(rs.Users))

	field := getField(canvas)

	field.addUsersTowers(TownCode, rs)

	field.addUsersTowers(TownDocs, rs)

	field.addUsersTowers(TownTests, rs)

	for i, u := range rs.Users {
		//field.addTower(TownCode, u)

		fmt.Printf("%v(%v): %d lines/%d (%v%%)\n", u.Username, u.Email, u.CodeLines.Total, rs.CodeLines.Total, int(100*float64(u.CodeLines.Total)/float64(rs.CodeLines.Total)))
		if (i + 1) == MAX_USERS {
			break
		}
	}
	sort.Sort(ByZIndex(field.Towers))

	for _, tower := range field.Towers {
		x, y := tower.Cell.Pos()

		fmt.Printf("Tower zindex=%v id=%v w=%v z=%v h=%v x=%v y=%v color %+v\n", tower.ZIndex, tower.Cell.ID, tower.W, tower.Z, tower.H, x, y, tower.Color)
		drawCube(field.Canvas, x+CELL_PADDING*int((tower.Cell.ID%CELLS_SIDE)), y+CELL_PADDING*int((tower.Cell.ID/CELLS_SIDE)), tower.W*CellSize, tower.H*CellSize, tower.Z*CellSize, tower.Color.R, tower.Color.G, tower.Color.B, tower.Email)
	}

	canvas.End()

}

var m3 = math.Sqrt(3)

func p(i int) int {
	return int(m3*2*float64(i) + 0.5)
}

type Tower struct {
	ZIndex int
	W      int
	Z      int
	H      int
	Email  string
	Color  Color
	Cell   *Cell
}
type Field struct {
	Canvas    *svg.SVG
	FreeCells map[TownType][]int
	Building  map[int]int // number of floors
	CellType  map[int]TownType

	Towers []Tower
}

type Cell struct {
	Field  *Field
	ID     int
	ZIndex int
	MaxW   int
	MaxZ   int
	MaxH   int
}

var FieldSideSize = CANVAS_WIDTH * WIDTH_TO_SIDE
var CellSize = int(FieldSideSize/CELLS_SIDE) - CELL_PADDING

func getField(canvas *svg.SVG) *Field {
	f := Field{Canvas: canvas}
	f.CellType = make(map[int]TownType)
	f.Building = make(map[int]int)
	f.FreeCells = make(map[TownType][]int)
	for i := 1; i <= 40; i++ {
		f.FreeCells[TownCode] = append(f.FreeCells[TownCode], i)
		f.CellType[i] = TownCode
	}
	for i := 51; i <= 70; i++ {
		f.FreeCells[TownDocs] = append(f.FreeCells[TownDocs], i)
		f.CellType[i] = TownDocs
	}

	for i := 81; i <= 100; i++ {
		f.FreeCells[TownTests] = append(f.FreeCells[TownTests], i)
		f.CellType[i] = TownTests
	}
	return &f

}
func (f *Field) isCellUsed(cellID int) bool {
	_, cellUsed := f.Building[cellID]
	return cellUsed
}

func (f *Field) isCellOfType(cellID int, townType TownType) bool {
	return f.CellType[cellID] == townType
}

func (f *Field) getFreeCell(townType TownType) *Cell {
	if len(f.FreeCells[townType]) == 0 {
		return nil
	}
	cell := Cell{ID: f.FreeCells[townType][random(0, len(f.FreeCells[townType])-1)], Field: f}

	for (cell.ID+cell.MaxW) <= (cell.ID/CELLS_SIDE+CELLS_SIDE) && !f.isCellUsed(cell.ID+cell.MaxW) && f.isCellOfType(cell.ID+cell.MaxW, townType) {
		cell.MaxW++
	}

	for (cell.ID+cell.MaxZ*CELLS_SIDE) <= (CELLS_SIDE*CELLS_SIDE) && !f.isCellUsed(cell.ID+cell.MaxZ*CELLS_SIDE) && f.isCellOfType(cell.ID+cell.MaxZ*CELLS_SIDE, townType) {
		cell.MaxZ++
	}
	return &cell

}

func (c *Cell) Pos() (x, y int) {
	return (c.ID % CELLS_SIDE) * CellSize, (c.ID / CELLS_SIDE) * CellSize
}

func (c *Cell) Use(w, z, h int) {
	//fmt.Printf("Use id=%v w=%v z=%v\n", c.ID, w, z)
	//spew.Dump(c.Field.FreeCells)
	for i := c.ID; i < c.ID+w; i++ {
		//fmt.Printf("i: %d\n", i)
		for j := i; j < i+z*CELLS_SIDE; j = j + CELLS_SIDE {
			//fmt.Printf("j: %d\n", i)
			if cellType, exists := c.Field.CellType[j]; exists {
				zindex := 100*(j/CELLS_SIDE+j%CELLS_SIDE) - int(math.Abs(float64(j/CELLS_SIDE-j%CELLS_SIDE)))
				if zindex > c.ZIndex {
					c.ZIndex = zindex
				}

				//	fmt.Printf("ok\n")
				for pi, cellID := range c.Field.FreeCells[cellType] {
					//	fmt.Printf("cellID: %v (%v)\n", cellID, j)
					if cellID == j {
						//	fmt.Printf("removed \n")
						c.Field.FreeCells[cellType] = append(c.Field.FreeCells[cellType][:pi], c.Field.FreeCells[cellType][pi+1:]...)
						break
					}
				}
				c.Field.Building[c.ID] = h
			}
		}

	}
	//spew.Dump(c.Field.FreeCells)

}

func getUserTotal(us *UserStat, townType TownType) int {
	if townType == TownCode {
		return us.CodeLines.Total
	} else if townType == TownDocs {
		return us.DocLines.Total
	} else if townType == TownTests {
		return us.TestLines.Total
	}
	return 0
}

func getRepoTotal(us *RepoStat, townType TownType) int {
	if townType == TownCode {
		return us.CodeLines.Total
	} else if townType == TownDocs {
		return us.DocLines.Total
	} else if townType == TownTests {
		return us.TestLines.Total
	}
	return 0
}

func (f *Field) addUsersTowers(townType TownType, rs *RepoStat) {

	totalUsers := 10
	if len(rs.Users) < 10 {
		totalUsers = len(rs.Users)
	}
	var volume, totalVolume float64
	var r, g, b int

	userLines := 0
	userVolume := 0.0
	downscale := float64(getRepoTotal(rs, townType)) / 500

	for i, us := range rs.Users[0 : totalUsers-1] {
		volume = 0
		if us.Color.R > 0 {
			r, g, b = us.Color.R, us.Color.B, us.Color.G
		} else {
			r, g, b = randColorPastelle()
		}
		fmt.Printf("Adding tower for user %v r=%v g=%v b=%v \n", us.Email, r, g, b)
		if i == 0 {

			totalVolume = float64(getUserTotal(us, townType)) / (500.0 * downscale)
			if totalVolume > 30 {
				totalVolume = 30
			}

		} else {
			totalVolume = totalVolume * float64(getUserTotal(us, townType)) / float64(getUserTotal(rs.Users[i-1], townType))
		}
		userLines += getUserTotal(us, townType)

		for volume < totalVolume {
			cell := f.getFreeCell(townType)

			if cell == nil {
				break
			}
			x, y := cell.Pos()

			var h int
			if y == 0 || x == 0 {
				h = random(2, 10)
			} else {
				h = random(1, 6)
			}
			if y == 0 || x == 0 || x > int(FieldSideSize)-CellSize || y > int(FieldSideSize)-CellSize {
				if cell.MaxW > 3 {
					cell.MaxW = 3
				}
				if cell.MaxZ > 3 {
					cell.MaxZ = 3
				}

			} else {
				if cell.MaxW > 2 {
					cell.MaxW = 2
				}
				if cell.MaxZ > 2 {
					cell.MaxZ = 2
				}
			}
			w := random(1, cell.MaxW)
			z := random(1, cell.MaxZ)

			if float64(w*h*z) > totalVolume*1.3 {
				if z > 2 {
					z--
				} else if w > 2 {
					w--
				} else if h > 2 {
					h--
				}
			}
			volume += float64(w * h * z)

			cell.Use(w, z, h)
			//us.Color = Color{R: r, G: g, B: b}
			fmt.Printf("Adding tower for user %v r=%v g=%v b=%v \n", us.Email, r, g, b)
			f.Towers = append(f.Towers, Tower{Email: us.Email, ZIndex: cell.ZIndex, W: w, Z: z, H: h, Cell: cell, Color: Color{R: r, G: g, B: b}})
		}
		userVolume += volume
	}

	totalVolume = userVolume * (float64(getRepoTotal(rs, townType)) / float64(userLines))
	volume = 0
	for volume < totalVolume {
		r, g, b = randColorGrey()
		cell := f.getFreeCell(townType)

		if cell == nil {
			break
		}
		x, y := cell.Pos()

		var h int
		if y == 0 || x == 0 {
			h = random(4, 9)
		} else {
			h = random(1, 6)
		}
		if y == 0 || x == 0 || x > int(FieldSideSize)-CellSize || y > int(FieldSideSize)-CellSize {
			if cell.MaxW > 3 {
				cell.MaxW = 3
			}
			if cell.MaxZ > 3 {
				cell.MaxZ = 3
			}

		} else {
			if cell.MaxW > 2 {
				cell.MaxW = 2
			}
			if cell.MaxZ > 2 {
				cell.MaxZ = 2
			}
		}
		w := random(1, cell.MaxW)
		z := random(1, cell.MaxZ)

		if float64(w*h*z) > totalVolume*1.3 {
			if z > 2 {
				z--
			} else if w > 2 {
				w--
			} else if h > 2 {
				h--
			}
		}
		volume += float64(w * h * z)

		cell.Use(w, z, h)
		//us.Color = Color{R: r, G: g, B: b}
		fmt.Printf("Adding tower for anon \n")
		f.Towers = append(f.Towers, Tower{ZIndex: cell.ZIndex, W: w, Z: z, H: h, Cell: cell, Color: Color{R: r, G: g, B: b}})
	}

}

func drawCube(canvas *svg.SVG, xt, yt, w, h, z, r, g, b int, id string) {
	w = int(w / 4)
	z = int(z / 4)
	y := CANVAS_HEIGHT - int(CANVAS_WIDTH*(1/m3)) - h
	x := CANVAS_WIDTH / 2

	x -= int((m3 / 2) * float64(xt))
	y += int(xt / 2)

	x += int((m3 / 2) * float64(yt))
	y += int(yt / 2)
	canvas.Gid(id)
	tx := []int{x, x + p(z), x - p(w) + p(z), x - p(w), x}
	ty := []int{y, y + z*2, y + (z+w)*2, y + w*2, y}
	canvas.Polygon(tx, ty, fill(lightColor(r, g, b, 0.1)))

	lx := []int{x - p(w), x - p(w) + p(z), x - p(w) + p(z), x - p(w), x - p(w)}
	ly := []int{y + w*2, y + (z+w)*2, y + (z+w)*2 + h, y + w*2 + h, y + w*2}
	canvas.Polygon(lx, ly, fill(darkColor(r, g, b, 0.2)))

	rx := []int{x + p(z), x + p(z), x - p(w) + p(z), x - p(w) + p(z), x + p(z)}
	ry := []int{y + z*2, y + z*2 + h, y + (z+w)*2 + h, y + (z+w)*2, y + z*2}
	//fmt.Printf("x: %+v\ny: %+v\n", rx, ry)
	canvas.Polygon(rx, ry, fill(r, g, b))
	canvas.Gend()
}

func drawFloor(canvas *svg.SVG) {
	r, g, b := random(190, 196), random(190, 240), random(60, 72)
	t := 1 / m3
	x := int(CANVAS_WIDTH / 2)
	y := CANVAS_HEIGHT - int(CANVAS_WIDTH*t)

	tx := []int{x, CANVAS_WIDTH, x, 0, x}
	ty := []int{y, y + int(CANVAS_HEIGHT*t/2), CANVAS_HEIGHT, y + int(CANVAS_HEIGHT*t/2), y}
	canvas.Polygon(tx, ty, fill(r, g, b))

}

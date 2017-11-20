package svg

import (
	strconvStdlib "strconv"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/parse"
	"github.com/tdewolff/parse/strconv"
)

type PathData struct {
	o *Minifier

	x, y        float64
	coords      [][]byte
	coordFloats []float64

	state       PathDataState
	curBuffer   []byte
	altBuffer   []byte
	coordBuffer []byte
}

type PathDataState struct {
	cmd            byte
	prevDigit      bool
	prevDigitIsInt bool
}

func NewPathData(o *Minifier) *PathData {
	return &PathData{
		o: o,
	}
}

func (p *PathData) ShortenPathData(b []byte) []byte {
	var x0, y0 float64
	var cmd byte

	p.x, p.y = 0.0, 0.0
	p.coords = p.coords[:0]
	p.coordFloats = p.coordFloats[:0]
	p.state = PathDataState{}

	j := 0
	for i := 0; i < len(b); i++ {
		c := b[i]
		if c == ' ' || c == ',' || c == '\n' || c == '\r' || c == '\t' {
			continue
		} else if c >= 'A' && (cmd == 0 || cmd != c || c == 'M' || c == 'm') { // any command
			if cmd != 0 {
				j += p.copyInstruction(b[j:], cmd)
				if cmd == 'M' || cmd == 'm' {
					x0 = p.x
					y0 = p.y
				} else if cmd == 'Z' || cmd == 'z' {
					p.x = x0
					p.y = y0
				}
			}
			cmd = c
			p.coords = p.coords[:0]
			p.coordFloats = p.coordFloats[:0]
		} else if n := parse.Number(b[i:]); n > 0 {
			f, _ := strconv.ParseFloat(b[i : i+n])
			p.coords = append(p.coords, b[i:i+n])
			p.coordFloats = append(p.coordFloats, f)
			i += n - 1
		}
	}
	if cmd != 0 {
		j += p.copyInstruction(b[j:], cmd)
	}
	return b[:j]
}

func (p *PathData) copyInstruction(b []byte, cmd byte) int {
	n := len(p.coords)
	if n == 0 {
		if cmd == 'Z' || cmd == 'z' {
			b[0] = 'z'
			return 1
		}
		return 0
	}
	isRelCmd := cmd >= 'a'

	// get new cursor coordinates
	di := 0
	if (cmd == 'M' || cmd == 'm' || cmd == 'L' || cmd == 'l' || cmd == 'T' || cmd == 't') && n%2 == 0 {
		di = 2
		// reprint M always, as the first pair is a move but subsequent pairs are L
		if cmd == 'M' || cmd == 'm' {
			p.state.cmd = byte(0)
		}
	} else if cmd == 'H' || cmd == 'h' || cmd == 'V' || cmd == 'v' {
		di = 1
	} else if (cmd == 'S' || cmd == 's' || cmd == 'Q' || cmd == 'q') && n%4 == 0 {
		di = 4
	} else if (cmd == 'C' || cmd == 'c') && n%6 == 0 {
		di = 6
	} else if (cmd == 'A' || cmd == 'a') && n%7 == 0 {
		di = 7
	} else {
		return 0
	}

	j := 0
	origCmd := cmd
	ax, ay := 0.0, 0.0
	for i := 0; i < n; i += di {
		// subsequent coordinate pairs for M are really L
		if i > 0 && (origCmd == 'M' || origCmd == 'm') {
			origCmd = 'L' + (origCmd - 'M')
		}

		cmd = origCmd
		coords := p.coords[i : i+di]
		coordFloats := p.coordFloats[i : i+di]

		if cmd == 'H' || cmd == 'h' {
			ax = coordFloats[di-1]
			if isRelCmd {
				ay = 0
			} else {
				ay = p.y
			}
		} else if cmd == 'V' || cmd == 'v' {
			if isRelCmd {
				ax = 0
			} else {
				ax = p.x
			}
			ay = coordFloats[di-1]
		} else {
			ax = coordFloats[di-2]
			ay = coordFloats[di-1]
		}

		// switch from L to H or V whenever possible
		if cmd == 'L' || cmd == 'l' {
			if isRelCmd {
				if coordFloats[0] == 0 {
					cmd = 'v'
					coords = coords[1:]
					coordFloats = coordFloats[1:]
				} else if coordFloats[1] == 0 {
					cmd = 'h'
					coords = coords[:1]
					coordFloats = coordFloats[:1]
				}
			} else {
				if coordFloats[0] == p.x {
					cmd = 'V'
					coords = coords[1:]
					coordFloats = coordFloats[1:]
				} else if coordFloats[1] == p.y {
					cmd = 'H'
					coords = coords[:1]
					coordFloats = coordFloats[:1]
				}
			}
		}

		// make a current and alternated path with absolute/relative altered
		var curState, altState PathDataState
		curState = p.shortenCurPosInstruction(cmd, coords)
		if isRelCmd {
			altState = p.shortenAltPosInstruction(cmd-'a'+'A', coordFloats, p.x, p.y)
		} else {
			altState = p.shortenAltPosInstruction(cmd-'A'+'a', coordFloats, -p.x, -p.y)
		}

		// choose shortest, relative or absolute path?
		if len(p.altBuffer) < len(p.curBuffer) {
			j += copy(b[j:], p.altBuffer)
			p.state = altState
		} else {
			j += copy(b[j:], p.curBuffer)
			p.state = curState
		}

		if isRelCmd {
			p.x += ax
			p.y += ay
		} else {
			p.x = ax
			p.y = ay
		}
	}
	return j
}

func (p *PathData) shortenCurPosInstruction(cmd byte, coords [][]byte) PathDataState {
	state := p.state
	p.curBuffer = p.curBuffer[:0]
	if cmd != state.cmd && !(state.cmd == 'M' && cmd == 'L' || state.cmd == 'm' && cmd == 'l') {
		p.curBuffer = append(p.curBuffer, cmd)
		state.cmd = cmd
		state.prevDigit = false
		state.prevDigitIsInt = false
	}
	for i, coord := range coords {
		isFlag := false
		if (cmd == 'A' || cmd == 'a') && (i%7 == 3 || i%7 == 4) {
			isFlag = true
		}

		coord = minify.Number(coord, p.o.Decimals)
		state.copyNumber(&p.curBuffer, coord, isFlag)
	}
	return state
}

func (p *PathData) shortenAltPosInstruction(cmd byte, coordFloats []float64, x, y float64) PathDataState {
	state := p.state
	p.altBuffer = p.altBuffer[:0]
	if cmd != state.cmd && !(state.cmd == 'M' && cmd == 'L' || state.cmd == 'm' && cmd == 'l') {
		p.altBuffer = append(p.altBuffer, cmd)
		state.cmd = cmd
		state.prevDigit = false
		state.prevDigitIsInt = false
	}
	for i, f := range coordFloats {
		isFlag := false
		if cmd == 'L' || cmd == 'l' || cmd == 'C' || cmd == 'c' || cmd == 'S' || cmd == 's' || cmd == 'Q' || cmd == 'q' || cmd == 'T' || cmd == 't' || cmd == 'M' || cmd == 'm' {
			if i%2 == 0 {
				f += x
			} else {
				f += y
			}
		} else if cmd == 'H' || cmd == 'h' {
			f += x
		} else if cmd == 'V' || cmd == 'v' {
			f += y
		} else if cmd == 'A' || cmd == 'a' {
			if i%7 == 5 {
				f += x
			} else if i%7 == 6 {
				f += y
			} else if i%7 == 3 || i%7 == 4 {
				isFlag = true
			}
		}

		p.coordBuffer = strconvStdlib.AppendFloat(p.coordBuffer[:0], f, 'g', -1, 64)
		coord := minify.Number(p.coordBuffer, p.o.Decimals)
		state.copyNumber(&p.altBuffer, coord, isFlag)
	}
	return state
}

func (state *PathDataState) copyNumber(buffer *[]byte, coord []byte, isFlag bool) {
	if state.prevDigit && (coord[0] >= '0' && coord[0] <= '9' || coord[0] == '.' && state.prevDigitIsInt) {
		if coord[0] == '0' && !state.prevDigitIsInt {
			if isFlag {
				*buffer = append(*buffer, ' ', '0')
				state.prevDigitIsInt = true
			} else {
				*buffer = append(*buffer, '.', '0') // aggresively add dot so subsequent numbers could drop leading space
				// prevDigit stays true and prevDigitIsInt stays false
			}
			return
		}
		*buffer = append(*buffer, ' ')
	}
	state.prevDigit = true
	state.prevDigitIsInt = true
	if len(coord) > 2 && coord[len(coord)-2] == '0' && coord[len(coord)-1] == '0' {
		coord[len(coord)-2] = 'e'
		coord[len(coord)-1] = '2'
		state.prevDigitIsInt = false
	} else {
		for _, c := range coord {
			if c == '.' || c == 'e' || c == 'E' {
				state.prevDigitIsInt = false
				break
			}
		}
	}
	*buffer = append(*buffer, coord...)
}

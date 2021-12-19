package finalseg

import (
	"fmt"
	"sort"
)

const minFloat = -3.14e100

var (
	prevStatus = make(map[byte][]byte)
	probStart  = make(map[byte]float64)
)

func init() {
	prevStatus['B'] = []byte{'E', 'S'}
	prevStatus['M'] = []byte{'M', 'B'}
	prevStatus['S'] = []byte{'S', 'E'}
	prevStatus['E'] = []byte{'B', 'M'}
	probStart['B'] = -0.26268660809250016
	probStart['E'] = -3.14e+100
	probStart['M'] = -3.14e+100
	probStart['S'] = -1.4652633398537678
}

type probState struct {
	prob  float64
	state byte
}

func (p probState) String() string {
	return fmt.Sprintf("(%f, %x)", p.prob, p.state)
}

type probStates []*probState

func (ps probStates) Len() int {
	return len(ps)
}

func (ps probStates) Less(i, j int) bool {
	if ps[i].prob == ps[j].prob {
		return ps[i].state < ps[j].state
	}
	return ps[i].prob < ps[j].prob
}

func (ps probStates) Swap(i, j int) {
	ps[i], ps[j] = ps[j], ps[i]
}

func viterbi(obs []rune, states []byte) (float64, []byte) {
	path := make(map[byte][]byte)
	V := make([]map[byte]float64, len(obs))
	V[0] = make(map[byte]float64)
	for _, y := range states {
		if val, ok := probEmit[y][obs[0]]; ok {
			V[0][y] = val + probStart[y]
		} else {
			V[0][y] = minFloat + probStart[y]
		}
		path[y] = []byte{y}
	}

	for t := 1; t < len(obs); t++ {
		newPath := make(map[byte][]byte)
		V[t] = make(map[byte]float64)
		for _, y := range states {
			ps0 := make(probStates, 0)
			var emP float64
			if val, ok := probEmit[y][obs[t]]; ok {
				emP = val
			} else {
				emP = minFloat
			}
			for _, y0 := range prevStatus[y] {
				var transP float64
				if tp, ok := probTrans[y0][y]; ok {
					transP = tp
				} else {
					transP = minFloat
				}
				prob0 := V[t-1][y0] + transP + emP
				ps0 = append(ps0, &probState{prob: prob0, state: y0})
			}
			sort.Sort(sort.Reverse(ps0))
			V[t][y] = ps0[0].prob
			pp := make([]byte, len(path[ps0[0].state]))
			copy(pp, path[ps0[0].state])
			newPath[y] = append(pp, y)
		}
		path = newPath
	}
	ps := make(probStates, 0)
	for _, y := range []byte{'E', 'S'} {
		ps = append(ps, &probState{V[len(obs)-1][y], y})
	}
	sort.Sort(sort.Reverse(ps))
	v := ps[0]
	return v.prob, path[v.state]
}

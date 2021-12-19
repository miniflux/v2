package finalseg

var probTrans = make(map[byte]map[byte]float64)

func init() {
	probTrans['B'] = map[byte]float64{'E': -0.510825623765990,
		'M': -0.916290731874155}
	probTrans['E'] = map[byte]float64{'B': -0.5897149736854513,
		'S': -0.8085250474669937}
	probTrans['M'] = map[byte]float64{'E': -0.33344856811948514,
		'M': -1.2603623820268226}
	probTrans['S'] = map[byte]float64{'B': -0.7211965654669841,
		'S': -0.6658631448798212}
}

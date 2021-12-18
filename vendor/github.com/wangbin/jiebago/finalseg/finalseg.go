// Package finalseg is the Golang implementation of Jieba's finalseg module.
package finalseg

import (
	"regexp"
)

var (
	reHan  = regexp.MustCompile(`\p{Han}+`)
	reSkip = regexp.MustCompile(`(\d+\.\d+|[a-zA-Z0-9]+)`)
)

func cutHan(sentence string) chan string {
	result := make(chan string)
	go func() {
		runes := []rune(sentence)
		_, posList := viterbi(runes, []byte{'B', 'M', 'E', 'S'})
		begin, next := 0, 0
		for i, char := range runes {
			pos := posList[i]
			switch pos {
			case 'B':
				begin = i
			case 'E':
				result <- string(runes[begin : i+1])
				next = i + 1
			case 'S':
				result <- string(char)
				next = i + 1
			}
		}
		if next < len(runes) {
			result <- string(runes[next:])
		}
		close(result)
	}()
	return result
}

// Cut cuts sentence into words using Hidden Markov Model with Viterbi
// algorithm. It is used by Jiebago for unknonw words.
func Cut(sentence string) chan string {
	result := make(chan string)
	s := sentence
	var hans string
	var hanLoc []int
	var nonhanLoc []int
	go func() {
		for {
			hanLoc = reHan.FindStringIndex(s)
			if hanLoc == nil {
				if len(s) == 0 {
					break
				}
			} else if hanLoc[0] == 0 {
				hans = s[hanLoc[0]:hanLoc[1]]
				s = s[hanLoc[1]:]
				for han := range cutHan(hans) {
					result <- han
				}
				continue
			}
			nonhanLoc = reSkip.FindStringIndex(s)
			if nonhanLoc == nil {
				if len(s) == 0 {
					break
				}
			} else if nonhanLoc[0] == 0 {
				nonhans := s[nonhanLoc[0]:nonhanLoc[1]]
				s = s[nonhanLoc[1]:]
				if nonhans != "" {
					result <- nonhans
					continue
				}
			}
			var loc []int
			if hanLoc == nil && nonhanLoc == nil {
				if len(s) > 0 {
					result <- s
					break
				}
			} else if hanLoc == nil {
				loc = nonhanLoc
			} else if nonhanLoc == nil {
				loc = hanLoc
			} else if hanLoc[0] < nonhanLoc[0] {
				loc = hanLoc
			} else {
				loc = nonhanLoc
			}
			result <- s[:loc[0]]
			s = s[loc[0]:]
		}
		close(result)
	}()
	return result
}

// Package jiebago is the Golang implemention of [Jieba](https://github.com/fxsjy/jieba), Python Chinese text segmentation module.
package jiebago

import (
	"math"
	"regexp"
	"strings"

	"github.com/Soontao/jiebago/dictionary"
	"github.com/Soontao/jiebago/finalseg"
	"github.com/Soontao/jiebago/util"
)

var (
	reEng         = regexp.MustCompile(`[[:alnum:]]`)
	reHanCutAll   = regexp.MustCompile(`(\p{Han}+)`)
	reSkipCutAll  = regexp.MustCompile(`[^[:alnum:]+#\n]`)
	reHanDefault  = regexp.MustCompile(`([\p{Han}+[:alnum:]+#&\._]+)`)
	reSkipDefault = regexp.MustCompile(`(\r\n|\s)`)
)

// Segmenter is a Chinese words segmentation struct.
type Segmenter struct {
	dict *Dictionary
}

// Frequency returns a word's frequency and existence
func (seg *Segmenter) Frequency(word string) (float64, bool) {
	return seg.dict.Frequency(word)
}

// AddWord adds a new word with frequency to dictionary
func (seg *Segmenter) AddWord(word string, frequency float64) {
	seg.dict.AddToken(dictionary.NewToken(word, frequency, ""))
}

// DeleteWord removes a word from dictionary
func (seg *Segmenter) DeleteWord(word string) {
	seg.dict.AddToken(dictionary.NewToken(word, 0.0, ""))
}

/*
SuggestFrequency returns a suggested frequncy of a word or a long word
cutted into several short words.

This method is useful when a word in the sentence is not cutted out correctly.

If a word should not be further cutted, for example word "石墨烯" should not be
cutted into "石墨" and "烯", SuggestFrequency("石墨烯") will return the maximu
frequency for this word.

If a word should be further cutted, for example word "今天天气" should be
further cutted into two words "今天" and "天气",  SuggestFrequency("今天", "天气")
should return the minimum frequency for word "今天天气".
*/
func (seg *Segmenter) SuggestFrequency(words ...string) float64 {
	frequency := 1.0
	if len(words) > 1 {
		for _, word := range words {
			if freq, ok := seg.dict.Frequency(word); ok {
				frequency *= freq
			}
			frequency /= seg.dict.total
		}
		frequency, _ = math.Modf(frequency * seg.dict.total)
		wordFreq := 0.0
		if freq, ok := seg.dict.Frequency(strings.Join(words, "")); ok {
			wordFreq = freq
		}
		if wordFreq < frequency {
			frequency = wordFreq
		}
	} else {
		word := words[0]
		for segment := range seg.Cut(word, false) {
			if freq, ok := seg.dict.Frequency(segment); ok {
				frequency *= freq
			}
			frequency /= seg.dict.total
		}
		frequency, _ = math.Modf(frequency * seg.dict.total)
		frequency += 1.0
		wordFreq := 1.0
		if freq, ok := seg.dict.Frequency(word); ok {
			wordFreq = freq
		}
		if wordFreq > frequency {
			frequency = wordFreq
		}
	}
	return frequency
}

// LoadDictionary loads dictionary from given file name. Everytime
// LoadDictionary is called, previously loaded dictionary will be cleard.
func (seg *Segmenter) LoadDictionary(resource string) error {
	seg.dict = &Dictionary{freqMap: make(map[string]float64)}
	return seg.dict.loadDictionary(resource)
}

// LoadUserDictionary loads a user specified dictionary, it must be called
// after LoadDictionary, and it will not clear any previous loaded dictionary,
// instead it will override exist entries.
func (seg *Segmenter) LoadUserDictionary(resource string) error {
	return seg.dict.loadDictionary(resource)
}

func (seg *Segmenter) dag(runes []rune) map[int][]int {
	dag := make(map[int][]int)
	n := len(runes)
	var frag []rune
	var i int
	for k := 0; k < n; k++ {
		dag[k] = make([]int, 0)
		i = k
		frag = runes[k : k+1]
		for {
			freq, ok := seg.dict.Frequency(string(frag))
			if !ok {
				break
			}
			if freq > 0.0 {
				dag[k] = append(dag[k], i)
			}
			i++
			if i >= n {
				break
			}
			frag = runes[k : i+1]
		}
		if len(dag[k]) == 0 {
			dag[k] = append(dag[k], k)
		}
	}
	return dag
}

type route struct {
	frequency float64
	index     int
}

func (seg *Segmenter) calc(runes []rune) map[int]route {
	dag := seg.dag(runes)
	n := len(runes)
	rs := make(map[int]route)
	rs[n] = route{frequency: 0.0, index: 0}
	var r route
	for idx := n - 1; idx >= 0; idx-- {
		for _, i := range dag[idx] {
			if freq, ok := seg.dict.Frequency(string(runes[idx : i+1])); ok {
				r = route{frequency: math.Log(freq) - seg.dict.logTotal + rs[i+1].frequency, index: i}
			} else {
				r = route{frequency: math.Log(1.0) - seg.dict.logTotal + rs[i+1].frequency, index: i}
			}
			if v, ok := rs[idx]; !ok {
				rs[idx] = r
			} else {
				if v.frequency < r.frequency || (v.frequency == r.frequency && v.index < r.index) {
					rs[idx] = r
				}
			}
		}
	}
	return rs
}

type cutFunc func(sentence string) <-chan string

func (seg *Segmenter) cutDAG(sentence string) <-chan string {
	result := make(chan string)
	go func() {
		runes := []rune(sentence)
		routes := seg.calc(runes)
		var y int
		length := len(runes)
		var buf []rune
		for x := 0; x < length; {
			y = routes[x].index + 1
			frag := runes[x:y]
			if y-x == 1 {
				buf = append(buf, frag...)
			} else {
				if len(buf) > 0 {
					bufString := string(buf)
					if len(buf) == 1 {
						result <- bufString
					} else {
						if v, ok := seg.dict.Frequency(bufString); !ok || v == 0.0 {
							for x := range finalseg.Cut(bufString) {
								result <- x
							}
						} else {
							for _, elem := range buf {
								result <- string(elem)
							}
						}
					}
					buf = make([]rune, 0)
				}
				result <- string(frag)
			}
			x = y
		}

		if len(buf) > 0 {
			bufString := string(buf)
			if len(buf) == 1 {
				result <- bufString
			} else {
				if v, ok := seg.dict.Frequency(bufString); !ok || v == 0.0 {
					for t := range finalseg.Cut(bufString) {
						result <- t
					}
				} else {
					for _, elem := range buf {
						result <- string(elem)
					}
				}
			}
		}
		close(result)
	}()
	return result
}

func (seg *Segmenter) cutDAGNoHMM(sentence string) <-chan string {
	result := make(chan string)

	go func() {
		runes := []rune(sentence)
		routes := seg.calc(runes)
		var y int
		length := len(runes)
		var buf []rune
		for x := 0; x < length; {
			y = routes[x].index + 1
			frag := runes[x:y]
			if reEng.MatchString(string(frag)) && len(frag) == 1 {
				buf = append(buf, frag...)
				x = y
				continue
			}
			if len(buf) > 0 {
				result <- string(buf)
				buf = make([]rune, 0)
			}
			result <- string(frag)
			x = y
		}
		if len(buf) > 0 {
			result <- string(buf)
			buf = make([]rune, 0)
		}
		close(result)
	}()
	return result
}

// Cut cuts a sentence into words using accurate mode.
// Parameter hmm controls whether to use the Hidden Markov Model.
// Accurate mode attempts to cut the sentence into the most accurate
// segmentations, which is suitable for text analysis.
func (seg *Segmenter) Cut(sentence string, hmm bool) <-chan string {
	result := make(chan string)
	var cut cutFunc
	if hmm {
		cut = seg.cutDAG
	} else {
		cut = seg.cutDAGNoHMM
	}

	go func() {
		for _, block := range util.RegexpSplit(reHanDefault, sentence, -1) {
			if len(block) == 0 {
				continue
			}
			if reHanDefault.MatchString(block) {
				for x := range cut(block) {
					result <- x
				}
				continue
			}
			for _, subBlock := range util.RegexpSplit(reSkipDefault, block, -1) {
				if reSkipDefault.MatchString(subBlock) {
					result <- subBlock
					continue
				}
				for _, r := range subBlock {
					result <- string(r)
				}
			}
		}
		close(result)
	}()
	return result
}

func (seg *Segmenter) cutAll(sentence string) <-chan string {
	result := make(chan string)
	go func() {
		runes := []rune(sentence)
		dag := seg.dag(runes)
		start := -1
		ks := make([]int, len(dag))
		for k := range dag {
			ks[k] = k
		}
		var l []int
		for k := range ks {
			l = dag[k]
			if len(l) == 1 && k > start {
				result <- string(runes[k : l[0]+1])
				start = l[0]
				continue
			}
			for _, j := range l {
				if j > k {
					result <- string(runes[k : j+1])
					start = j
				}
			}
		}
		close(result)
	}()
	return result
}

// CutAll cuts a sentence into words using full mode.
// Full mode gets all the possible words from the sentence.
// Fast but not accurate.
func (seg *Segmenter) CutAll(sentence string) <-chan string {
	result := make(chan string)
	go func() {
		for _, block := range util.RegexpSplit(reHanCutAll, sentence, -1) {
			if len(block) == 0 {
				continue
			}
			if reHanCutAll.MatchString(block) {
				for x := range seg.cutAll(block) {
					result <- x
				}
				continue
			}
			for _, subBlock := range reSkipCutAll.Split(block, -1) {
				result <- subBlock
			}
		}
		close(result)
	}()
	return result
}

// CutForSearch cuts sentence into words using search engine mode.
// Search engine mode, based on the accurate mode, attempts to cut long words
// into several short words, which can raise the recall rate.
// Suitable for search engines.
func (seg *Segmenter) CutForSearch(sentence string, hmm bool) <-chan string {
	result := make(chan string)
	go func() {
		for word := range seg.Cut(sentence, hmm) {
			runes := []rune(word)
			for _, increment := range []int{2, 3} {
				if len(runes) <= increment {
					continue
				}
				var gram string
				for i := 0; i < len(runes)-increment+1; i++ {
					gram = string(runes[i : i+increment])
					if v, ok := seg.dict.Frequency(gram); ok && v > 0.0 {
						result <- gram
					}
				}
			}
			result <- word
		}
		close(result)
	}()
	return result
}

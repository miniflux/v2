// Package getlang provides fast natural language detection for various languages
//
// getlang compares input text to a characteristic profile of each supported language and
// returns the language that best matches the input text
package getlang

import (
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
	"io"
	"io/ioutil"
	"math"
	"sort"
	"unicode"
)

const undeterminedRate int = 41
const undetermined string = "und"
const rescale = 0.5
const scriptCountFactor int = 2
const expOverflow = 7.09e+02

var langs = map[string][]string{
	"de":      de,
	"en":      en,
	"es":      es,
	"fr":      fr,
	"hi":      hi,
	"hu":      hu,
	"it":      it,
	"nl":      nl,
	"pl":      pl,
	"pt":      pt,
	"ru":      ru,
	"sr-Latn": srLatin,
	"sr-Cyrl": srCyr,
	"tl":      tl,
	"uk":      uk,
	"vi":      vi,
}

var scripts = map[string][]*unicode.RangeTable{
	"ar": {unicode.Arabic},
	"bn": {unicode.Bengali},
	"el": {unicode.Greek},
	"gu": {unicode.Gujarati},
	"he": {unicode.Hebrew},
	"hy": {unicode.Armenian},
	"ja": {unicode.Hiragana, unicode.Katakana},
	"kn": {unicode.Kannada},
	"ko": {unicode.Hangul},
	"pa": {unicode.Gurmukhi},
	"ta": {unicode.Tamil},
	"te": {unicode.Telugu},
	"th": {unicode.Thai},
	"zh": {unicode.Han},
}

// Info is the language detection result
type Info struct {
	lang        string
	probability float64
	langTag     language.Tag
}

// Tag returns the language.Tag of the detected language
func (info Info) Tag() language.Tag {
	return info.langTag
}

// LanguageCode returns the ISO 639-1 code for the detected language
func (info Info) LanguageCode() string {
	codeLen := len(info.lang)
	if codeLen < 4 {
		return info.lang
	}
	return info.lang[:2]
}

// Confidence returns a measure of reliability for the language classification
//
// The output value is in the range [0, 1.0] inclusive
func (info Info) Confidence() float64 {
	return info.probability
}

// LanguageName returns the English name of the detected language
func (info Info) LanguageName() string {
	return display.English.Tags().Name(info.langTag)
}

// SelfName returns the name of the language in the language itself
func (info Info) SelfName() string {
	return display.Self.Name(info.langTag)
}

// FromReader detects the language from an io.Reader
//
// This function will read all bytes until an EOF is reached
func FromReader(reader io.Reader) (Info, error) {
	bytes, err := ioutil.ReadAll(reader)
	return FromString(string(bytes)), err
}

// FromString detects the language from the given string
func FromString(text string) Info {
	langMatches := make(map[string]int)
	langMatches[undetermined] = 1

	trigs := sortedTrigs(text)
	for k, v := range langs {
		matchWith(k, trigs, v, langMatches)
	}

	for k, v := range scripts {
		matchScript(k, text, langMatches, v...)
	}

	smx := softMax(langMatches)
	maxk := maxKey(langMatches)
	return Info{maxk, smx[maxk], language.MustParse(maxk)}
}

func softMax(mapping map[string]int) map[string]float64 {
	softMaxMap := make(map[string]float64)
	var denom float64
	overflowed := false
	for _, v := range mapping {
		denom += math.Exp(float64(v) * rescale)
		if v > expOverflow {
			overflowed = true
		}
	}
	for k := range mapping {
		if !overflowed {
			softMaxMap[k] = math.Exp(rescale*float64(mapping[k])) / denom
		} else {
			softMaxMap[k] = 1.0
		}
	}
	return softMaxMap
}

func maxKey(mapping map[string]int) string {
	var max int
	var key string
	for k, v := range mapping {
		if v > max {
			max = v
			key = k
		}
	}
	return key
}

func matchScript(langName, text string, matches map[string]int, ranges ...*unicode.RangeTable) {
	for _, r := range text {
		if unicode.In(r, ranges...) {
			matches[langName] += scriptCountFactor
		}
	}
}

func matchWith(langName string, trigs []trigram, langProfile []string, matches map[string]int) {
	var undeterminedCount int
	prof := make(map[string]int)
	for _, x := range langProfile {
		prof[x] = 1
	}

	for _, trig := range trigs {
		if _, exists := prof[trig.trigram]; exists {
			matches[langName] += trig.count
		} else {
			undeterminedCount++
			if (undeterminedCount % undeterminedRate) == 0 {
				matches[undetermined]++
			}
		}
	}
}

func countedTrigrams(text string) map[string]int {
	trigrams := map[string]int{}
	var txt []rune

	for _, r := range text {
		txt = append(txt, unicode.ToLower(toTrigramChar(r)))
	}
	txt = append(txt, ' ')

	r1, r2 := ' ', txt[0]
	var r3 rune
	for i := 1; i < len(txt); i++ {
		r3 = txt[i]
		if !(r2 == ' ' && (r1 == ' ' || r3 == ' ')) {
			trigram := []rune{r1, r2, r3}
			if trigrams[string(trigram)] == 0 {
				trigrams[string(trigram)] = 1
			} else {
				trigrams[string(trigram)]++
			}
		}
		r1, r2 = r2, r3
	}

	return trigrams
}

type trigram struct {
	trigram string
	count   int
}

func sortedTrigs(s string) []trigram {
	counterMap := countedTrigrams(s)
	trigrams := make([]trigram, len(counterMap))

	var i int
	for tg, count := range counterMap {
		trigrams[i] = trigram{tg, count}
		i++
	}
	sort.SliceStable(trigrams, func(i, j int) bool {
		if trigrams[i].count == trigrams[j].count {
			return trigrams[i].trigram < trigrams[j].trigram
		}
		return trigrams[i].count > trigrams[j].count
	})
	return trigrams
}

func toTrigramChar(ch rune) rune {
	if unicode.IsPunct(ch) || unicode.IsSpace(ch) {
		return ' '
	}
	return ch
}

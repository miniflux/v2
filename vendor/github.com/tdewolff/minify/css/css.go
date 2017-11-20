// Package css minifies CSS3 following the specifications at http://www.w3.org/TR/css-syntax-3/.
package css // import "github.com/tdewolff/minify/css"

import (
	"bytes"
	"encoding/hex"
	"io"
	"strconv"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/parse"
	"github.com/tdewolff/parse/css"
)

var (
	spaceBytes          = []byte(" ")
	colonBytes          = []byte(":")
	semicolonBytes      = []byte(";")
	commaBytes          = []byte(",")
	leftBracketBytes    = []byte("{")
	rightBracketBytes   = []byte("}")
	zeroBytes           = []byte("0")
	msfilterBytes       = []byte("-ms-filter")
	backgroundNoneBytes = []byte("0 0")
)

type cssMinifier struct {
	m *minify.M
	w io.Writer
	p *css.Parser
	o *Minifier
}

////////////////////////////////////////////////////////////////

// DefaultMinifier is the default minifier.
var DefaultMinifier = &Minifier{Decimals: -1}

// Minifier is a CSS minifier.
type Minifier struct {
	Decimals int
}

// Minify minifies CSS data, it reads from r and writes to w.
func Minify(m *minify.M, w io.Writer, r io.Reader, params map[string]string) error {
	return DefaultMinifier.Minify(m, w, r, params)
}

// Minify minifies CSS data, it reads from r and writes to w.
func (o *Minifier) Minify(m *minify.M, w io.Writer, r io.Reader, params map[string]string) error {
	isInline := params != nil && params["inline"] == "1"
	c := &cssMinifier{
		m: m,
		w: w,
		p: css.NewParser(r, isInline),
		o: o,
	}
	defer c.p.Restore()

	if err := c.minifyGrammar(); err != nil && err != io.EOF {
		return err
	}
	return nil
}

func (c *cssMinifier) minifyGrammar() error {
	semicolonQueued := false
	for {
		gt, _, data := c.p.Next()
		if gt == css.ErrorGrammar {
			if perr, ok := c.p.Err().(*parse.Error); ok && perr.Message == "unexpected token in declaration" {
				if semicolonQueued {
					if _, err := c.w.Write(semicolonBytes); err != nil {
						return err
					}
				}

				// write out the offending declaration
				if _, err := c.w.Write(data); err != nil {
					return err
				}
				for _, val := range c.p.Values() {
					if _, err := c.w.Write(val.Data); err != nil {
						return err
					}
				}
				semicolonQueued = true
				continue
			} else {
				return c.p.Err()
			}
		} else if gt == css.EndAtRuleGrammar || gt == css.EndRulesetGrammar {
			if _, err := c.w.Write(rightBracketBytes); err != nil {
				return err
			}
			semicolonQueued = false
			continue
		}

		if semicolonQueued {
			if _, err := c.w.Write(semicolonBytes); err != nil {
				return err
			}
			semicolonQueued = false
		}

		if gt == css.AtRuleGrammar {
			if _, err := c.w.Write(data); err != nil {
				return err
			}
			for _, val := range c.p.Values() {
				if _, err := c.w.Write(val.Data); err != nil {
					return err
				}
			}
			semicolonQueued = true
		} else if gt == css.BeginAtRuleGrammar {
			if _, err := c.w.Write(data); err != nil {
				return err
			}
			for _, val := range c.p.Values() {
				if _, err := c.w.Write(val.Data); err != nil {
					return err
				}
			}
			if _, err := c.w.Write(leftBracketBytes); err != nil {
				return err
			}
		} else if gt == css.QualifiedRuleGrammar {
			if err := c.minifySelectors(data, c.p.Values()); err != nil {
				return err
			}
			if _, err := c.w.Write(commaBytes); err != nil {
				return err
			}
		} else if gt == css.BeginRulesetGrammar {
			if err := c.minifySelectors(data, c.p.Values()); err != nil {
				return err
			}
			if _, err := c.w.Write(leftBracketBytes); err != nil {
				return err
			}
		} else if gt == css.DeclarationGrammar {
			if _, err := c.w.Write(data); err != nil {
				return err
			}
			if _, err := c.w.Write(colonBytes); err != nil {
				return err
			}
			if err := c.minifyDeclaration(data, c.p.Values()); err != nil {
				return err
			}
			semicolonQueued = true
		} else if gt == css.CustomPropertyGrammar {
			if _, err := c.w.Write(data); err != nil {
				return err
			}
			if _, err := c.w.Write(colonBytes); err != nil {
				return err
			}
			if _, err := c.w.Write(c.p.Values()[0].Data); err != nil {
				return err
			}
			semicolonQueued = true
		} else if gt == css.CommentGrammar {
			if len(data) > 5 && data[1] == '*' && data[2] == '!' {
				if _, err := c.w.Write(data[:3]); err != nil {
					return err
				}
				comment := parse.TrimWhitespace(parse.ReplaceMultipleWhitespace(data[3 : len(data)-2]))
				if _, err := c.w.Write(comment); err != nil {
					return err
				}
				if _, err := c.w.Write(data[len(data)-2:]); err != nil {
					return err
				}
			}
		} else if _, err := c.w.Write(data); err != nil {
			return err
		}
	}
}

func (c *cssMinifier) minifySelectors(property []byte, values []css.Token) error {
	inAttr := false
	isClass := false
	for _, val := range c.p.Values() {
		if !inAttr {
			if val.TokenType == css.IdentToken {
				if !isClass {
					parse.ToLower(val.Data)
				}
				isClass = false
			} else if val.TokenType == css.DelimToken && val.Data[0] == '.' {
				isClass = true
			} else if val.TokenType == css.LeftBracketToken {
				inAttr = true
			}
		} else {
			if val.TokenType == css.StringToken && len(val.Data) > 2 {
				s := val.Data[1 : len(val.Data)-1]
				if css.IsIdent([]byte(s)) {
					if _, err := c.w.Write(s); err != nil {
						return err
					}
					continue
				}
			} else if val.TokenType == css.RightBracketToken {
				inAttr = false
			}
		}
		if _, err := c.w.Write(val.Data); err != nil {
			return err
		}
	}
	return nil
}

func (c *cssMinifier) minifyDeclaration(property []byte, values []css.Token) error {
	if len(values) == 0 {
		return nil
	}
	prop := css.ToHash(property)
	inProgid := false
	for i, value := range values {
		if inProgid {
			if value.TokenType == css.FunctionToken {
				inProgid = false
			}
			continue
		} else if value.TokenType == css.IdentToken && css.ToHash(value.Data) == css.Progid {
			inProgid = true
			continue
		}
		value.TokenType, value.Data = c.shortenToken(prop, value.TokenType, value.Data)
		if prop == css.Font || prop == css.Font_Family || prop == css.Font_Weight {
			if value.TokenType == css.IdentToken && (prop == css.Font || prop == css.Font_Weight) {
				val := css.ToHash(value.Data)
				if val == css.Normal && prop == css.Font_Weight {
					// normal could also be specified for font-variant, not just font-weight
					value.TokenType = css.NumberToken
					value.Data = []byte("400")
				} else if val == css.Bold {
					value.TokenType = css.NumberToken
					value.Data = []byte("700")
				}
			} else if value.TokenType == css.StringToken && (prop == css.Font || prop == css.Font_Family) && len(value.Data) > 2 {
				unquote := true
				parse.ToLower(value.Data)
				s := value.Data[1 : len(value.Data)-1]
				if len(s) > 0 {
					for _, split := range bytes.Split(s, spaceBytes) {
						val := css.ToHash(split)
						// if len is zero, it contains two consecutive spaces
						if val == css.Inherit || val == css.Serif || val == css.Sans_Serif || val == css.Monospace || val == css.Fantasy || val == css.Cursive || val == css.Initial || val == css.Default ||
							len(split) == 0 || !css.IsIdent(split) {
							unquote = false
							break
						}
					}
				}
				if unquote {
					value.Data = s
				}
			}
		} else if prop == css.Outline || prop == css.Border || prop == css.Border_Bottom || prop == css.Border_Left || prop == css.Border_Right || prop == css.Border_Top {
			if css.ToHash(value.Data) == css.None {
				value.TokenType = css.NumberToken
				value.Data = zeroBytes
			}
		}
		values[i].TokenType, values[i].Data = value.TokenType, value.Data
	}

	important := false
	if len(values) > 2 && values[len(values)-2].TokenType == css.DelimToken && values[len(values)-2].Data[0] == '!' && css.ToHash(values[len(values)-1].Data) == css.Important {
		values = values[:len(values)-2]
		important = true
	}

	if len(values) == 1 {
		if prop == css.Background && css.ToHash(values[0].Data) == css.None {
			values[0].Data = backgroundNoneBytes
		} else if bytes.Equal(property, msfilterBytes) {
			alpha := []byte("progid:DXImageTransform.Microsoft.Alpha(Opacity=")
			if values[0].TokenType == css.StringToken && bytes.HasPrefix(values[0].Data[1:len(values[0].Data)-1], alpha) {
				values[0].Data = append(append([]byte{values[0].Data[0]}, []byte("alpha(opacity=")...), values[0].Data[1+len(alpha):]...)
			}
		}
	} else {
		if prop == css.Margin || prop == css.Padding || prop == css.Border_Width {
			if (values[0].TokenType == css.NumberToken || values[0].TokenType == css.DimensionToken || values[0].TokenType == css.PercentageToken) && (len(values)+1)%2 == 0 {
				valid := true
				for i := 1; i < len(values); i += 2 {
					if values[i].TokenType != css.WhitespaceToken || values[i+1].TokenType != css.NumberToken && values[i+1].TokenType != css.DimensionToken && values[i+1].TokenType != css.PercentageToken {
						valid = false
						break
					}
				}
				if valid {
					n := (len(values) + 1) / 2
					if n == 2 {
						if bytes.Equal(values[0].Data, values[2].Data) {
							values = values[:1]
						}
					} else if n == 3 {
						if bytes.Equal(values[0].Data, values[2].Data) && bytes.Equal(values[0].Data, values[4].Data) {
							values = values[:1]
						} else if bytes.Equal(values[0].Data, values[4].Data) {
							values = values[:3]
						}
					} else if n == 4 {
						if bytes.Equal(values[0].Data, values[2].Data) && bytes.Equal(values[0].Data, values[4].Data) && bytes.Equal(values[0].Data, values[6].Data) {
							values = values[:1]
						} else if bytes.Equal(values[0].Data, values[4].Data) && bytes.Equal(values[2].Data, values[6].Data) {
							values = values[:3]
						} else if bytes.Equal(values[2].Data, values[6].Data) {
							values = values[:5]
						}
					}
				}
			}
		} else if prop == css.Filter && len(values) == 11 {
			if bytes.Equal(values[0].Data, []byte("progid")) &&
				values[1].TokenType == css.ColonToken &&
				bytes.Equal(values[2].Data, []byte("DXImageTransform")) &&
				values[3].Data[0] == '.' &&
				bytes.Equal(values[4].Data, []byte("Microsoft")) &&
				values[5].Data[0] == '.' &&
				bytes.Equal(values[6].Data, []byte("Alpha(")) &&
				bytes.Equal(parse.ToLower(values[7].Data), []byte("opacity")) &&
				values[8].Data[0] == '=' &&
				values[10].Data[0] == ')' {
				values = values[6:]
				values[0].Data = []byte("alpha(")
			}
		}
	}

	for i := 0; i < len(values); i++ {
		if values[i].TokenType == css.FunctionToken {
			n, err := c.minifyFunction(values[i:])
			if err != nil {
				return err
			}
			i += n - 1
		} else if _, err := c.w.Write(values[i].Data); err != nil {
			return err
		}
	}
	if important {
		if _, err := c.w.Write([]byte("!important")); err != nil {
			return err
		}
	}
	return nil
}

func (c *cssMinifier) minifyFunction(values []css.Token) (int, error) {
	n := 1
	simple := true
	for i, value := range values[1:] {
		if value.TokenType == css.RightParenthesisToken {
			n++
			break
		}
		if i%2 == 0 && (value.TokenType != css.NumberToken && value.TokenType != css.PercentageToken) || (i%2 == 1 && value.TokenType != css.CommaToken) {
			simple = false
		}
		n++
	}
	values = values[:n]
	if simple && (n-1)%2 == 0 {
		fun := css.ToHash(values[0].Data[:len(values[0].Data)-1])
		nArgs := (n - 1) / 2
		if (fun == css.Rgba || fun == css.Hsla) && nArgs == 4 {
			d, _ := strconv.ParseFloat(string(values[7].Data), 32) // can never fail because if simple == true than this is a NumberToken or PercentageToken
			if d-1.0 > -minify.Epsilon {
				if fun == css.Rgba {
					values[0].Data = []byte("rgb(")
					fun = css.Rgb
				} else {
					values[0].Data = []byte("hsl(")
					fun = css.Hsl
				}
				values = values[:len(values)-2]
				values[len(values)-1].Data = []byte(")")
				nArgs = 3
			} else if d < minify.Epsilon {
				values[0].Data = []byte("transparent")
				values = values[:1]
				fun = 0
				nArgs = 0
			}
		}
		if fun == css.Rgb && nArgs == 3 {
			var err [3]error
			rgb := [3]byte{}
			for j := 0; j < 3; j++ {
				val := values[j*2+1]
				if val.TokenType == css.NumberToken {
					var d int64
					d, err[j] = strconv.ParseInt(string(val.Data), 10, 32)
					if d < 0 {
						d = 0
					} else if d > 255 {
						d = 255
					}
					rgb[j] = byte(d)
				} else if val.TokenType == css.PercentageToken {
					var d float64
					d, err[j] = strconv.ParseFloat(string(val.Data[:len(val.Data)-1]), 32)
					if d < 0.0 {
						d = 0.0
					} else if d > 100.0 {
						d = 100.0
					}
					rgb[j] = byte((d / 100.0 * 255.0) + 0.5)
				}
			}
			if err[0] == nil && err[1] == nil && err[2] == nil {
				val := make([]byte, 7)
				val[0] = '#'
				hex.Encode(val[1:], rgb[:])
				parse.ToLower(val)
				if s, ok := ShortenColorHex[string(val)]; ok {
					if _, err := c.w.Write(s); err != nil {
						return 0, err
					}
				} else {
					if len(val) == 7 && val[1] == val[2] && val[3] == val[4] && val[5] == val[6] {
						val[2] = val[3]
						val[3] = val[5]
						val = val[:4]
					}
					if _, err := c.w.Write(val); err != nil {
						return 0, err
					}
				}
				return n, nil
			}
		} else if fun == css.Hsl && nArgs == 3 {
			if values[1].TokenType == css.NumberToken && values[3].TokenType == css.PercentageToken && values[5].TokenType == css.PercentageToken {
				h, err1 := strconv.ParseFloat(string(values[1].Data), 32)
				s, err2 := strconv.ParseFloat(string(values[3].Data[:len(values[3].Data)-1]), 32)
				l, err3 := strconv.ParseFloat(string(values[5].Data[:len(values[5].Data)-1]), 32)
				if err1 == nil && err2 == nil && err3 == nil {
					r, g, b := css.HSL2RGB(h/360.0, s/100.0, l/100.0)
					rgb := []byte{byte((r * 255.0) + 0.5), byte((g * 255.0) + 0.5), byte((b * 255.0) + 0.5)}
					val := make([]byte, 7)
					val[0] = '#'
					hex.Encode(val[1:], rgb[:])
					parse.ToLower(val)
					if s, ok := ShortenColorHex[string(val)]; ok {
						if _, err := c.w.Write(s); err != nil {
							return 0, err
						}
					} else {
						if len(val) == 7 && val[1] == val[2] && val[3] == val[4] && val[5] == val[6] {
							val[2] = val[3]
							val[3] = val[5]
							val = val[:4]
						}
						if _, err := c.w.Write(val); err != nil {
							return 0, err
						}
					}
					return n, nil
				}
			}
		}
	}
	for _, value := range values {
		if _, err := c.w.Write(value.Data); err != nil {
			return 0, err
		}
	}
	return n, nil
}

func (c *cssMinifier) shortenToken(prop css.Hash, tt css.TokenType, data []byte) (css.TokenType, []byte) {
	if tt == css.NumberToken || tt == css.PercentageToken || tt == css.DimensionToken {
		if tt == css.NumberToken && (prop == css.Z_Index || prop == css.Counter_Increment || prop == css.Counter_Reset || prop == css.Orphans || prop == css.Widows) {
			return tt, data // integers
		}
		n := len(data)
		if tt == css.PercentageToken {
			n--
		} else if tt == css.DimensionToken {
			n = parse.Number(data)
		}
		dim := data[n:]
		parse.ToLower(dim)
		data = minify.Number(data[:n], c.o.Decimals)
		if tt == css.PercentageToken && (len(data) != 1 || data[0] != '0' || prop == css.Color) {
			data = append(data, '%')
		} else if tt == css.DimensionToken && (len(data) != 1 || data[0] != '0' || requiredDimension[string(dim)]) {
			data = append(data, dim...)
		}
	} else if tt == css.IdentToken {
		//parse.ToLower(data) // TODO: not all identifiers are case-insensitive; all <custom-ident> properties are case-sensitive
		if hex, ok := ShortenColorName[css.ToHash(data)]; ok {
			tt = css.HashToken
			data = hex
		}
	} else if tt == css.HashToken {
		parse.ToLower(data)
		if ident, ok := ShortenColorHex[string(data)]; ok {
			tt = css.IdentToken
			data = ident
		} else if len(data) == 7 && data[1] == data[2] && data[3] == data[4] && data[5] == data[6] {
			tt = css.HashToken
			data[2] = data[3]
			data[3] = data[5]
			data = data[:4]
		}
	} else if tt == css.StringToken {
		// remove any \\\r\n \\\r \\\n
		for i := 1; i < len(data)-2; i++ {
			if data[i] == '\\' && (data[i+1] == '\n' || data[i+1] == '\r') {
				// encountered first replacee, now start to move bytes to the front
				j := i + 2
				if data[i+1] == '\r' && len(data) > i+2 && data[i+2] == '\n' {
					j++
				}
				for ; j < len(data); j++ {
					if data[j] == '\\' && len(data) > j+1 && (data[j+1] == '\n' || data[j+1] == '\r') {
						if data[j+1] == '\r' && len(data) > j+2 && data[j+2] == '\n' {
							j++
						}
						j++
					} else {
						data[i] = data[j]
						i++
					}
				}
				data = data[:i]
				break
			}
		}
	} else if tt == css.URLToken {
		parse.ToLower(data[:3])
		if len(data) > 10 {
			uri := data[4 : len(data)-1]
			delim := byte('"')
			if uri[0] == '\'' || uri[0] == '"' {
				delim = uri[0]
				uri = uri[1 : len(uri)-1]
			}
			uri = minify.DataURI(c.m, uri)
			if css.IsURLUnquoted(uri) {
				data = append(append([]byte("url("), uri...), ')')
			} else {
				data = append(append(append([]byte("url("), delim), uri...), delim, ')')
			}
		}
	}
	return tt, data
}

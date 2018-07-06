// Package html minifies HTML5 following the specifications at http://www.w3.org/TR/html5/syntax.html.
package html // import "github.com/tdewolff/minify/html"

import (
	"bytes"
	"io"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/parse"
	"github.com/tdewolff/parse/buffer"
	"github.com/tdewolff/parse/html"
)

var (
	gtBytes         = []byte(">")
	isBytes         = []byte("=")
	spaceBytes      = []byte(" ")
	doctypeBytes    = []byte("<!doctype html>")
	jsMimeBytes     = []byte("text/javascript")
	cssMimeBytes    = []byte("text/css")
	htmlMimeBytes   = []byte("text/html")
	svgMimeBytes    = []byte("image/svg+xml")
	mathMimeBytes   = []byte("application/mathml+xml")
	dataSchemeBytes = []byte("data:")
	jsSchemeBytes   = []byte("javascript:")
	httpBytes       = []byte("http")
)

////////////////////////////////////////////////////////////////

// DefaultMinifier is the default minifier.
var DefaultMinifier = &Minifier{}

// Minifier is an HTML minifier.
type Minifier struct {
	KeepConditionalComments bool
	KeepDefaultAttrVals     bool
	KeepDocumentTags        bool
	KeepEndTags             bool
	KeepWhitespace          bool
}

// Minify minifies HTML data, it reads from r and writes to w.
func Minify(m *minify.M, w io.Writer, r io.Reader, params map[string]string) error {
	return DefaultMinifier.Minify(m, w, r, params)
}

// Minify minifies HTML data, it reads from r and writes to w.
func (o *Minifier) Minify(m *minify.M, w io.Writer, r io.Reader, _ map[string]string) error {
	var rawTagHash html.Hash
	var rawTagMediatype []byte

	omitSpace := true // if true the next leading space is omitted
	inPre := false

	defaultScriptType := jsMimeBytes
	defaultScriptParams := map[string]string(nil)
	defaultStyleType := cssMimeBytes
	defaultStyleParams := map[string]string(nil)
	defaultInlineStyleParams := map[string]string{"inline": "1"}

	attrMinifyBuffer := buffer.NewWriter(make([]byte, 0, 64))
	attrByteBuffer := make([]byte, 0, 64)

	l := html.NewLexer(r)
	defer l.Restore()

	tb := NewTokenBuffer(l)
	for {
		t := *tb.Shift()
	SWITCH:
		switch t.TokenType {
		case html.ErrorToken:
			if l.Err() == io.EOF {
				return nil
			}
			return l.Err()
		case html.DoctypeToken:
			if _, err := w.Write(doctypeBytes); err != nil {
				return err
			}
		case html.CommentToken:
			if o.KeepConditionalComments && len(t.Text) > 6 && (bytes.HasPrefix(t.Text, []byte("[if ")) || bytes.Equal(t.Text, []byte("[endif]")) || bytes.Equal(t.Text, []byte("<![endif]"))) {
				// [if ...] is always 7 or more characters, [endif] is only encountered for downlevel-revealed
				// see https://msdn.microsoft.com/en-us/library/ms537512(v=vs.85).aspx#syntax
				if bytes.HasPrefix(t.Data, []byte("<!--[if ")) && len(t.Data) > len("<!--[if ]><![endif]-->") { // downlevel-hidden
					begin := bytes.IndexByte(t.Data, '>') + 1
					end := len(t.Data) - len("<![endif]-->")
					if _, err := w.Write(t.Data[:begin]); err != nil {
						return err
					}
					if err := o.Minify(m, w, buffer.NewReader(t.Data[begin:end]), nil); err != nil {
						return err
					}
					if _, err := w.Write(t.Data[end:]); err != nil {
						return err
					}
				} else if _, err := w.Write(t.Data); err != nil { // downlevel-revealed or short downlevel-hidden
					return err
				}
			}
		case html.SvgToken:
			if err := m.MinifyMimetype(svgMimeBytes, w, buffer.NewReader(t.Data), nil); err != nil {
				if err != minify.ErrNotExist {
					return err
				} else if _, err := w.Write(t.Data); err != nil {
					return err
				}
			}
		case html.MathToken:
			if err := m.MinifyMimetype(mathMimeBytes, w, buffer.NewReader(t.Data), nil); err != nil {
				if err != minify.ErrNotExist {
					return err
				} else if _, err := w.Write(t.Data); err != nil {
					return err
				}
			}
		case html.TextToken:
			// CSS and JS minifiers for inline code
			if rawTagHash != 0 {
				if rawTagHash == html.Style || rawTagHash == html.Script || rawTagHash == html.Iframe {
					var mimetype []byte
					var params map[string]string
					if rawTagHash == html.Iframe {
						mimetype = htmlMimeBytes
					} else if len(rawTagMediatype) > 0 {
						mimetype, params = parse.Mediatype(rawTagMediatype)
					} else if rawTagHash == html.Script {
						mimetype = defaultScriptType
						params = defaultScriptParams
					} else if rawTagHash == html.Style {
						mimetype = defaultStyleType
						params = defaultStyleParams
					}
					if err := m.MinifyMimetype(mimetype, w, buffer.NewReader(t.Data), params); err != nil {
						if err != minify.ErrNotExist {
							return err
						} else if _, err := w.Write(t.Data); err != nil {
							return err
						}
					}
				} else if _, err := w.Write(t.Data); err != nil {
					return err
				}
			} else if inPre {
				if _, err := w.Write(t.Data); err != nil {
					return err
				}
			} else {
				t.Data = parse.ReplaceMultipleWhitespace(t.Data)

				// whitespace removal; trim left
				if omitSpace && (t.Data[0] == ' ' || t.Data[0] == '\n') {
					t.Data = t.Data[1:]
				}

				// whitespace removal; trim right
				omitSpace = false
				if len(t.Data) == 0 {
					omitSpace = true
				} else if t.Data[len(t.Data)-1] == ' ' || t.Data[len(t.Data)-1] == '\n' {
					omitSpace = true
					i := 0
					for {
						next := tb.Peek(i)
						// trim if EOF, text token with leading whitespace or block token
						if next.TokenType == html.ErrorToken {
							t.Data = t.Data[:len(t.Data)-1]
							omitSpace = false
							break
						} else if next.TokenType == html.TextToken {
							// this only happens when a comment, doctype or phrasing end tag (only for !o.KeepWhitespace) was in between
							// remove if the text token starts with a whitespace
							if len(next.Data) > 0 && parse.IsWhitespace(next.Data[0]) {
								t.Data = t.Data[:len(t.Data)-1]
								omitSpace = false
							}
							break
						} else if next.TokenType == html.StartTagToken || next.TokenType == html.EndTagToken {
							if o.KeepWhitespace {
								break
							}
							// remove when followed up by a block tag
							if next.Traits&nonPhrasingTag != 0 {
								t.Data = t.Data[:len(t.Data)-1]
								omitSpace = false
								break
							} else if next.TokenType == html.StartTagToken {
								break
							}
						}
						i++
					}
				}

				if _, err := w.Write(t.Data); err != nil {
					return err
				}
			}
		case html.StartTagToken, html.EndTagToken:
			rawTagHash = 0
			hasAttributes := false
			if t.TokenType == html.StartTagToken {
				if next := tb.Peek(0); next.TokenType == html.AttributeToken {
					hasAttributes = true
				}
				if t.Traits&rawTag != 0 {
					// ignore empty script and style tags
					if !hasAttributes && (t.Hash == html.Script || t.Hash == html.Style) {
						if next := tb.Peek(1); next.TokenType == html.EndTagToken {
							tb.Shift()
							tb.Shift()
							break
						}
					}
					rawTagHash = t.Hash
					rawTagMediatype = nil
				}
			} else if t.Hash == html.Template {
				omitSpace = true // EndTagToken
			}

			if t.Hash == html.Pre {
				inPre = t.TokenType == html.StartTagToken
			}

			// remove superfluous tags, except for html, head and body tags when KeepDocumentTags is set
			if !hasAttributes && (!o.KeepDocumentTags && (t.Hash == html.Html || t.Hash == html.Head || t.Hash == html.Body) || t.Hash == html.Colgroup) {
				break
			} else if t.TokenType == html.EndTagToken {
				if !o.KeepEndTags {
					if t.Hash == html.Thead || t.Hash == html.Tbody || t.Hash == html.Tfoot || t.Hash == html.Tr || t.Hash == html.Th || t.Hash == html.Td ||
						t.Hash == html.Optgroup || t.Hash == html.Option || t.Hash == html.Dd || t.Hash == html.Dt ||
						t.Hash == html.Li || t.Hash == html.Rb || t.Hash == html.Rt || t.Hash == html.Rtc || t.Hash == html.Rp {
						break
					} else if t.Hash == html.P {
						i := 0
						for {
							next := tb.Peek(i)
							i++
							// continue if text token is empty or whitespace
							if next.TokenType == html.TextToken && parse.IsAllWhitespace(next.Data) {
								continue
							}
							if next.TokenType == html.ErrorToken || next.TokenType == html.EndTagToken && next.Traits&keepPTag == 0 || next.TokenType == html.StartTagToken && next.Traits&omitPTag != 0 {
								break SWITCH // omit p end tag
							}
							break
						}
					}
				}

				if o.KeepWhitespace || t.Traits&objectTag != 0 {
					omitSpace = false
				} else if t.Traits&nonPhrasingTag != 0 {
					omitSpace = true // omit spaces after block elements
				}

				if len(t.Data) > 3+len(t.Text) {
					t.Data[2+len(t.Text)] = '>'
					t.Data = t.Data[:3+len(t.Text)]
				}
				if _, err := w.Write(t.Data); err != nil {
					return err
				}
				break
			}

			if o.KeepWhitespace || t.Traits&objectTag != 0 {
				omitSpace = false
			} else if t.Traits&nonPhrasingTag != 0 {
				omitSpace = true // omit spaces after block elements
			}

			if _, err := w.Write(t.Data); err != nil {
				return err
			}

			if hasAttributes {
				if t.Hash == html.Meta {
					attrs := tb.Attributes(html.Content, html.Http_Equiv, html.Charset, html.Name)
					if content := attrs[0]; content != nil {
						if httpEquiv := attrs[1]; httpEquiv != nil {
							if charset := attrs[2]; charset == nil && parse.EqualFold(httpEquiv.AttrVal, []byte("content-type")) {
								content.AttrVal = minify.Mediatype(content.AttrVal)
								if bytes.Equal(content.AttrVal, []byte("text/html;charset=utf-8")) {
									httpEquiv.Text = nil
									content.Text = []byte("charset")
									content.Hash = html.Charset
									content.AttrVal = []byte("utf-8")
								}
							} else if parse.EqualFold(httpEquiv.AttrVal, []byte("content-style-type")) {
								content.AttrVal = minify.Mediatype(content.AttrVal)
								defaultStyleType, defaultStyleParams = parse.Mediatype(content.AttrVal)
								if defaultStyleParams != nil {
									defaultInlineStyleParams = defaultStyleParams
									defaultInlineStyleParams["inline"] = "1"
								} else {
									defaultInlineStyleParams = map[string]string{"inline": "1"}
								}
							} else if parse.EqualFold(httpEquiv.AttrVal, []byte("content-script-type")) {
								content.AttrVal = minify.Mediatype(content.AttrVal)
								defaultScriptType, defaultScriptParams = parse.Mediatype(content.AttrVal)
							}
						}
						if name := attrs[3]; name != nil {
							if parse.EqualFold(name.AttrVal, []byte("keywords")) {
								content.AttrVal = bytes.Replace(content.AttrVal, []byte(", "), []byte(","), -1)
							} else if parse.EqualFold(name.AttrVal, []byte("viewport")) {
								content.AttrVal = bytes.Replace(content.AttrVal, []byte(" "), []byte(""), -1)
								for i := 0; i < len(content.AttrVal); i++ {
									if content.AttrVal[i] == '=' && i+2 < len(content.AttrVal) {
										i++
										if n := parse.Number(content.AttrVal[i:]); n > 0 {
											minNum := minify.Number(content.AttrVal[i:i+n], -1)
											if len(minNum) < n {
												copy(content.AttrVal[i:i+len(minNum)], minNum)
												copy(content.AttrVal[i+len(minNum):], content.AttrVal[i+n:])
												content.AttrVal = content.AttrVal[:len(content.AttrVal)+len(minNum)-n]
											}
											i += len(minNum)
										}
										i-- // mitigate for-loop increase
									}
								}
							}
						}
					}
				} else if t.Hash == html.Script {
					attrs := tb.Attributes(html.Src, html.Charset)
					if attrs[0] != nil && attrs[1] != nil {
						attrs[1].Text = nil
					}
				}

				// write attributes
				htmlEqualIdName := false
				for {
					attr := *tb.Shift()
					if attr.TokenType != html.AttributeToken {
						break
					} else if attr.Text == nil {
						continue // removed attribute
					}

					if t.Hash == html.A && (attr.Hash == html.Id || attr.Hash == html.Name) {
						if attr.Hash == html.Id {
							if name := tb.Attributes(html.Name)[0]; name != nil && bytes.Equal(attr.AttrVal, name.AttrVal) {
								htmlEqualIdName = true
							}
						} else if htmlEqualIdName {
							continue
						} else if id := tb.Attributes(html.Id)[0]; id != nil && bytes.Equal(id.AttrVal, attr.AttrVal) {
							continue
						}
					}

					val := attr.AttrVal
					if len(val) == 0 && (attr.Hash == html.Class ||
						attr.Hash == html.Dir ||
						attr.Hash == html.Id ||
						attr.Hash == html.Lang ||
						attr.Hash == html.Name ||
						attr.Hash == html.Title ||
						attr.Hash == html.Action && t.Hash == html.Form ||
						attr.Hash == html.Value && t.Hash == html.Input) {
						continue // omit empty attribute values
					}
					if attr.Traits&caselessAttr != 0 {
						val = parse.ToLower(val)
						if attr.Hash == html.Enctype || attr.Hash == html.Codetype || attr.Hash == html.Accept || attr.Hash == html.Type && (t.Hash == html.A || t.Hash == html.Link || t.Hash == html.Object || t.Hash == html.Param || t.Hash == html.Script || t.Hash == html.Style || t.Hash == html.Source) {
							val = minify.Mediatype(val)
						}
					}
					if rawTagHash != 0 && attr.Hash == html.Type {
						rawTagMediatype = parse.Copy(val)
					}

					// default attribute values can be omitted
					if !o.KeepDefaultAttrVals && (attr.Hash == html.Type && (t.Hash == html.Script && bytes.Equal(val, []byte("text/javascript")) ||
						t.Hash == html.Style && bytes.Equal(val, []byte("text/css")) ||
						t.Hash == html.Link && bytes.Equal(val, []byte("text/css")) ||
						t.Hash == html.Input && bytes.Equal(val, []byte("text")) ||
						t.Hash == html.Button && bytes.Equal(val, []byte("submit"))) ||
						attr.Hash == html.Language && t.Hash == html.Script ||
						attr.Hash == html.Method && bytes.Equal(val, []byte("get")) ||
						attr.Hash == html.Enctype && bytes.Equal(val, []byte("application/x-www-form-urlencoded")) ||
						attr.Hash == html.Colspan && bytes.Equal(val, []byte("1")) ||
						attr.Hash == html.Rowspan && bytes.Equal(val, []byte("1")) ||
						attr.Hash == html.Shape && bytes.Equal(val, []byte("rect")) ||
						attr.Hash == html.Span && bytes.Equal(val, []byte("1")) ||
						attr.Hash == html.Clear && bytes.Equal(val, []byte("none")) ||
						attr.Hash == html.Frameborder && bytes.Equal(val, []byte("1")) ||
						attr.Hash == html.Scrolling && bytes.Equal(val, []byte("auto")) ||
						attr.Hash == html.Valuetype && bytes.Equal(val, []byte("data")) ||
						attr.Hash == html.Media && t.Hash == html.Style && bytes.Equal(val, []byte("all"))) {
						continue
					}

					// CSS and JS minifiers for attribute inline code
					if attr.Hash == html.Style {
						attrMinifyBuffer.Reset()
						if err := m.MinifyMimetype(defaultStyleType, attrMinifyBuffer, buffer.NewReader(val), defaultInlineStyleParams); err == nil {
							val = attrMinifyBuffer.Bytes()
						} else if err != minify.ErrNotExist {
							return err
						}
						if len(val) == 0 {
							continue
						}
					} else if len(attr.Text) > 2 && attr.Text[0] == 'o' && attr.Text[1] == 'n' {
						if len(val) >= 11 && parse.EqualFold(val[:11], jsSchemeBytes) {
							val = val[11:]
						}
						attrMinifyBuffer.Reset()
						if err := m.MinifyMimetype(defaultScriptType, attrMinifyBuffer, buffer.NewReader(val), defaultScriptParams); err == nil {
							val = attrMinifyBuffer.Bytes()
						} else if err != minify.ErrNotExist {
							return err
						}
						if len(val) == 0 {
							continue
						}
					} else if len(val) > 5 && attr.Traits&urlAttr != 0 { // anchors are already handled
						if parse.EqualFold(val[:4], httpBytes) {
							if val[4] == ':' {
								if m.URL != nil && m.URL.Scheme == "http" {
									val = val[5:]
								} else {
									parse.ToLower(val[:4])
								}
							} else if (val[4] == 's' || val[4] == 'S') && val[5] == ':' {
								if m.URL != nil && m.URL.Scheme == "https" {
									val = val[6:]
								} else {
									parse.ToLower(val[:5])
								}
							}
						} else if parse.EqualFold(val[:5], dataSchemeBytes) {
							val = minify.DataURI(m, val)
						}
					}

					if _, err := w.Write(spaceBytes); err != nil {
						return err
					}
					if _, err := w.Write(attr.Text); err != nil {
						return err
					}
					if len(val) > 0 && attr.Traits&booleanAttr == 0 {
						if _, err := w.Write(isBytes); err != nil {
							return err
						}
						// no quotes if possible, else prefer single or double depending on which occurs more often in value
						val = html.EscapeAttrVal(&attrByteBuffer, attr.AttrVal, val)
						if _, err := w.Write(val); err != nil {
							return err
						}
					}
				}
			}
			if _, err := w.Write(gtBytes); err != nil {
				return err
			}
		}
	}
}

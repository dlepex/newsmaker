package words

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Pattern struct {
	re   *regexp.Regexp
	expr string
}

var matchAny *regexp.Regexp = regexp.MustCompile("")
var emptyPattern Pattern = Pattern{}
var metaMap map[rune]string = make(map[rune]string)

func init() {
	initMetaMap()
}

func metaRegexp(b *bytes.Buffer, s string) string {
	ff := strings.Fields(s)
	b.WriteString(ff[0])
	for _, f := range ff[1:] {
		b.WriteRune('|')
		b.WriteString(f)
	}
	b.WriteString(")$")
	return b.String()
}

func initMetaMap() {

	m := make(map[rune]string)

	m['а'] = "а у ы е ой"
	m['А'] = "а у ы е ой   ы ам ах ами"

	m['я'] = "я ю и ей"
	m['Я'] = "я ю и ей   и ь ям ями ях"

	// -ия -ий
	m['и'] = "ия ий ие ию ии ией ием "
	m['И'] = "ия ий ие ию ии ией ием  ий иев иям иями иях"

	m['ъ'] = "е а у ом"
	m['Ъ'] = "е а у ом   ы и ов ей ам ами ах"

	m['o'] = "o е а у ом"
	m['O'] = "o е а у ом   ы и ов ей ам ами ах"

	m['ь'] = "ь й е ё я ю ью и ем ём"
	m['Ь'] = "ь й е ё я ю ью и ем ём   ей ев ёв ям ам ями ами ах ях"

	b := bytes.NewBuffer(make([]byte, 0, 128))
	for k, v := range m {
		b.Reset()
		b.WriteString("(?i:")
		if k == 'ъ' || k == 'Ъ' {
			b.WriteRune('|')
		}
		metaMap[k] = metaRegexp(b, v)
	}
	metaMap['a'] = metaMap['а']
	metaMap['A'] = metaMap['А']

	metaMap['е'] = metaMap['ь']
	metaMap['Е'] = metaMap['Ь']

	metaMap['e'] = metaMap['ь']
	metaMap['E'] = metaMap['Ь']
}

func NewPattern(expr string) (Pattern, error) {
	if expr == "" {
		return Pattern{matchAny, ""}, nil
	}

	b := bytes.NewBuffer(make([]byte, 0, len(expr)*2))
	s := expr
	if s[0] != '*' {
		b.WriteRune('^')
	} else {
		s = s[1:]
	}

	braces := 0
	iflag := false
	dollarPos := 0
loop:
	for i, r := range s {
		switch r {
		case '(':
			braces++
			if iflag {
				b.WriteRune(')')
				iflag = false
			}
			b.WriteString("(?i:")
		case ')':
			if braces == 0 {
				return emptyPattern, fmt.Errorf("wrong ) at: %d", i)
			}
			braces--
			b.WriteRune(r)
		case '*':
			if braces == 0 {
				return emptyPattern, fmt.Errorf("wrong * at: %d. '*' must be first char, or inside regexp braces()", i)
			}
			b.WriteRune(r)
		case '$':
			if braces > 0 {
				return emptyPattern, fmt.Errorf("$ can't be used inside braces() at: %d", i)
			}
			dollarPos = i
			break loop
		default:
			if braces == 0 && unicode.IsLetter(r) {
				lower := unicode.IsLower(r)
				if lower != iflag {
					if iflag {
						b.WriteRune(')')
					} else {
						b.WriteString("(?i:")
					}
				}
				iflag = lower
			}
			b.WriteRune(r)
		}
	}
	if iflag {
		b.WriteRune(')')
	}

	if dollarPos != 0 {
		pos := dollarPos + utf8.RuneLen('$')
		switch len(s) - pos {
		case 0:
			b.WriteRune('$')
		case 1, 2:
			meta, _ := utf8.DecodeLastRuneInString(s)
			if str, ok := metaMap[meta]; ok {
				b.WriteString(str)
			} else {
				return emptyPattern, fmt.Errorf("invalid metachar after $: %v", meta)
			}
		default:
			return emptyPattern, fmt.Errorf("$ must be last or penultimate char (if using meta chars for morphology))")
		}
	}

	re, err := regexp.Compile(b.String())
	if err != nil {
		return emptyPattern, fmt.Errorf("cant compile regexp: %s", err)
	}

	return Pattern{re, expr}, nil

}

func (p Pattern) String() string {
	if p.re == nil {
		return "<error>"
	}
	return fmt.Sprintf("%s ==> %s", p.expr, p.re.String())
}

func (p Pattern) Match(s string) bool {
	return p.re.MatchString(s)
}

func Split(s string) []string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return nil
	}
	result := words[:0]
	for _, w := range words {
		w = strings.TrimFunc(w, unicode.IsPunct)
		if len(w) > 0 {
			result = append(result, w)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

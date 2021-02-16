package json

const (
	ObjectStart = '{' // {
	ObjectEnd   = '}' // }
	String      = '"' // "
	Colon       = ':' // :
	Comma       = ',' // ,
	ArrayStart  = '[' // [
	ArrayEnd    = ']' // ]
	True        = 't' // t
	False       = 'f' // f
	Null        = 'n' // n
)

// NewScanner returns a new Scanner for the data frame.
// A Scanner reads the supplied data frame and produces via Next a stream
// of tokens, expressed as []byte slices.
func NewScanner(data []byte) *Scanner {
	return &Scanner{
		br: byteReader{
			data: data,
		},
	}
}

// Scanner implements a JSON scanner as defined in RFC 7159.
type Scanner struct {
	br  byteReader
	pos int
}

var whitespace = [256]bool{
	' ':  true,
	'\r': true,
	'\n': true,
	'\t': true,
}

// Next returns a []byte referencing the the next lexical token in the stream.
// If the stream is at its end, or an error has occured, Next returns a zero
// length []byte slice.
//
// A valid token begins with one of the following:
//
//  { Object start
//  [ Array start
//  } Object end
//  ] Array End
//  , Literal comma
//  : Literal colon
//  t JSON true
//  f JSON false
//  n JSON null
//  " A string, possibly containing backslash escaped entites.
//  -, 0-9 A number
func (s *Scanner) Next() []byte {
	s.br.release(s.pos)
	w := s.br.window(0)
	for pos, c := range w {
		// strip any leading whitespace.
		if whitespace[c] {
			continue
		}

		// simple case
		switch c {
		case ObjectStart, ObjectEnd, Colon, Comma, ArrayStart, ArrayEnd:
			s.pos = pos + 1
			return w[pos:s.pos]
		}

		s.br.release(pos)
		switch c {
		case True:
			s.pos = validateToken(&s.br, "true")
		case False:
			s.pos = validateToken(&s.br, "false")
		case Null:
			s.pos = validateToken(&s.br, "null")
		case String:
			if s.parseString() < 2 {
				return nil
			}
		default:
			// ensure the number is correct.
			s.pos = s.parseNumber(c)
		}
		return s.br.window(0)[:s.pos]
	}

	// it's all whitespace, ignore it
	s.br.release(len(w))

	// eof
	return nil
}

func validateToken(br *byteReader, expected string) int {
	w := br.window(0)
	n := len(expected)
	if len(w) < n {
		// Insufficient data.
		return 0
	}
	if string(w[:n]) != expected {
		// Doesn't match.
		return 0
	}
	return n
}

// parseString returns the length of the string token
// located at the start of the window or 0 if there is no closing
// " before the end of the byteReader.
func (s *Scanner) parseString() int {
	escaped := false
	w := s.br.window(1)
	pos := 0
	{
		for _, c := range w {
			pos++
			switch {
			case escaped:
				escaped = false
			case c == '"':
				// finished
				s.pos = pos + 1
				return s.pos
			case c == '\\':
				escaped = true
			}
		}
		// EOF.
		return 0
	}
}

func (s *Scanner) parseNumber(c byte) int {
	const (
		begin = iota
		leadingzero
		anydigit1
		decimal
		anydigit2
		exponent
		expsign
		anydigit3
	)

	pos := 0
	w := s.br.window(0)
	// int vs uint8 costs 10% on canada.json
	var state uint8 = begin

	// handle the case that the first character is a hyphen
	if c == '-' {
		pos++
		w = s.br.window(1)
	}
	{
		for _, elem := range w {
			switch state {
			case begin:
				if elem >= '1' && elem <= '9' {
					state = anydigit1
				} else if elem == '0' {
					state = leadingzero
				} else {
					// error
					return 0
				}
			case anydigit1:
				if elem >= '0' && elem <= '9' {
					// stay in this state
					break
				}
				fallthrough
			case leadingzero:
				if elem == '.' {
					state = decimal
					break
				}
				if elem == 'e' || elem == 'E' {
					state = exponent
					break
				}
				return pos // finished.
			case decimal:
				if elem >= '0' && elem <= '9' {
					state = anydigit2
				} else {
					// error
					return 0
				}
			case anydigit2:
				if elem >= '0' && elem <= '9' {
					break
				}
				if elem == 'e' || elem == 'E' {
					state = exponent
					break
				}
				return pos // finished.
			case exponent:
				if elem == '+' || elem == '-' {
					state = expsign
					break
				}
				fallthrough
			case expsign:
				if elem >= '0' && elem <= '9' {
					state = anydigit3
					break
				}
				// error
				return 0
			case anydigit3:
				if elem < '0' || elem > '9' {
					return pos
				}
			}
			pos++
		}
		// end of the item. However, not necessarily an error. Make
		// sure we are in a state that allows ending the number.
		switch state {
		case leadingzero, anydigit1, anydigit2, anydigit3:
			return pos
		default:
			// error otherwise, the number isn't complete.
			return 0
		}
	}
}

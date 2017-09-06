package cloudformation

//import "fmt"

type lexer struct {
	input        []byte
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
}

func newLexer(input []byte) *lexer {
	l := &lexer{input: input}
	return l
}

func (l *lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
}

func isBlockStart(ch byte) bool {
	return '{' == ch
}

func isBlockEnd(ch byte) bool {
	return '{' == ch
}
func (l *lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	} else {
		return l.input[l.readPosition]
	}
}

func (l *lexer) blockKey() string {

	position := l.position
	instring := false
	strstart := 0
	for position = l.position; position < len(l.input) && '}' != l.input[position]; position = position + 1 {
		if '\\' == l.input[position] {
			position = position + 1
		} else {
			if instring && '"' == l.input[position] {
				break
			}
			if !instring && '"' == l.input[position] {
				strstart = position + 1
				instring = !instring

			}

		}
	}

	return string(l.input[strstart:position])
}

func (l *lexer) readBlock() []byte {

	position := l.position
	instring := false
	lbrace := 0
	for ch := l.peekChar(); ch != 0; ch = l.peekChar() {
		if '\\' == ch {
			l.readChar()
		} else {
			if '"' == ch {
				instring = !instring
			}

			if !instring {
				if '{' == ch {
					lbrace = lbrace + 1
				} else if '}' == ch {
					lbrace = lbrace - 1
				}
			}
		}
		l.readChar()
		if 0 == lbrace {
			break
		}
	}

	return l.input[position:l.readPosition]
}

package asm
import (
	"fmt"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"
)
type stateFn func(*lexer) stateFn
type token struct {
	typ    tokenType
	lineno int
	text   string
}
type tokenType int
const (
	eof              tokenType = iota 
	lineStart                         
	lineEnd                           
	invalidStatement                  
	element                           
	label                             
	labelDef                          
	number                            
	stringValue                       
	Numbers            = "1234567890"                                           
	HexadecimalNumbers = Numbers + "aAbBcCdDeEfF"                               
	Alpha              = "abcdefghijklmnopqrstuwvxyzABCDEFGHIJKLMNOPQRSTUWVXYZ" 
)
func (it tokenType) String() string {
	if int(it) > len(stringtokenTypes) {
		return "invalid"
	}
	return stringtokenTypes[it]
}
var stringtokenTypes = []string{
	eof:              "EOF",
	invalidStatement: "invalid statement",
	element:          "element",
	lineEnd:          "end of line",
	lineStart:        "new line",
	label:            "label",
	labelDef:         "label definition",
	number:           "number",
	stringValue:      "string",
}
type lexer struct {
	input string 
	tokens chan token 
	state  stateFn    
	lineno            int 
	start, pos, width int 
	debug bool 
}
func Lex(source []byte, debug bool) <-chan token {
	ch := make(chan token)
	l := &lexer{
		input:  string(source),
		tokens: ch,
		state:  lexLine,
		debug:  debug,
	}
	go func() {
		l.emit(lineStart)
		for l.state != nil {
			l.state = l.state(l)
		}
		l.emit(eof)
		close(l.tokens)
	}()
	return ch
}
func (l *lexer) next() (rune rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return 0
	}
	rune, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return rune
}
func (l *lexer) backup() {
	l.pos -= l.width
}
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}
func (l *lexer) ignore() {
	l.start = l.pos
}
func (l *lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}
func (l *lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}
	l.backup()
}
func (l *lexer) acceptRunUntil(until rune) bool {
	for i := l.next(); !strings.ContainsRune(string(until), i); i = l.next() {
		if i == 0 {
			return false
		}
	}
	return true
}
func (l *lexer) blob() string {
	return l.input[l.start:l.pos]
}
func (l *lexer) emit(t tokenType) {
	token := token{t, l.lineno, l.blob()}
	if l.debug {
		fmt.Fprintf(os.Stderr, "%04d: (%-20v) %s\n", token.lineno, token.typ, token.text)
	}
	l.tokens <- token
	l.start = l.pos
}
func lexLine(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == '\n':
			l.emit(lineEnd)
			l.ignore()
			l.lineno++
			l.emit(lineStart)
		case r == ';' && l.peek() == ';':
			return lexComment
		case isSpace(r):
			l.ignore()
		case isLetter(r) || r == '_':
			return lexElement
		case isNumber(r):
			return lexNumber
		case r == '@':
			l.ignore()
			return lexLabel
		case r == '"':
			return lexInsideString
		default:
			return nil
		}
	}
}
func lexComment(l *lexer) stateFn {
	l.acceptRunUntil('\n')
	l.ignore()
	return lexLine
}
func lexLabel(l *lexer) stateFn {
	l.acceptRun(Alpha + "_" + Numbers)
	l.emit(label)
	return lexLine
}
func lexInsideString(l *lexer) stateFn {
	if l.acceptRunUntil('"') {
		l.emit(stringValue)
	}
	return lexLine
}
func lexNumber(l *lexer) stateFn {
	acceptance := Numbers
	if l.accept("0") || l.accept("xX") {
		acceptance = HexadecimalNumbers
	}
	l.acceptRun(acceptance)
	l.emit(number)
	return lexLine
}
func lexElement(l *lexer) stateFn {
	l.acceptRun(Alpha + "_" + Numbers)
	if l.peek() == ':' {
		l.emit(labelDef)
		l.accept(":")
		l.ignore()
	} else {
		l.emit(element)
	}
	return lexLine
}
func isLetter(t rune) bool {
	return unicode.IsLetter(t)
}
func isSpace(t rune) bool {
	return unicode.IsSpace(t)
}
func isNumber(t rune) bool {
	return unicode.IsNumber(t)
}

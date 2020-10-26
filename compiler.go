package asm
import (
	"fmt"
	"math/big"
	"os"
	"strings"
	"github.com/Cryptochain-VON/common/math"
	"github.com/Cryptochain-VON/core/vm"
)
type Compiler struct {
	tokens []token
	binary []interface{}
	labels map[string]int
	pc, pos int
	debug bool
}
func NewCompiler(debug bool) *Compiler {
	return &Compiler{
		labels: make(map[string]int),
		debug:  debug,
	}
}
func (c *Compiler) Feed(ch <-chan token) {
	var prev token
	for i := range ch {
		switch i.typ {
		case number:
			num := math.MustParseBig256(i.text).Bytes()
			if len(num) == 0 {
				num = []byte{0}
			}
			c.pc += len(num)
		case stringValue:
			c.pc += len(i.text) - 2
		case element:
			c.pc++
		case labelDef:
			c.labels[i.text] = c.pc
			c.pc++
		case label:
			c.pc += 4
			if prev.typ == element && isJump(prev.text) {
				c.pc++
			}
		}
		c.tokens = append(c.tokens, i)
		prev = i
	}
	if c.debug {
		fmt.Fprintln(os.Stderr, "found", len(c.labels), "labels")
	}
}
func (c *Compiler) Compile() (string, []error) {
	var errors []error
	for c.pos < len(c.tokens) {
		if err := c.compileLine(); err != nil {
			errors = append(errors, err)
		}
	}
	var bin string
	for _, v := range c.binary {
		switch v := v.(type) {
		case vm.OpCode:
			bin += fmt.Sprintf("%x", []byte{byte(v)})
		case []byte:
			bin += fmt.Sprintf("%x", v)
		}
	}
	return bin, errors
}
func (c *Compiler) next() token {
	token := c.tokens[c.pos]
	c.pos++
	return token
}
func (c *Compiler) compileLine() error {
	n := c.next()
	if n.typ != lineStart {
		return compileErr(n, n.typ.String(), lineStart.String())
	}
	lvalue := c.next()
	switch lvalue.typ {
	case eof:
		return nil
	case element:
		if err := c.compileElement(lvalue); err != nil {
			return err
		}
	case labelDef:
		c.compileLabel()
	case lineEnd:
		return nil
	default:
		return compileErr(lvalue, lvalue.text, fmt.Sprintf("%v or %v", labelDef, element))
	}
	if n := c.next(); n.typ != lineEnd {
		return compileErr(n, n.text, lineEnd.String())
	}
	return nil
}
func (c *Compiler) compileNumber(element token) (int, error) {
	num := math.MustParseBig256(element.text).Bytes()
	if len(num) == 0 {
		num = []byte{0}
	}
	c.pushBin(num)
	return len(num), nil
}
func (c *Compiler) compileElement(element token) error {
	if isJump(element.text) {
		rvalue := c.next()
		switch rvalue.typ {
		case number:
			c.compileNumber(rvalue)
		case stringValue:
			c.pushBin(rvalue.text[1 : len(rvalue.text)-2])
		case label:
			c.pushBin(vm.PUSH4)
			pos := big.NewInt(int64(c.labels[rvalue.text])).Bytes()
			pos = append(make([]byte, 4-len(pos)), pos...)
			c.pushBin(pos)
		case lineEnd:
			c.pos--
		default:
			return compileErr(rvalue, rvalue.text, "number, string or label")
		}
		c.pushBin(toBinary(element.text))
		return nil
	} else if isPush(element.text) {
		var value []byte
		rvalue := c.next()
		switch rvalue.typ {
		case number:
			value = math.MustParseBig256(rvalue.text).Bytes()
			if len(value) == 0 {
				value = []byte{0}
			}
		case stringValue:
			value = []byte(rvalue.text[1 : len(rvalue.text)-1])
		case label:
			value = big.NewInt(int64(c.labels[rvalue.text])).Bytes()
			value = append(make([]byte, 4-len(value)), value...)
		default:
			return compileErr(rvalue, rvalue.text, "number, string or label")
		}
		if len(value) > 32 {
			return fmt.Errorf("%d type error: unsupported string or number with size > 32", rvalue.lineno)
		}
		c.pushBin(vm.OpCode(int(vm.PUSH1) - 1 + len(value)))
		c.pushBin(value)
	} else {
		c.pushBin(toBinary(element.text))
	}
	return nil
}
func (c *Compiler) compileLabel() {
	c.pushBin(vm.JUMPDEST)
}
func (c *Compiler) pushBin(v interface{}) {
	if c.debug {
		fmt.Printf("%d: %v\n", len(c.binary), v)
	}
	c.binary = append(c.binary, v)
}
func isPush(op string) bool {
	return strings.ToUpper(op) == "PUSH"
}
func isJump(op string) bool {
	return strings.ToUpper(op) == "JUMPI" || strings.ToUpper(op) == "JUMP"
}
func toBinary(text string) vm.OpCode {
	return vm.StringToOp(strings.ToUpper(text))
}
type compileError struct {
	got  string
	want string
	lineno int
}
func (err compileError) Error() string {
	return fmt.Sprintf("%d syntax error: unexpected %v, expected %v", err.lineno, err.got, err.want)
}
func compileErr(c token, got, want string) error {
	return compileError{
		got:    got,
		want:   want,
		lineno: c.lineno,
	}
}

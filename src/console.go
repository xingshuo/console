package console

/*
#include <termios.h>
struct termios orginal_settings;

void console_init()
{
	struct termios new_settings;
	tcgetattr(0, &orginal_settings);
	new_settings = orginal_settings;
	new_settings.c_lflag &= (~ICANON);
	new_settings.c_lflag &= (~ECHO);
	new_settings.c_lflag &=~ISIG;
	tcsetattr(0,TCSANOW,&new_settings);
}

void console_fini()
{
	tcsetattr(0,TCSANOW,&orginal_settings);
}
*/
import "C"
import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	MAX_CONSOLE_CMD_HISTORY = 32
	MAX_CONSOLE_CMD_LEN     = 256 * 128

	KEY_BACKSPACE = 0x80
	KEY_DEL       = 0x7f
	KEY_ENTER     = 0x0A
	KEY_TAB       = 0x09
)

var (
	reservedHookKeys = map[byte]string{
		KEY_BACKSPACE: "KEY_BACKSPACE",
		KEY_DEL:       "KEY_DEL",
		KEY_ENTER:     "KEY_ENTER",
	}
)

type WaitKeyState int

const (
	WKS_WAIT WaitKeyState = iota
	WKS_RECV1B
	WKS_UDLR
)

type ParseCmdFunc func(*Console, string)
type KeyDownFunc func(*Console, string)

type Console struct {
	wks           WaitKeyState
	historyCmds   []string
	historyCursor int
	inputCmd      []byte
	inputCursor   int
	onParseCmd    ParseCmdFunc
	keyDownHooks  map[byte]KeyDownFunc
	exit          bool
}

func (co *Console) Init(pstParseFunc ParseCmdFunc) {
	co.onParseCmd = pstParseFunc
}

func (co *Console) SetKeyDownHook(key byte, hookfn KeyDownFunc) {
	if desc, ok := reservedHookKeys[key]; ok {
		fmt.Printf("%v\n", fmt.Errorf("Invalid SetHook On Reserved Key %s", desc))
		return
	}
	co.keyDownHooks[key] = hookfn
}

func (co *Console) Fini() {
	co.exit = true
}

func (co *Console) LoopCmd() {
	C.console_init()
	fmt.Print("\nCMD>")
	reader := bufio.NewReader(os.Stdin)
	for {
		if co.exit {
			break
		}
		c, err := reader.ReadByte()
		if err != nil {
			fmt.Printf("read stdin err: %v", err)
			break
		}
		switch co.wks {
		case WKS_WAIT:
			if strconv.IsPrint(rune(c)) {
				co.OnPrintChar(c)
			} else {
				if c == KEY_BACKSPACE || c == KEY_DEL {
					co.OnBackspace()
				} else if c == KEY_ENTER {
					co.OnEnter()
				} else if c == '\x1b' {
					co.wks = WKS_RECV1B
				} else if hookf, ok := co.keyDownHooks[c]; ok {
					hookf(co, string(co.inputCmd))
				}
			}
		case WKS_RECV1B:
			if c == '\x1b' {
				co.wks = WKS_RECV1B
			} else if c == '[' {
				co.wks = WKS_UDLR
			} else {
				if strconv.IsPrint(rune(c)) {
					co.OnPrintChar(c)
				}
				co.wks = WKS_WAIT
			}
		case WKS_UDLR:
			if c == 'A' {
				co.OnUp()
			} else if c == 'B' {
				co.OnDown()
			} else if c == 'D' {
				co.OnLeft()
			} else if c == 'C' {
				co.OnRight()
			} else if strconv.IsPrint(rune(c)) {
				co.OnPrintChar(c)
			}
			co.wks = WKS_WAIT
		default:
			break
		}
	}
	C.console_fini()
}

func (co *Console) ClearInput() {
	co.inputCmd = co.inputCmd[:0]
	co.inputCursor = 0
}

func (co *Console) OnPrintChar(c byte) error {
	if len(co.inputCmd) >= MAX_CONSOLE_CMD_LEN {
		return errors.New("cmd len overflow")
	}
	if co.inputCursor < len(co.inputCmd) {
		co.inputCmd = append(co.inputCmd[:co.inputCursor+1], co.inputCmd[co.inputCursor:]...)
		co.inputCmd[co.inputCursor] = c
		fmt.Printf("%s", string(co.inputCmd[co.inputCursor:]))
		bc := len(co.inputCmd) - co.inputCursor - 1
		bs := strings.Repeat("\b", bc)
		fmt.Print(bs)
	} else {
		co.inputCmd = append(co.inputCmd, c)
		fmt.Print(string(c))
	}
	co.inputCursor++
	return nil
}

func (co *Console) OnPrintString(s string) int {
	n := 0
	bs := []byte(s)
	for _, c := range bs {
		err := co.OnPrintChar(c)
		if err != nil {
			break
		}
		n++
	}
	return n
}

func (co *Console) OnBackspace() {
	if co.inputCursor > 0 {
		co.inputCmd = append(co.inputCmd[:co.inputCursor-1], co.inputCmd[co.inputCursor:]...)
		fmt.Printf("\b%s ", string(co.inputCmd[co.inputCursor-1:]))
		bc := len(co.inputCmd) - (co.inputCursor - 1)
		bs := strings.Repeat("\b", bc+1)
		fmt.Print(bs)
		co.inputCursor--
	}
}

func (co *Console) OnEnter() {
	if co.onParseCmd != nil {
		co.onParseCmd(co, string(co.inputCmd))
	}
	co.historyCmds = append(co.historyCmds, string(co.inputCmd))
	if len(co.historyCmds) > MAX_CONSOLE_CMD_HISTORY {
		co.historyCmds = append(co.historyCmds[:0], co.historyCmds[1:]...)
	}
	co.historyCursor = len(co.historyCmds)
	co.ClearInput()
	fmt.Print("\nCMD>")
}

func (co *Console) OnEsc() {
	if len(co.inputCmd) > 0 {
		if co.inputCursor < len(co.inputCmd) {
			fmt.Printf("%s", string(co.inputCmd[co.inputCursor:]))
			co.inputCursor = len(co.inputCmd)
		}
		for i := 0; i < len(co.inputCmd); i++ {
			fmt.Print("\b \b")
		}
		co.ClearInput()
	}
}

func (co *Console) OnUp() {
	if co.historyCursor <= 0 {
		return
	}
	co.historyCursor--
	co.OnEsc()
	co.inputCmd = []byte(co.historyCmds[co.historyCursor])
	co.inputCursor = len(co.inputCmd)
	fmt.Print(string(co.inputCmd))
}

func (co *Console) OnDown() {
	if co.historyCursor >= len(co.historyCmds)-1 {
		return
	}
	co.historyCursor++
	co.OnEsc()
	co.inputCmd = []byte(co.historyCmds[co.historyCursor])
	co.inputCursor = len(co.inputCmd)
	fmt.Print(string(co.inputCmd))
}

func (co *Console) OnLeft() {
	if co.inputCursor > 0 {
		fmt.Print("\b")
		co.inputCursor--
	}
}

func (co *Console) OnRight() {
	if co.inputCursor < len(co.inputCmd) {
		fmt.Print(string(co.inputCmd[co.inputCursor]))
		co.inputCursor++
	}
}

func NewConsole() *Console {
	return &Console{
		wks:           WKS_WAIT,
		historyCmds:   make([]string, 0),
		inputCmd:      make([]byte, 0),
		historyCursor: 0,
		inputCursor:   0,
		onParseCmd:    nil,
		keyDownHooks:  make(map[byte]KeyDownFunc),
		exit:          false,
	}
}

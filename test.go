package main

import (
	"fmt"
	"github.com/xingshuo/console/src"
	"strings"
)

var words = []string{"abhor", "abrasive", "absolution", "accolade", "acuity", "adamant"}

func onParseCmd(pCo *console.Console, cmd string) {
	if cmd == "quit" || cmd == "exit" {
		pCo.Fini()
	}
}

func onTabKeyDown(pCo *console.Console, cmd string) {
	var match []string
	for _, w := range words {
		if strings.HasPrefix(w, cmd) {
			match = append(match, w)
		}
	}
	if len(match) == 1 { //complete
		pCo.OnEsc()
		pCo.OnPrintString(match[0])
	} else if len(match) > 1 { //show all match string
		fmt.Printf("\n%s\nCMD>", strings.Join(match, " "))
		pCo.ClearInput()
		pCo.OnPrintString(cmd)
	}
}

func main() {
	var pCo = console.NewConsole()
	fmt.Print("--console start--")
	pCo.Init(onParseCmd)
	pCo.SetKeyDownHook(console.KEY_TAB, onTabKeyDown)
	pCo.LoopCmd()
	fmt.Print("\n--console end--\n")
}

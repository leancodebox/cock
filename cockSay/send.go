package cockSay

import (
	"fmt"
	"github.com/gen2brain/beeep"
)

func Send(msg string) {
	err := beeep.Notify("Cock", msg, "")
	if err != nil {
		fmt.Println(err)
	}
}

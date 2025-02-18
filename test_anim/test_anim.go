package main

import (
	"fmt"
	"time"
)

func main() {
	looking_anim := []string{
		" ʕ ´•ᴥ•ʔ ",
		" ʕ´•ᴥ•`ʔ ",
		" ʕ•ᴥ•` ʔ ",
		" ʕ´•ᴥ•`ʔ ",
		// "ʕꈍᴥꈍʔ",
		// "ʕ –ᴥ– ʔ",
		" ʕ´•ᴥ•`ʔก",
		"กʕ´•ᴥ•`ʔ ",
		" ʕ´•ᴥ•`ʔก",
		"กʕ´•ᴥ•`ʔ ",
		"୧ʕ´•ᴥ•`ʔ୨",
	}

	bear_frames := looking_anim

	for {
		for _, frame := range bear_frames {
			fmt.Print("\r" + frame)
			time.Sleep(1000 * time.Millisecond)
		}
	}
}

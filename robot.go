package main

import (
	"fmt"
	"github.com/go-vgo/robotgo"
	"github.com/lxn/win"
	"github.com/micmonay/keybd_event"
	"runtime"
	"time"
)

func keybd() {

	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		panic(err)
	}

	// For linux, it is very important to wait 2 seconds
	if runtime.GOOS == "linux" {
		time.Sleep(2 * time.Second)
	}

	// Select keys to be pressed
	kb.SetKeys(keybd_event.VK_DOWN)

	// Set shift to be pressed
	//kb.HasSHIFT(true)

	// Press the selected keys
	err = kb.Launching()
	if err != nil {
		panic(err)
	}

	// Or you can use Press and Release
	kb.Press()
	time.Sleep(10 * time.Millisecond)
	kb.Release()

	// Here, the program will generate "ABAB" as if they were pressed on the keyboard.
}

func robot() {
	//user32.ShowMessage()

	robotgo.SetKeyDelay(50)

	fpids, err := robotgo.FindIds("EliteDangerous64")
	if err == nil {
		fmt.Println("pids... ", fpids)

		for _, fpid := range fpids {
			if fpid > 0 {
				robotgo.ActivePID(int32(fpid))
				//win.SetForegroundWindow(win.HWND(robotgo.GetHandle()))
				robotgo.Sleep(1)
				//robotgo.KeyTap("n", "control")

				var rect win.RECT
				win.GetWindowRect(win.HWND(robotgo.GetHandle()), &rect)

				fmt.Println(fpid, rect)

				//robotgo.DragMouse(int(rect.Left) + (int(rect.Right) - int(rect.Left)/2), int(rect.Bottom) - 100)
				//robotgo.Sleep(2)
				//robotgo.DragMouse(int(rect.Right) - int(rect.Left), int(rect.Bottom) - 100)
				//robotgo.MouseClick("left")

				err := robotgo.KeyTap("1")
				if err != "" {
					fmt.Println("robotgo.KeyTap run error is: ", err)
				}
				robotgo.Sleep(2)
				robotgo.KeyTap("esc")
				robotgo.Sleep(2)
				//keybd()
				//keybd()
				//robotgo.Sleep(2)
				//robotgo.KeyToggle("down", "down")
				//robotgo.MilliSleep(10)
				//robotgo.KeyToggle("down", "up")
				//err =robotgo.KeyTap("down")
				//if err != "" {
				//	fmt.Println("robotgo.KeyTap run error is: ", err)
				//}
				robotgo.KeyTap("s")
				robotgo.KeyTap("s")
				robotgo.KeyTap("down")
				robotgo.KeyTap("down")
				//robotgo.KeyTap("down")
				//robotgo.KeyTap("down")
				//robotgo.KeyTap("down")
				//robotgo.KeyTap("down")
				//robotgo.KeyTap("down")
				//robotgo.Sleep(2)
				//robotgo.KeyTap("enter")
				//robotgo.Sleep(5)
				//robotgo.Sleep(2)

				//robotgo.KeyTap("enter")

				//robotgo.TypeStr("Hello Vassily!")
			}
		}
	}
}

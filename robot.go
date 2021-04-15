package main

import (
	"fmt"
	"github.com/go-vgo/robotgo"
	"github.com/lxn/win"
)

func robot() {
	//user32.ShowMessage()

	fpid, err := robotgo.FindIds("Notepad++")
	if err == nil {
		fmt.Println("pids... ", fpid)

		if len(fpid) > 0 {
			robotgo.ActivePID(fpid[0])
			robotgo.Sleep(1)
			robotgo.KeyTap("n", "control")

			var rect win.RECT
			win.GetWindowRect(win.HWND(robotgo.GetHandle()), &rect)

			fmt.Println(rect)

			robotgo.MoveMouse(int(rect.Left) + 100, int(rect.Top) + 100)
			robotgo.MouseClick("left")

			robotgo.TypeStr("Hello Vassily!")
		}
	}
}

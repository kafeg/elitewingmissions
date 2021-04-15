package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

var (
	Log      *log.Logger
)

func initLogs() {
	// set location of log file
	var logpath = "_goedlaunch.log"

	flag.Parse()
	var file, err1 = os.Create(logpath)

	if err1 != nil {
		panic(err1)
	}
	Log = log.New(file, "", log.LstdFlags|log.Lshortfile)
	Log.Println("LogFile : " + logpath)
}

func checkError(err error) {
	if err != nil {
		Log.Fatalf("Error: %s", err)
	}
}


func sometrash() {
	initLogs()

	//recalcAllMissions()

	//robot()

	//argsWithProg := os.Args
	argsWithoutProg := os.Args[1:]

	//d1 := []byte(strings.Join(argsWithoutProg, ", "))
	//ioutil.WriteFile("args.txt", d1, 0644)

	//arg := os.Args[3]
	//fmt.Println(argsWithProg)
	fmt.Println(argsWithoutProg)
	//fmt.Println(arg)

	Log.Printf("Input args: %s", strings.Join(argsWithoutProg, " "))

	command := exec.Command("MinEdLauncher.Bootstrap.exe", strings.Join(argsWithoutProg, " "))
	os.Setenv("USER", "filinka")
	command.Env = os.Environ()
	command.Env = append(command.Env, "USER=filinka")

	// Create stdout, stderr streams of type io.Reader
	stdout, err := command.StdoutPipe()
	checkError(err)
	stderr, err := command.StderrPipe()
	checkError(err)
	err = command.Start()
	checkError(err)

	// Non-blockingly echo command output to terminal
	//go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)

	buff := bufio.NewScanner(stdout)

	for buff.Scan() {
		Log.Printf("APP: " + buff.Text()+"\n")
	}

	command.Wait()
	if err != nil {
		Log.Printf("Command finished with error: %v", err)
	}


}
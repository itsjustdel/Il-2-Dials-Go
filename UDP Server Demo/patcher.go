package main

import "fmt"

func patcher() {

	pid, e := getProcessID("pservice.exe")
	if e != nil {
		panic(e)
	}

	fmt.Println("PID: ", pid)

	// if (hProcessIL2 == 0)
	// 	return false;

	getModule(pid)
}

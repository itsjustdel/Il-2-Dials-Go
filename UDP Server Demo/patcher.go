package main

import (
	"github.com/TheTitanrain/w32"
)

const PROCESS_ALL_ACCESS = 0x1F0FFF

func patcher() {

	pid, e := getProcessID("Il-2.exe")
	if e != nil {
		return
	}

	rseModule := getModule(pid, "RSE.dll")
	//not dry
	snapshot := w32.CreateToolhelp32Snapshot(w32.TH32CS_SNAPMODULE, pid)
	defer w32.CloseHandle(snapshot)

	//test(pid)
	functionInDLL(pid, rseModule)

}

package main

const PROCESS_ALL_ACCESS = 0x1F0FFF

func patcher() {

	pid, e := getProcessID("Il-2.exe")
	if e != nil {
		return
	}

	rseModule := getModule(pid, "RSE.dll")
}

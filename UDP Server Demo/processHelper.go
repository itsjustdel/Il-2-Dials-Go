package main

import (
	"fmt"
	//	"log"

	"github.com/TheTitanrain/w32"
	"golang.org/x/sys/windows"
)

const processEntrySize = 568
const moduleEntrySize = 1080

func getProcessID(name string) (uint32, error) {
	h, e := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if e != nil {
		return 0, e
	}
	defer windows.CloseHandle(h)

	p := windows.ProcessEntry32{Size: processEntrySize}
	for {
		e := windows.Process32Next(h, &p)
		if e != nil {
			return 0, e
		}
		if windows.UTF16ToString(p.ExeFile[:]) == name {
			return p.ProcessID, nil
		}

	}
	return 0, fmt.Errorf("%q not found", name)
}

func getModule(pid uint32, targetModuleName string) w32.MODULEENTRY32 {

	snapshot := w32.CreateToolhelp32Snapshot(w32.TH32CS_SNAPMODULE, pid)
	defer w32.CloseHandle(snapshot)

	me := w32.MODULEENTRY32{Size: moduleEntrySize}

	if w32.Module32First(snapshot, &me) {
		for {

			if w32.Module32Next(snapshot, &me) {
				//array needs converted from uint16 to string so we can read it
				s := windows.UTF16ToString(me.SzModule[:])
				if s == targetModuleName {
					fmt.Println("found ", s)
				}
			} else {
				break
			}
		}
	}

	return me
}

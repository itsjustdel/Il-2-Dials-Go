package main

import (
	"fmt"

	"golang.org/x/sys/windows"

	"github.com/Andoryuuta/kiwi/w32" //has some missing types
)

// unsafe.Sizeof(windows.ProcessEntry32{})
const processEntrySize = 568

func getProcessID(name string) (uint32, error) {
	h, e := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if e != nil {
		return 0, e
	}
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

func getModule(pid uint32) w32.MODULEENTRY32 {

	//Get handle by OpenProcess

	hProcessIL2 := OpenProcess(PROCESS_ALL_ACCESS, false, pid) //PROCESS_ALL_ACCESS needed to create code cave  //HERE

	var modEntry w32.MODULEENTRY32

	h, e := w32.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)

	if e == false {
		return modEntry
	}

	if int(h) != w32.INVALID_HANDLE_VALUE {

		var temp w32.MODULEENTRY32

		w32.Module32First(h, &temp)

		//place first module in to temp (if successful)
		if w32.Module32First(h, &temp) {
			for {

				if w32.Module32Next(h, &temp) {
					//print
					fmt.Println((h))
				} else {
					break
				}

			}
		}
		modEntry = temp
	}

	return modEntry
}

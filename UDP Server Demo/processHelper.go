package main

import (
	"fmt"
	"unsafe"

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

//code cave test
func test(pid uint32) {
	kernel32DLL := windows.NewLazySystemDLL("kernel32.dll")

	WriteProcessMemory := kernel32DLL.NewProc("WriteProcessMemory")
	VirtualAllocEx := kernel32DLL.NewProc("VirtualAllocEx")
	VirtualProtectEx := kernel32DLL.NewProc("VirtualProtectEx")
	// Get a handle on remote process
	pHandle, errProc := windows.OpenProcess(windows.PROCESS_CREATE_THREAD|windows.PROCESS_VM_OPERATION|windows.PROCESS_VM_WRITE|windows.PROCESS_VM_READ|windows.PROCESS_QUERY_INFORMATION, false, pid)
	if errProc != nil {
		fmt.Println("Open Process Error")
	}

	buf := []byte("Code Cave is here!")

	// Get a pointer to the cave of code carved out in the remote process
	pRemoteCode, _, errVirtualAlloc := VirtualAllocEx.Call(uintptr(pHandle), 0, uintptr(len(buf)), windows.MEM_COMMIT|windows.MEM_RESERVE, windows.PAGE_EXECUTE_READWRITE)
	if errVirtualAlloc != nil && errVirtualAlloc.Error() != "The operation completed successfully." {
		fmt.Println("Virtual Alloc Error")
	}

	// Write the payload into the code cave
	_, _, errWriteProcessMemory := WriteProcessMemory.Call(uintptr(pHandle), pRemoteCode, (uintptr)(unsafe.Pointer(&buf[0])), uintptr(len(buf)))

	if errWriteProcessMemory != nil && errWriteProcessMemory.Error() != "The operation completed successfully." {
		fmt.Println("Write Process Error")
	}

	oldProtect := windows.PAGE_READWRITE
	_, _, errVirtualProtectEx := VirtualProtectEx.Call(uintptr(pHandle), pRemoteCode, uintptr(len(buf)), windows.PAGE_EXECUTE_READ, uintptr(unsafe.Pointer(&oldProtect)))
	if errVirtualProtectEx != nil && errVirtualProtectEx.Error() != "The operation completed successfully." {
		fmt.Println("Virtual Protect Error")
	}

	errCloseHandle := windows.CloseHandle(pHandle)
	if errCloseHandle != nil {
		fmt.Println("Close Handle Error")

	}

}

package main

import (
	"fmt"
	"reflect"
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
//reference
//https://github.com/itsjustdel/GodeInjection/blob/main/remoteInject.go#L60
func test(pid uint32) {

	//build tools from windows dlls and processes
	kernel32DLL := windows.NewLazySystemDLL("kernel32.dll")
	WriteProcessMemory := kernel32DLL.NewProc("WriteProcessMemory")
	VirtualAllocEx := kernel32DLL.NewProc("VirtualAllocEx")
	VirtualProtectEx := kernel32DLL.NewProc("VirtualProtectEx")

	// Get a handle on remote process
	pHandle := getProcessHandleExternal(pid)

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

func getProcessHandleExternal(pid uint32) windows.Handle {
	// Get a handle on remote process
	pHandle, errProc := windows.OpenProcess(windows.PROCESS_CREATE_THREAD|windows.PROCESS_VM_OPERATION|windows.PROCESS_VM_WRITE|windows.PROCESS_VM_READ|windows.PROCESS_QUERY_INFORMATION, false, pid)
	if errProc != nil {
		fmt.Println("Open Process Error")
	}
	return pHandle

}

// typedef uint32_t DWORD;   // DWORD = unsigned 32 bit value
// typedef uint16_t WORD;    // WORD = unsigned 16 bit value
// typedef uint8_t BYTE;

type IMAGE_DOS_HEADER struct {
	E_magic    uint16     // Magic number
	E_cblp     uint16     // Bytes on last page of file
	E_cp       uint16     // Pages in file
	E_crlc     uint16     // Relocations
	E_cparhdr  uint16     // Size of header in paragraphs
	E_minalloc uint16     // Minimum extra paragraphs needed
	E_maxalloc uint16     // Maximum extra paragraphs needed
	E_ss       uint16     // Initial (relative) SS value
	E_sp       uint16     // Initial SP value
	E_csum     uint16     // Checksum
	E_ip       uint16     // Initial IP value
	E_cs       uint16     // Initial (relative) CS value
	E_lfarlc   uint16     // File address of relocation table
	E_ovno     uint16     // Overlay number
	E_res      [4]uint16  // Reserved uint16_ts
	E_oemid    uint16     // OEM identifier (for e_oeminfo)
	E_oeminfo  uint16     // OEM information; e_oemid specific
	E_res2     [10]uint16 // Reserved words
	E_lfanew   uint32     // File address of new exe header //LONG
}

func functionInDLL(pid uint32, targetModule w32.MODULEENTRY32) {

	var dosHeader IMAGE_DOS_HEADER
	dosHeader.E_magic = 65432

	var remoteModuleBaseVA = targetModule.HModule

	pHandle := getProcessHandleExternal(pid)

	buffer, _, ok := w32.ReadProcessMemory(w32.HANDLE(pHandle), uintptr(remoteModuleBaseVA), uintptr(64))
	if !ok {
		fmt.Println("Read Error")
	}

	//use reflection tp populate struct from buffer
	//need to change value stored inside address
	//go.dev/blog/laws-of-refection
	s := reflect.ValueOf(&dosHeader).Elem()
	//typeOfT := s.Type()
	for i := 0; i < s.NumField(); i++ {
		field := s.Field(i)

		switch field.Kind() {
		case reflect.Slice:
			fmt.Println("is a slice with element type")
		case reflect.Array:
			fmt.Println("is an array with element type")
		default:
			s.Field(i).SetUint(uint64(buffer[i]))
		}

		// fmt.Printf("%d: %s %s = %v\n", i,
		// 	typeOfT.Field(i).Name, field.Type(), field.Interface())
	}

	fmt.Println(dosHeader)

}

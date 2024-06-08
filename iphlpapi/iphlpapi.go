//go:build generate || windows

package iphlpapi

import (
	"fmt"
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

const (
	tcpTableBasicListener int32 = iota
	tcpTableBasicConnections
	tcpTableBasicAll
	tcpTableOwnerPidListener
	tcpTableOwnerPidConnections
	tcpTableOwnerPidAll
	tcpTableOwnerModuleListener
	tcpTableOwnerModuleConnections
	tcpTableOwnerModuleAll
)

type netConnectionKindType struct {
	family   uint32
	sockType uint32
	filename string
}

var kindTCP4 = netConnectionKindType{
	family:   syscall.AF_INET,
	sockType: syscall.SOCK_STREAM,
	filename: "tcp",
}

type mibTCPRowOwnerPid struct {
	DwState      uint32
	DwLocalAddr  uint32
	DwLocalPort  uint32
	DwRemoteAddr uint32
	DwRemotePort uint32
	DwOwningPid  uint32
}

type mibTCPTableOwnerPid struct {
	DwNumEntries uint32
	Table        [1]mibTCPRowOwnerPid
}

type pmibTCPTableOwnerPidAll *mibTCPTableOwnerPid

func GetExtendedTcpTable() error {

	var (
		pTcpTable    uintptr
		size         uintptr
		buf          []byte
		pmibTCPTable pmibTCPTableOwnerPidAll
	)

	for {
		if len(buf) > 0 {
			pmibTCPTable = (*mibTCPTableOwnerPid)((unsafe.Pointer(&buf[0])))
			pTcpTable = uintptr(unsafe.Pointer(pmibTCPTable))
		} else {
			pTcpTable = uintptr(unsafe.Pointer(pmibTCPTable))
		}
		err := getExtendedTcpTable(pTcpTable, &size, true, kindTCP4.family, tcpTableOwnerPidAll, 0)
		if err == nil {
			break
		}

		if err != windows.ERROR_INSUFFICIENT_BUFFER {
			return err
		}
		buf = make([]byte, size)
	}
	for i := uint32(0); i < pmibTCPTable.DwNumEntries; i++ {
		addr_0 := uintptr(unsafe.Pointer(&pmibTCPTable.Table[0]))
		next_addr_i := uintptr(i) * unsafe.Sizeof(pmibTCPTable.Table[0])

		row_i_pointer := unsafe.Pointer(addr_0 + next_addr_i)
		row_i := (*mibTCPRowOwnerPid)(row_i_pointer)
		fmt.Println(row_i)
	}
	return nil

}

//go:generate go run golang.org/x/sys/windows/mkwinsyscall -output ziphlpapi.go iphlpapi.go
//sys getExtendedTcpTable(pTcpTable uintptr,pdwSize *uintptr,bOrder bool,ulAf uint32,TableClass int32,Reserved uint32) (ret error) = Iphlpapi.GetExtendedTcpTable

//go:build linux || freebsd

package monero

import (
	"syscall"
)

func init() {
	// If the parent swapd or other executable that starts monero-wallet-rpc exits
	// without shutting down the child process, instruct the OS to send
	// monero-wallet-rpc SIGTERM. Unfortunately, the syscall to support this only
	// exists on Linux and FreeBSD.
	getSysProcArgs = func() *syscall.SysProcAttr {
		return &syscall.SysProcAttr{
			Pdeathsig: syscall.SIGTERM,
		}
	}
}

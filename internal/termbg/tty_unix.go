//go:build darwin || linux || freebsd || netbsd || openbsd || dragonfly

package termbg

import (
	"errors"
	"time"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

// tty is the controlling terminal opened directly, so detection works even
// when stdin is a pipe.
type tty struct {
	fd int
}

func (t tty) Write(p []byte) (int, error) {
	return unix.Write(t.fd, p)
}

// ReadTimeout waits with select, which, unlike poll and kqueue, works on
// terminal devices on macOS.
func (t tty) ReadTimeout(p []byte, d time.Duration) (int, error) {
	for {
		var fds unix.FdSet
		fds.Set(t.fd)
		tv := unix.NsecToTimeval(d.Nanoseconds())
		n, err := unix.Select(t.fd+1, &fds, nil, nil, &tv)
		if errors.Is(err, unix.EINTR) {
			continue
		}
		if err != nil || n == 0 {
			return 0, err
		}
		return unix.Read(t.fd, p)
	}
}

// Light queries the controlling terminal and reports whether its
// background is light. ok is false when there is no terminal or it did not
// answer within timeout.
func Light(timeout time.Duration) (light, ok bool) {
	fd, err := unix.Open("/dev/tty", unix.O_RDWR|unix.O_NOCTTY, 0)
	if err != nil {
		return false, false
	}
	defer unix.Close(fd)
	state, err := term.MakeRaw(fd)
	if err != nil {
		return false, false
	}
	defer term.Restore(fd, state)
	return Detect(tty{fd: fd}, timeout)
}

package termbg

import (
	"bytes"
	"testing"
	"time"
)

// scripted is a fake terminal that answers with a fixed reply, or never
// answers when reply is empty.
type scripted struct {
	reply   []byte
	written bytes.Buffer
}

func (s *scripted) Write(p []byte) (int, error) {
	return s.written.Write(p)
}

func (s *scripted) ReadTimeout(p []byte, d time.Duration) (int, error) {
	if len(s.reply) == 0 {
		time.Sleep(d)
		return 0, nil
	}
	n := copy(p, s.reply)
	s.reply = s.reply[n:]
	return n, nil
}

const da1 = "\x1b[?62;22c"

func TestDetect(t *testing.T) {
	tests := []struct {
		name      string
		reply     string
		wantLight bool
		wantOK    bool
	}{
		{"white background", "\x1b]11;rgb:ffff/ffff/ffff\x1b\\" + da1, true, true},
		{"black background", "\x1b]11;rgb:0000/0000/0000\x07" + da1, false, true},
		{"solarized light, 2-digit components", "\x1b]11;rgb:fd/f6/e3\x1b\\" + da1, true, true},
		{"DA1 only", da1, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			term := &scripted{reply: []byte(tt.reply)}
			light, ok := Detect(term, 100*time.Millisecond)
			if light != tt.wantLight || ok != tt.wantOK {
				t.Errorf("Detect = (%v, %v), want (%v, %v)", light, ok, tt.wantLight, tt.wantOK)
			}
			if term.written.String() != query {
				t.Errorf("written %q, want %q", term.written.String(), query)
			}
		})
	}
}

func TestDetectNoReply(t *testing.T) {
	start := time.Now()
	light, ok := Detect(&scripted{}, 100*time.Millisecond)
	if light || ok {
		t.Errorf("Detect = (%v, %v), want (false, false)", light, ok)
	}
	if elapsed := time.Since(start); elapsed > 150*time.Millisecond {
		t.Errorf("waited %v, want at most about 100ms", elapsed)
	}
}

func TestLuminanceThreshold(t *testing.T) {
	if Luminance(1, 1, 1) < 0.99 || Luminance(0, 0, 0) != 0 {
		t.Error("luminance of white or black is wrong")
	}
	// Mid gray is perceptually dark: relative luminance is about 0.22.
	if l := Luminance(0.5, 0.5, 0.5); l >= 0.5 {
		t.Errorf("mid gray luminance %v, want < 0.5", l)
	}
}

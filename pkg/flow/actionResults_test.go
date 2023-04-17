package flow

import "testing"

func TestTruncateLogsMsg(t *testing.T) {
	root := "root" // TODO: Linter Dupword
	lsout := "total 12\ndrwxr-x--- 3 " + root + " " + root + "4096 Apr 17 11:34 .\ndrwxrwxrwx 3" + root + " " + root + " 4096 Apr 17 11:34 ..\ndrwxr-xr-x 5 " + root + " " + root + " 4096 Apr 17 11:34 out"
	in := lsout + "\n"
	out := truncateLogsMsg(in, 1024)
	if in != out {
		t.Errorf("got '%s' want '%s'", out, in)
	}
	in = lsout
	out = truncateLogsMsg(in, 1024)
	if in != out {
		t.Errorf("got '%s' want '%s'", out, in)
	}
	in = lsout
	dr := "dr" // TODO: Linter Dupword
	want := "to\n" + dr + "\n" + dr + "\n" + dr
	out = truncateLogsMsg(in, 2)
	if want != out {
		t.Errorf("got '%s' want '%s'", out, want)
	}
	in = "test without lineending"
	want = "test without lineending"
	out = truncateLogsMsg(in, 1024)
	if want != out {
		t.Errorf("got '%s' want '%s'", out, want)
	}
}

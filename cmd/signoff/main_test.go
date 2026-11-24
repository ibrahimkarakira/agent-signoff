// Copyright 2026 Ibrahim Karakira
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"testing"
)

func TestRun(t *testing.T) {
	cases := []struct {
		name     string
		args     []string
		wantCode int
		wantOut  string
		wantErr  bool
	}{
		{"version", []string{"version"}, 0, "signoff dev\n", false},
		{"no args", nil, 2, "", true},
		{"unknown", []string{"bogus"}, 2, "", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			code := run(c.args, &out, &errOut)
			if code != c.wantCode {
				t.Errorf("exit code = %d, want %d", code, c.wantCode)
			}
			if out.String() != c.wantOut {
				t.Errorf("stdout = %q, want %q", out.String(), c.wantOut)
			}
			if (errOut.Len() > 0) != c.wantErr {
				t.Errorf("stderr = %q, want non-empty: %v", errOut.String(), c.wantErr)
			}
		})
	}
}

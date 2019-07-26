// Copyright 2017 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wagon

import (
	"bufio"
	"bytes"
	"os/exec"
	"testing"
)

func TestGovet(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := exec.Command("go", "list", "./...")
	cmd.Stdout = buf
	cmd.Stderr = buf
	err := cmd.Run()
	if err != nil {
		t.Fatalf("error getting package list: %v\n%s", err, string(buf.Bytes()))
	}
	var pkgs []string
	s := bufio.NewScanner(buf)
	for s.Scan() {
		pkgs = append(pkgs, s.Text())
	}
	if err = s.Err(); err != nil {
		t.Fatalf("error parsing package list: %v", err)
	}

	cmd = exec.Command("go", append([]string{"vet"}, pkgs...)...)
	buf = new(bytes.Buffer)
	cmd.Stdout = buf
	cmd.Stderr = buf
	err = cmd.Run()
	if err != nil {
		t.Fatalf("error running %s:\n%s\n%v", "go vet", string(buf.Bytes()), err)
	}
}

package main

import (
	"io"
	"io/ioutil"
)

func drainReader(r io.Reader) {
	if r == nil {
		return
	}
	_, _ = io.Copy(ioutil.Discard, r)
	if rc, ok := r.(io.Closer); ok {
		_ = rc.Close()
	}
}

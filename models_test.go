package models

import (
	"os"
	"testing"
)

var (
	p   *Project
	err error
)

func TestMain(m *testing.M) {
	p, err = NewProject(CYCLES)
	if err != nil {
		panic(err)
	}

	defer p.Close()
	os.Exit(m.Run())
}

func TestNote(t *testing.T) {
	t.Parallel()

	for i := 0; i < 1000; i++ {
		go p.Note(Track(i%6), C4, 127, 250)
	}

	for i := 0; i < 1000; i++ {
		go p.CC(Track(i%6), 0, 0)
	}
}

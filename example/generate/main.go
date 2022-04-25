package main

import (
	m "github.com/bh90210/models"
	"github.com/bh90210/models/generate"
)

func main() {
	p, err := m.NewProject(m.CYCLES)
	if err != nil {
		panic(err)
	}
	defer p.Close()

	s := generate.NewSong(p)
	s.Play()
}

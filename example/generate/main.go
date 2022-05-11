package main

import (
	"github.com/bh90210/models"
	"github.com/bh90210/models/generate"
)

func main() {
	p, err := models.NewProject(models.CYCLES)
	if err != nil {
		panic(err)
	}

	defer p.Close()

	s := generate.NewSong(p)
	// s.Save("./y/" + time.Now().String() + ".yaml")
	s.Play()
}

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

	// s := generate.NewSong(p)
	s, _ := generate.LoadSong(p, "/media/byron/work/models/y/2022-05-11 23:03:36.84007864 +0200 CEST m=+0.002526604.yaml")
	// s.Save("./y/" + time.Now().String() + ".yaml")
	s.Play()
}

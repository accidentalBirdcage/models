package main

import (
	"time"

	"github.com/bh90210/models"
	"github.com/bh90210/models/generate"
)

func main() {
	p, err := models.NewProject(models.CYCLES)
	if err != nil {
		panic(err)
	}

	defer p.Close()

	// s, err := generate.LoadSong(p, "/media/byron/work4/models/y/2022-06-18 22:54:01.896714561 +0200 CEST m=+0.005112534.yaml")
	// if err != nil {
	// 	panic(err)
	// }

	// ready := make(chan bool)
	// go func() {
	// 	record.Start(ready, "./out")
	// }()
	// <-ready

	s := generate.NewSong(p)
	s.Save("./y/" + time.Now().String() + ".yaml")

	err = s.Play()
	if err != nil {
		panic(err)
	}
}

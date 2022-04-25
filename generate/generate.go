package generate

import (
	"time"

	"github.com/bh90210/models"
)

type Song struct {
	project  *models.Project
	patterns map[int]pattern
	player   player
}

type pattern struct {
	len    int
	tracks map[models.Track]track
}

type track struct {
	trigs map[int]trig
}

type trig struct {
	key    models.Note
	vel    int8
	dur    float64
	nudge  float64
	preset models.Preset
}

type player struct {
	tempo int
}

func LoadSong(pathToYaml string, p *models.Project) *Song {
	return &Song{}
}

func NewSong(p *models.Project) *Song {
	return &Song{project: p}
}

func (s *Song) Play() {
	// tik := time.NewTicker(270 * time.Millisecond)
	// go func() {
	// 	for i := 0; ; i++ {
	// 		log.Println("ticker", <-tik.C)
	// 	}
	// }()

	for i := models.A4; ; i++ {
		tik := time.NewTimer(500 * time.Millisecond)
		<-tik.C
		s.project.Note(models.T1, models.A4, 127, 20)
	}
}

func (s *Song) Stop() {

}

func (s *Song) Save() {

}

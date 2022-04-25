package generate

import (
	"log"
	"time"

	"github.com/bh90210/models"
)

type Song struct {
	project  *models.Project
	patterns map[int]pattern
	player   *player
	imxport  *yamler
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
	// mu    *sync.RWMutex
	tick  *time.Ticker
	tempo chan float64
}

type yamler struct {
}

func LoadSong(pathToYaml string, p *models.Project) *Song {
	return &Song{}
}

func NewSong(project *models.Project) *Song {
	// Generate patterns and tempo.
	pat := make(map[int]pattern, 0)
	t := make(chan float64, 1)
	t <- 120.0
	return &Song{
		project:  project,
		patterns: pat,
		player: &player{
			// mu:    new(sync.RWMutex),
			tempo: t,
		},
	}
}

func (s *Song) Play() {
	s.player.play(s.project, s.patterns)
}

func (s *Song) Stop() {
	s.player.stop()
}

func (s *Song) Save() {

}

func (p *player) play(project *models.Project, patterns map[int]pattern) {
	tempo := <-p.tempo
	p.tick = time.NewTicker(time.Duration(60000/(tempo)) * time.Millisecond)
	for i := 0; ; i++ {
		select {
		case <-p.tick.C:
			log.Println("ticker")

		case tempo = <-p.tempo:
			p.tick.Reset(time.Duration(60000/(tempo)) * time.Millisecond)
		}
	}
	// tik.
	// for i := models.A4; ; i++ {
	// 	tik := time.NewTimer(500 * time.Millisecond)
	// 	<-tik.C
	// 	project.Note(models.T1, models.A4, 127, 2000)
	// }
}

func (p *player) stop() {

}

func (p *player) newTempo(t float64) {
	p.tempo <- t
}

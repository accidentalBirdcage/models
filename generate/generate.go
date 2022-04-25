package generate

import (
	"time"

	"github.com/bh90210/models"
)

// Song .
type Song struct {
	project *models.Project
	// patterns map[int]pattern
	patterns []pattern
	player   *player
	imxport  *yamler
}

type pattern struct {
	len    int
	tracks map[models.Track]*track
	tempo  float64
}

type track struct {
	trigs map[int]*trig
}

type trig struct {
	key    models.Note
	vel    int8
	dur    float64
	nudge  float64
	preset models.Preset
	tempo  float64
}

type player struct {
	tick  *time.Ticker
	tempo chan float64
}

type yamler struct {
}

// LoadSong .
func LoadSong(pathToYaml string, p *models.Project) *Song {
	return &Song{}
}

// NewSong .
func NewSong(project *models.Project) *Song {
	// Generate patterns.
	var tempo []float64 = []float64{120.0, 60.0}
	var pats []pattern
	for j := 0; j < 2; j++ {
		p := pattern{
			len:    6,
			tempo:  tempo[j],
			tracks: make(map[models.Track]*track),
		}

		for i := 0; i < 6; i++ {
			p.tracks[models.Track(i)] = &track{
				trigs: make(map[int]*trig),
			}

			p.tracks[models.Track(i)].trigs[i] = &trig{
				key: models.A4,
				vel: 127,
				dur: 250.0,
			}
		}

		pats = append(pats, p)
	}

	return &Song{
		project:  project,
		patterns: pats,
		player: &player{
			tempo: make(chan float64, 1),
		},
	}
}

// Play .
func (s *Song) Play() {
	s.player.play(s.project, s.patterns)
}

// Stop .
func (s *Song) Stop() {
	s.player.stop()
}

// Save .
func (s *Song) Save() {

}

// func (p *player) play(project *models.Project, patterns map[int]pattern) {
func (p *player) play(project *models.Project, patterns []pattern) {
	if len(patterns) != 0 {
		// Tempo is set by the tempo value of the first pattern to be played.
		tempo := patterns[0].tempo
		p.tick = time.NewTicker(time.Duration(60000/(tempo)) * time.Millisecond)
		for _, pat := range patterns {
			if tempo != pat.tempo {
				p.newTempo(pat.tempo)
			}

			for i := 0; i < pat.len; i++ {
				if i == pat.len {
					break
				}

				select {
				case <-p.tick.C:
					// log.Fatal(pat.tracks[models.T1].trigs[0].key)
					for tra, tri := range pat.tracks {
						// log.Fatal(pat.tracks[tra].trigs[0].key)
						if tri != nil {
							if tri.trigs[i] != nil {
								t := tri.trigs[i]
								// Tempo change check.
								if t.tempo != 0 {
									p.newTempo(t.tempo)
								}

								// Preset change check.
								if t.preset != nil {
									project.Preset(tra, t.preset)
								}

								// Noteon check.
								if t.key != 0 {
									project.Note(tra, t.key, t.vel, t.dur)
								}
							}
						}
					}

				case tempo = <-p.tempo:
					p.tick.Reset(time.Duration(60000/(tempo)) * time.Millisecond)
				}
			}
		}
	}

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

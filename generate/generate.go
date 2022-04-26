package generate

import (
	"time"

	"github.com/bh90210/models"
)

// Song .
type Song struct {
	project  *models.Project
	player   *player
	patterns []pattern
}

type player struct {
	tick  *time.Ticker
	tempo chan float64
}

type pattern struct {
	len    int
	tempo  float64
	tracks map[models.Track]*track
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

type generator struct {
}

type yamler struct {
}

// LoadSong .
func LoadSong(pathToYaml string, p *models.Project) *Song {
	y := new(yamler)
	y.load(pathToYaml)

	return &Song{}
}

// NewSong .
func NewSong(pro *models.Project) *Song {
	g := new(generator)

	return &Song{
		project: pro,
		player: &player{
			tempo: make(chan float64, 1),
		},
		patterns: g.generate(),
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
	y := new(yamler)
	y.save(s.patterns)
}

func (p *player) play(project *models.Project, patterns []pattern) {
	if len(patterns) != 0 {
		// Tempo is set by the tempo value of the first pattern to be played.
		tempo := patterns[0].tempo
		p.tick = time.NewTicker(time.Duration((60000/(tempo))/16) * time.Millisecond)
		for _, pat := range patterns {
			go func() {
				for {
					tempo = <-p.tempo
					p.tick.Reset(time.Duration(60000/(tempo)/16) * time.Millisecond)
				}
			}()

			if tempo != pat.tempo {
				p.newTempo(pat.tempo)
			}

			for i := 0; i < pat.len; i++ {
				if i == pat.len {
					break
				}

				<-p.tick.C
				for tra, tri := range pat.tracks {
					if tri != nil {
						if tri.trigs[i] != nil {
							// TODO: nudge
							// tik := time.NewTimer(500 * time.Millisecond)
							// <-tik.C

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
							if t.key != 0 && t.vel != 0 && t.dur != 0.0 {
								project.Note(tra, t.key, t.vel, t.dur)
							}
						}
					}
				}
			}
		}
	}
}

func (p *player) stop() {

}

func (p *player) newTempo(t float64) {
	p.tempo <- t
}

func (g *generator) generate() []pattern {
	var pat []pattern
	var tempo []float64 = []float64{100.0, 50.0}
	trigPos := []int{0, 16, 32, 48, 64, 80}
	for j := 0; j < 2; j++ {
		p := pattern{
			len:    6 * 16,
			tempo:  tempo[j],
			tracks: make(map[models.Track]*track),
		}

		for i := 0; i < 6; i++ {
			p.tracks[models.Track(i)] = &track{
				trigs: make(map[int]*trig),
			}

			p.tracks[models.Track(i)].trigs[trigPos[i]] = &trig{
				key: models.A4,
				vel: 127,
				dur: 25.0,
			}
		}

		pat = append(pat, p)
	}

	return pat
}

func (y *yamler) save(pat []pattern) {

}

func (y *yamler) load(filePath string) {

}

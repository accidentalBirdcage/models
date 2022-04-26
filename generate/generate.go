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
		// /16 = 64th note per beat.
		p.tick = time.NewTicker(time.Duration((60000/(tempo))/16) * time.Millisecond)
		for _, pat := range patterns {
			go func() {
				for {
					p.tick.Reset(time.Duration(60000/(<-p.tempo)/16) * time.Millisecond)
				}
			}()

			if tempo != pat.tempo {
				p.tick.Reset(time.Duration(60000/(pat.tempo)/16) * time.Millisecond)
			}

			for i := 0; i < pat.len; i++ {
				if i == pat.len {
					break
				}

				<-p.tick.C
				for k, track := range pat.tracks {
					if track != nil {
						if track.trigs[i] != nil {
							go func(k models.Track, trig *trig) {
								tik := time.NewTimer((time.Duration((60000/pat.tempo/16)+trig.nudge) * time.Millisecond))
								<-tik.C

								// Tempo change check.
								if trig.tempo != 0 {
									p.newTempo(trig.tempo)
								}

								// Preset change check.
								if trig.preset != nil {
									project.Preset(k, trig.preset)
								}

								// Noteon check.
								if trig.key != 0 && trig.vel != 0 && trig.dur != 0.0 {
									project.Note(k, trig.key, trig.vel, trig.dur)
								}
							}(k, track.trigs[i])
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
	var tempo []float64 = []float64{70.0, 70.0}
	var n []float64 = []float64{0.0, 300.0}
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

			p.tracks[models.Track(i%1)].trigs[trigPos[i]] = &trig{
				key:   models.A4,
				vel:   127,
				dur:   25.0,
				nudge: n[i%2],
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

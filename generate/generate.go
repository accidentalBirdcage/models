package generate

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/bh90210/models"
)

const (
	// 8 4/4 bars of /32 beats
	MAXPATLEN = 256
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

	tempo float64
}

type generator struct {
	songTempo float64
	bars      float64
	sections  int
	// If sections are more than 10 they are getting halfed.
	// This indicates longer songs.
	halfed bool
	// If true there is an extra 8 bar section that needs to
	// be inserted creatively somewhere based on secondary options.
	extraBar bool
	// 4/4 8 bars each.
	patterns []pattern
}

type yamler struct {
}

// LoadSong .
func LoadSong(pathToYaml string, p *models.Project) *Song {
	y := new(yamler)
	patterns := y.load(pathToYaml)

	return &Song{
		project: p,
		player: &player{
			tempo: make(chan float64, 1),
		},
		patterns: patterns,
	}
}

// NewSong .
func NewSong(p *models.Project) *Song {
	g := new(generator)
	patterns := g.generate()

	return &Song{
		project: p,
		player: &player{
			tempo: make(chan float64, 1),
		},
		patterns: patterns,
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
		// /8 = 32nd note per beat.
		p.tick = time.NewTicker(time.Duration((60000/(tempo))/8) * time.Millisecond)
		for _, pat := range patterns {
			go func() {
				for {
					// Block (<-p.tempo) until a new newTempo call happens.
					p.tick.Reset(time.Duration(60000/(<-p.tempo)/8) * time.Millisecond)
					// Here we don't set the tempo variable outside the for range patterns
					// because we expect tempo change to only come from triggers
					// thus pattern level tempo should remain the same.
				}
			}()

			// When starting a new pattern always check for tempo changes.
			if tempo != pat.tempo {
				p.tick.Reset(time.Duration(60000/(pat.tempo)/8) * time.Millisecond)
				// Non-thread safe.
				tempo = pat.tempo
			}

			for i := 0; i < MAXPATLEN; i++ {
				if i == MAXPATLEN {
					break
				}

				<-p.tick.C
				for k, track := range pat.tracks {
					if track != nil {
						if track.trigs[i] != nil {
							t := time.Now()
							go func(k models.Track, trig *trig, t time.Time) {
								tik := time.NewTimer(time.Duration((60000/pat.tempo/8)+trig.nudge)*time.Millisecond - time.Since(t))
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
							}(k, track.trigs[i], t)
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
	rand.Seed(time.Now().Unix())

	// Generate tempo.
	for {
		r := rand.Intn(170)
		if r < 60 {
			continue
		}

		g.songTempo = float64(r)
		break
	}

	// Generate number of bars.
bars:
	for {
		r := rand.Intn(160)
		switch {
		case g.songTempo > 70:
			if r < 40 || r%8 != 0 {
				continue bars
			}

		default:
			if r < 24 || r%8 != 0 {
				continue bars
			}
		}

		g.bars = float64(r)
		break
	}

	// Generate song structure.
	// Each pattern is an 8 bar part.
	g.patterns = make([]pattern, int(g.bars/8))
	g.sections = len(g.patterns)
	// If total bars are more than 10 half them (/2).
	if g.sections >= 10 {
		// If total bars (g.sections) is odd then minus one digit before halfing.
		if int(g.sections)%2 == 1 {
			g.extraBar = true
			g.sections -= 1
		}
		g.halfed = true
		g.sections /= 2
	}

	// Set patterns' tempo and tracks.
	// Set first pattern's tempo to song's tempo.
	g.patterns[0].tempo = g.songTempo
	// Adjust tempo per pattern and assign track in dump random
	// way going up and down comprared to previous pattern.
	// TODO: improve on the above fact.
	for i := range g.patterns {
		switch {
		case g.songTempo < 80:
			if i != 0 {
				g.patterns[i].tempo = g.patterns[i-1].tempo * 1.005
			}

		case g.songTempo > 160:
			if i != 0 {
				g.patterns[i].tempo = g.patterns[i-1].tempo * 0.996
			}

		case g.songTempo > 140:
			if i != 0 {
				g.patterns[i].tempo = g.patterns[i-1].tempo * 0.997
			}

		case g.songTempo > 120:
			if i != 0 {
				g.patterns[i].tempo = g.patterns[i-1].tempo * 0.998
			}

		case g.songTempo > 100:
			if i != 0 {
				g.patterns[i].tempo = g.patterns[i-1].tempo * 1.001
			}

		case g.songTempo > 80:
			if i != 0 {
				g.patterns[i].tempo = g.patterns[i-1].tempo * 1.000
			}
		}

		totalTrack := rand.Intn(5)
		g.patterns[i].tracks = make(map[models.Track]*track, totalTrack)
		for ; totalTrack >= 0; totalTrack-- {
			g.patterns[i].tracks[models.Track(rand.Intn(5))] = &track{trigs: make(map[int]*trig, MAXPATLEN)}
		}
	}

	// populate trigs for each pattern/track
	for i := range g.patterns {
		for k := range g.patterns[i].tracks {
			for j := 0; j < 200; j++ {
				g.patterns[i].tracks[k].trigs[j] = &trig{key: models.C4, vel: 127, dur: 100}
			}
		}
	}

	// process each trig
	fmt.Println("duration: ", time.Duration((60.0/g.songTempo)*(g.bars*4.0))*time.Second)
	fmt.Println("tempo: ", g.patterns[0].tempo)
	fmt.Println("total bars: ", g.bars)
	fmt.Println("total patterns (8 bars): ", len(g.patterns))
	fmt.Println("total sections: ", g.sections)
	fmt.Println("halfed: ", g.halfed)
	fmt.Println("extra 8 bar: ", g.extraBar)
	fmt.Println("patterns: ", g.patterns)
	fmt.Println("tracks: ", len(g.patterns[0].tracks))
	// os.Exit(0)

	return g.patterns
}

func (y *yamler) save(pat []pattern) {

}

func (y *yamler) load(filePath string) []pattern {
	return []pattern{}
}

// 4
// a a' b a''

// 5
// a a' a'' b' a'''
// a a' b b' a'''
// a a' a'' b b'

// 6
// a a' a'' a''' b a''''
// a a' a'' b a''' a''''
// a a' b b' c c'
// a a' b b' a'' a'''
// a a' a'' b b' a'''
// a a' a'' a''' b b'

// 7
// i v c v c b c
// i v c v v c o
// v c v c b c o
// v c v c b v c
// v c v c b c o

// 8
// i v c v c b c o
// v c v c b c c o

// 9
// i v c v c b c c o
// v c v c b b c c o

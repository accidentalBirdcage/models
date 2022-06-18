package generate

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"sort"
	"time"

	"github.com/bh90210/models"
	"gopkg.in/yaml.v3"
)

const (
	// MAXPATLEN maximum pattern length = 8 4/4 bars of /32nd beats (instead of /4).
	MAXPATLEN = 256
)

// Song .
type Song struct {
	Project  *models.Project `yaml:"project"`
	player   *player
	Patterns []*Pattern `yaml:"patterns"`
}

type player struct {
	tick  *time.Ticker
	tempo chan float64
}

type Pattern struct {
	Tempo  float64  `yaml:"pattern tempo"`
	Tracks []*Track `yaml:"tracks"`
}

type Track struct {
	ID      models.Track   `yaml:"id"`
	Machine models.Machine `yaml:"machine"`
	Trigs   map[int]*Trig  `yaml:"trigs"`
}

type Trig struct {
	Key    models.Note   `yaml:"key"`
	Vel    int8          `yaml:"velocity"`
	Dur    float64       `yaml:"duration"`
	Nudge  float64       `yaml:"nudge"`
	Preset models.Preset `yaml:"preset"`

	Tempo   float64         `yaml:"trig tempo"`
	Machine *models.Machine `yaml:"machine"`
}

type generator struct {
	songTempo float64
	bars      int
	sections  int
	// If sections are more than 10 they are getting halfed.
	// This indicates longer songs.
	halfed bool
	// If true there is an extra 8 bar section that needs to
	// be inserted creatively somewhere based on secondary options.
	extraBar bool
	// 4/4 8 bars each.
	patterns []*Pattern
}

type yamler struct{}

// LoadSong .
func LoadSong(p *models.Project, pathToYaml string) (*Song, error) {
	y := new(yamler)
	s, err := y.load(pathToYaml)
	if err != nil {
		return nil, fmt.Errorf("couldn't load song from YAML file: %w", err)
	}

	s.Project = p
	s.player = &player{
		tempo: make(chan float64, 1),
	}

	return s, nil
}

// NewSong .
func NewSong(p *models.Project) *Song {
	g := new(generator)

	return &Song{
		Project: p,
		player: &player{
			tempo: make(chan float64, 1),
		},
		Patterns: g.generate(),
	}
}

// Play .
func (s *Song) Play() {
	s.player.play(s.Project, s.Patterns)
}

// Stop .
func (s *Song) Stop() {
	s.player.stop()
}

// Save .
func (s *Song) Save(path string) error {
	y := new(yamler)
	return y.save(s, path)
}

func (p *player) play(project *models.Project, patterns []*Pattern) {
	if len(patterns) != 0 {
		// Tempo is set by the tempo value of the first pattern to be played.
		tempo := patterns[0].Tempo
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
			if tempo != pat.Tempo {
				p.tick.Reset(time.Duration(60000/(pat.Tempo)/8) * time.Millisecond)
				tempo = pat.Tempo
			}

			for _, track := range pat.Tracks {
				// Set Machine for track.
				project.CC(track.ID, models.MACHINE, int8(track.Machine))
			}

			for i := 0; i < MAXPATLEN; i++ {
				if i == MAXPATLEN {
					break
				}

				<-p.tick.C
				for _, track := range pat.Tracks {
					if track.Trigs[i] != nil {
						t := time.Now()
						go func(k models.Track, trig *Trig, t time.Time) {
							tik := time.NewTimer(time.Duration((60000/pat.Tempo/8)+trig.Nudge)*time.Millisecond - time.Since(t))
							<-tik.C

							// Tempo change check.
							if trig.Tempo != 0 {
								p.newTempo(trig.Tempo)
							}

							// Machine change check.
							if trig.Machine != nil {
								project.CC(track.ID, models.MACHINE, int8(*trig.Machine))
							}

							// Preset change check.
							if trig.Preset != nil {
								project.Preset(k, trig.Preset)
							}

							// Noteon check.
							if trig.Key != 0 && trig.Vel != 0 && trig.Dur != 0.0 {
								project.Note(k, trig.Key, trig.Vel, trig.Dur)
							}
						}(track.ID, track.Trigs[i], t)
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

func (g *generator) generate() []*Pattern {
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

		g.bars = r
		break
	}

	// Generate song structure.
	// Each pattern is an 8 bar part.
	g.patterns = make([]*Pattern, int(g.bars/8))
	for i := range g.patterns {
		g.patterns[i] = new(Pattern)
	}

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
	// g.patterns[0] = new(pattern)
	g.patterns[0].Tempo = g.songTempo
	// Adjust tempo per pattern and assign track in dump random
	// way going up and down comprared to previous pattern.
	// TODO: improve on the above fact.
	for i := range g.patterns {
		switch {
		case g.songTempo < 80:
			if i != 0 {
				g.patterns[i].Tempo = g.patterns[i-1].Tempo * 1.005
			}

		case g.songTempo > 160:
			if i != 0 {
				g.patterns[i].Tempo = g.patterns[i-1].Tempo * 0.996
			}

		case g.songTempo > 140:
			if i != 0 {
				g.patterns[i].Tempo = g.patterns[i-1].Tempo * 0.997
			}

		case g.songTempo > 120:
			if i != 0 {
				g.patterns[i].Tempo = g.patterns[i-1].Tempo * 0.998
			}

		case g.songTempo > 100:
			if i != 0 {
				g.patterns[i].Tempo = g.patterns[i-1].Tempo * 1.001
			}

		case g.songTempo > 80:
			if i != 0 {
				g.patterns[i].Tempo = g.patterns[i-1].Tempo * 1.000
			}
		}

		// decide total tracks
		totalTracks := rand.Intn(5) + 1
		// decide machines for each track

		for k := 0; k < totalTracks; k++ {
			g.patterns[i].Tracks = append(g.patterns[i].Tracks, &Track{
				ID:      models.Track(k),
				Machine: models.Machine(rand.Intn(5)),
				Trigs:   make(map[int]*Trig, MAXPATLEN),
			})
		}
	}

	// populate trigs for each pattern/track
	for _, pattern := range g.patterns {
		for _, track := range pattern.Tracks {
			valMin := 0.0
			valMax := float64(MAXPATLEN)
			min := 0.0
			max := 1.0
			randN := func() float64 {
				return (float64(rand.Intn(MAXPATLEN))-valMin)/(valMax-valMin)*(max-min) + min
			}

			var ar []float64
			for i := 0; i < MAXPATLEN/8; i++ {
				ar = append(ar, randN())
			}

			sort.Float64s(ar)

			// quantize rhythm.

			// quantize harmony.

			normN := func(val float64) int {
				return int((val-min)/(max-min)*(valMax-valMin) + valMin)
			}

			for _, j := range ar {
				var n int
				for {
					n = rand.Intn(int(models.B8))
					if n >= int(models.A0) {
						break
					}
				}

				p := make(models.Preset)
				p[models.COLOR] = int8(rand.Intn(126))
				p[models.CONTOUR] = int8(rand.Intn(126))
				p[models.DECAY] = int8(rand.Intn(126))
				p[models.REVERB] = int8(rand.Intn(126))
				p[models.SWEEP] = int8(rand.Intn(126))
				p[models.SHAPE] = int8(rand.Intn(126))
				p[models.DELAY] = int8(rand.Intn(126))
				p[models.GATE] = int8(rand.Intn(126))

				track.Trigs[normN(j)] = &Trig{
					Key: models.Note(n),
					Vel: int8(rand.Intn(126)),
					Dur: float64(rand.Intn(100)),

					Preset: p,
				}

			}

			// fmt.Println("track", track, fmt.Errorf("%.4f", stat.Entropy(ar)))
		}
	}

	// process each trig
	fmt.Println("duration: ", time.Duration((60.0/g.songTempo)*(float64(g.bars)*4.0))*time.Second)
	fmt.Println("tempo: ", g.patterns[0].Tempo)
	fmt.Println("total bars: ", g.bars)
	fmt.Println("total patterns (8 bars): ", len(g.patterns))
	fmt.Println("total sections: ", g.sections)
	fmt.Println("halfed: ", g.halfed)
	fmt.Println("extra 8 bar: ", g.extraBar)
	fmt.Println("patterns length: ", len(g.patterns))
	fmt.Println("pattern 0: ", g.patterns[0])

	for k, v := range g.patterns {
		fmt.Printf("pattern: %v %v\n", k, v.Tempo)
		fmt.Printf("tracks: %v \n", len(v.Tracks))
		for j, t := range v.Tracks {
			fmt.Printf("track: %v\n", models.Track(j).String())
			fmt.Printf("machine: %v\n", t.Machine)
		}
		fmt.Println()
	}

	return g.patterns
}

func (y *yamler) save(song *Song, path string) error {
	d, err := yaml.Marshal(&song)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, d, 0664)
	if err != nil {
		return err
	}

	return nil
}

func (y *yamler) load(path string) (*Song, error) {
	song, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var s *Song
	err = yaml.Unmarshal(song, &s)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func intervallic(s []int) []int {
	var i []int
	for k := range s {
		switch {
		case k+1 >= len(s):
			return i

		case s[k] < s[k+1]:
			i = append(i, s[k+1]-s[k])

		case s[k] > s[k+1]:
			i = append(i, s[k]-s[k+1])
		}
	}
	return i
}

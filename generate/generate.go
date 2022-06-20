package generate

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"sync"
	"time"

	"github.com/bh90210/models"
	"gopkg.in/yaml.v3"
)

// Song .
type Song struct {
	Project *models.Project `yaml:"project"`
	Tracks  []*Track        `yaml:"tracks"`
	player  *player
}

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
		Tracks:  g.generate(),
		player: &player{
			tempo: make(chan float64, 1),
		},
	}
}

// Play .
func (s *Song) Play() error {
	return s.player.play(s.Project, s.Tracks)
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

type player struct {
	tick  *time.Ticker
	tempo chan float64
}

func (p *player) play(project *models.Project, tracks []*Track) error {
	if len(tracks) != 0 {
		// Tempo is set by the tempo value of the first pattern to be played.
		var tempo float64
	te:
		for _, track := range tracks {
			for o, trig := range track.Trigs {
				if o != 0 {
					break
				}

				if trig.Tempo != nil {
					tempo = *trig.Tempo
					break te
				}
			}
		}

		if tempo == 0 {
			return errors.New("no init tempo found")
		}

		var wg sync.WaitGroup

		for _, track := range tracks {
			wg.Add(1)
			track.fire = make(chan models.Value)
		}

		p.tick = time.NewTicker(time.Duration(60000/tempo/64) * time.Millisecond)
		go func() {
			for {
				select {
				case <-p.tick.C:
					for _, track := range tracks {
						track.fire <- models.TwoHundredFiftySix
					}

				case tempo = <-p.tempo:
					p.tick.Reset(time.Duration(60000/tempo/64) * time.Millisecond)
				}
			}
		}()

		for _, track := range tracks {
			go func(track *Track) {
				for i, trig := range track.Trigs {
					if trig.First != nil {
						for {
							if *trig.First == 0 {
								break
							}
							*trig.First = *trig.First - <-track.fire
						}
					}

					// Tempo change check.
					if trig.Tempo != nil {
						if *trig.Tempo != tempo {
							p.newTempo(*trig.Tempo)
						}
					}

					// Machine change check.
					if trig.Machine != nil {
						project.CC(track.ID, models.MACHINE, int8(*trig.Machine))
					}

					// Preset change check.
					if trig.Preset != nil {
						project.Preset(track.ID, trig.Preset)
					}

					// Noteon check.
					if trig.Key != 0 && trig.Vel != 0 && trig.Dur != 0 && trig.Length != 0 {
						project.Note(track.ID, trig.Key, trig.Vel, trig.Dur.Float64()*(tempo*4))
					}

					for {
						if len(track.Trigs) >= i+1+1 {
							if track.Trigs[i+1].Nudge != 0 {
								trig.Length = trig.Length + track.Trigs[i+1].Nudge
							}
						}

						if trig.Length == 0 {
							break
						}
						trig.Length = trig.Length - <-track.fire
					}
				}

				wg.Done()
			}(track)
		}

		wg.Wait()
	}

	return nil
}

func (p *player) stop() {

}

func (p *player) newTempo(t float64) {
	p.tempo <- t
}

type Track struct {
	ID      models.Track   `yaml:"id"`
	Machine models.Machine `yaml:"machine"`
	Trigs   []*Trig        `yaml:"trigs"`

	fire chan models.Value
}

type Trig struct {
	Key    models.Note   `yaml:"key"`
	Vel    int8          `yaml:"velocity"`
	Dur    models.Value  `yaml:"duration"`
	Length models.Value  `yaml:"length"`
	Preset models.Preset `yaml:"preset,omitempty"`

	Nudge   models.Value    `yaml:"nudge"`
	Tempo   *float64        `yaml:"tempo,omitempty"`
	Machine *models.Machine `yaml:"machine,omitempty"`

	First *models.Value `yaml:"first,omitempty"`
}

type generator struct {
	songTempo float64
	bars      int
	// sections  int
	// If sections are more than 10 they are getting halfed.
	// This indicates longer songs.
	// halfed bool
	// If true there is an extra 8 bar section that needs to
	// be inserted creatively somewhere based on secondary options.
	// extraBar bool
	tracks []*Track
}

func (g *generator) generate() []*Track {
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

	for i := 0; i < 6; i++ {
		t := new(Track)
		t.ID = models.Track(i)
		t.Machine = models.Machine(i)

		bang := new(Trig)
		switch i {
		case 0:
			bang.Preset = models.PT1()
		case 1:
			bang.Preset = models.PT2()
		case 2:
			bang.Preset = models.PT3()
		case 3:
			bang.Preset = models.PT4()
		case 4:
			bang.Preset = models.PT5()
		case 5:
			bang.Preset = models.PT6()
		}

		bang.Key = models.C4
		bang.Dur = models.Quarter
		bang.Length = models.Quarter
		bang.Vel = 127
		f := models.Value(i)
		bang.First = &f
		t.Trigs = append(t.Trigs, bang)

		bang = new(Trig)
		bang.Key = models.C4
		bang.Dur = models.Quarter
		bang.Length = models.Quarter
		bang.Vel = 127
		t.Trigs = append(t.Trigs, bang)

		bang = new(Trig)
		bang.Key = models.C4
		bang.Dur = models.Quarter
		bang.Length = models.Quarter
		bang.Vel = 127
		t.Trigs = append(t.Trigs, bang)

		bang = new(Trig)
		bang.Key = models.C4
		bang.Dur = models.Quarter
		bang.Length = (models.Whole*8 - (models.Quarter * 4)) - models.Value(f)
		bang.Vel = 127
		t.Trigs = append(t.Trigs, bang)

		g.tracks = append(g.tracks, t)
	}

	tempo := 60.0
	tempo2 := 180.0
	g.tracks[0].Trigs[0].Tempo = &tempo
	g.tracks[2].Trigs[0].Tempo = &tempo2

	return g.tracks
}

type yamler struct{}

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

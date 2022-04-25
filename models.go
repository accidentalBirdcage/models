package models

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/writer"
	driver "gitlab.com/gomidi/rtmididrv"
)

type Model string

// Model
const (
	CYCLES  Model = "Model:Cycles"
	SAMPLES Model = "Model:Samples"
)

type Track int8

// Tracks
const (
	T1 Track = iota
	T2
	T3
	T4
	T5
	T6
)

type Note int8

// Keys/letter Note
const (
	A0 Note = iota + 21
	As0
	B0
	C1
	Cs1
	D1
	Ds1
	E1
	F1
	Fs1
	G1
	Gs1
	A1
	As1
	B1
	C2
	Cs2
	D2
	Ds2
	E2
	F2
	Fs2
	G2
	Gs2
	A2
	As2
	B2
	C3
	Cs3
	D3
	Ds3
	E3
	F3
	Fs3
	G3
	Gs3
	A3
	As3
	B3
	C4
	Cs4
	D4
	Ds4
	E4
	F4
	Fs4
	G4
	Gs4
	A4
	As4
	B4
	C5
	Cs5
	D5
	Ds5
	E5
	F5
	Fs5
	G5
	Gs5
	A5
	As5
	B5
	C6
	Cs6
	D6
	Ds6
	E6
	F6
	Fs6
	G6
	Gs6
	A6
	As6
	B6
	C7
	Cs7
	D7
	Ds7
	E7
	F7
	Fs7
	G7
	Gs7
	A7
	As7
	B7
	C8
	Cs8
	D8
	Ds8
	E8
	F8
	Fs8
	G8
	Gs8
	A8
	As8
	B8

	Bf0 Note = As0
	Df1 Note = Cs1
	Ef1 Note = Ds1
	Gf1 Note = Fs1
	Af1 Note = Gs1
	Bf1 Note = As1
	Df2 Note = Cs2
	Ef2 Note = Ds2
	Gf2 Note = Fs2
	Af2 Note = Gs2
	Bf2 Note = As2
	Df3 Note = Cs3
	Ef3 Note = Ds3
	Gf3 Note = Fs3
	Af3 Note = Gs3
	Bf3 Note = As3
	Df4 Note = Cs4
	Ef4 Note = Ds4
	Gf4 Note = Fs4
	Af4 Note = Gs4
	Bf4 Note = As4
	Df5 Note = Cs5
	Ef5 Note = Ds5
	Gf5 Note = Fs5
	Af5 Note = Gs5
	Bf5 Note = As5
	Df6 Note = Cs6
	Ef6 Note = Ds6
	Gf6 Note = Fs6
	Af6 Note = Gs6
	Bf6 Note = As6
	Df7 Note = Cs7
	Ef7 Note = Ds7
	Gf7 Note = Fs7
	Af7 Note = Gs7
	Bf7 Note = As7
	Df8 Note = Cs8
	Ef8 Note = Ds8
	Gf8 Note = Fs8
	Af8 Note = Gs8
	Bf8 Note = As8
)

type Chords int8

// Chords
const (
	Unisonx2 Chords = iota
	Unisonx3
	Unisonx4
	Minor
	Major
	Sus2
	Sus4
	MinorMinor7
	MajorMinor7
	MinorMajor7
	MajorMajor7
	MinorMinor7Sus4
	Dim7
	MinorAdd9
	MajorAdd9
	Minor6
	Major6
	Minorb5
	Majorb5
	MinorMinor7b5
	MajorMinor7b5
	MajorAug5
	MinorMinor7Aug5
	MajorMinor7Aug5
	Minorb6
	MinorMinor9no5
	MajorMinor9no5
	MajorAdd9b5
	MajorMajor7b5
	MajorMinor7b9no5
	Sus4Aug5b9
	Sus4AddAug5
	MajorAddb5
	Major6Add4no5
	MajorMajor76no5
	MajorMajor9no5
	Fourths
	Fifths
)

type Parameter int8

const (
	// NOTE       Parameter = 3
	TRACKLEVEL Parameter = 17
	MUTE       Parameter = 94
	PAN        Parameter = 10
	SWEEP      Parameter = 18
	CONTOUR    Parameter = 19
	DELAY      Parameter = 12
	REVERB     Parameter = 13
	VOLUMEDIST Parameter = 7
	// SWING      Parameter = 15
	// CHANCE     Parameter = 14

	// model:cycles
	MACHINE     Parameter = 64
	CYCLESPITCH Parameter = 65
	DECAY       Parameter = 80
	COLOR       Parameter = 16
	SHAPE       Parameter = 17
	PUNCH       Parameter = 66
	GATE        Parameter = 67

	// model:samples
	PITCH        Parameter = 16
	SAMPLESTART  Parameter = 19
	SAMPLELENGTH Parameter = 20
	CUTOFF       Parameter = 74
	RESONANCE    Parameter = 71
	LOOP         Parameter = 17
	REVERSE      Parameter = 18
)

// Reverb & Delay settings
const (
	DELAYTIME Parameter = iota + 85
	DELAYFEEDBACK
	REVERBSIZE
	REVERBTONE
)

// LFO settings
const (
	LFOSPEED Parameter = iota + 102
	LFOMULTIPIER
	LFOFADE
	LFODEST
	LFOWAVEFORM
	LFOSTARTPHASE
	LFORESET
	LFODEPTH
)

type Machine int8

// Machines
const (
	KICK Machine = iota
	SNARE
	METAL
	PERC
	TONE
	CHORD
)

type ScaleMode bool

const (
	PTN ScaleMode = true
	TRK ScaleMode = false
)

// Project long description of the data structure, methods, behaviors and useage.
type Project struct {
	Model

	mu *sync.RWMutex
	// midi fields
	drv midi.Driver
	in  midi.In
	out midi.Out
	wr  *writer.Writer

	on map[Track]*active
}

type active struct {
	Note
	cancel chan bool
}

type Preset map[Parameter]int8

// NewProject initiates and returns a *Project struct.
func NewProject(m Model) (*Project, error) {
	drv, err := driver.New()
	if err != nil {
		return nil, err
	}

	p := &Project{
		Model: m,
		mu:    new(sync.RWMutex),
		drv:   drv,
		on:    make(map[Track]*active),
	}

	// find elektron and assign it to in/out
	var helperIn, helperOut bool

	p.mu.Lock()
	ins, _ := drv.Ins()
	for _, in := range ins {
		if strings.Contains(in.String(), string(m)) {
			p.in = in
			helperIn = true
		}
	}
	outs, _ := drv.Outs()
	for _, out := range outs {
		if strings.Contains(out.String(), string(m)) {
			p.out = out
			helperOut = true
		}
	}
	// check if nothing found
	if !helperIn && !helperOut {
		return nil, fmt.Errorf("device %s not found", m)
	}

	err = p.in.Open()
	if err != nil {
		return nil, err
	}

	err = p.out.Open()
	if err != nil {
		return nil, err
	}

	wr := writer.New(p.out)
	p.wr = wr
	p.mu.Unlock()

	return p, nil
}

// Preset immediately sets (CC) provided parameterp.
func (p *Project) Preset(track Track, preset Preset) {
	for parameter, value := range preset {
		p.cc(track, parameter, value)
	}
}

// Note fires immediately a midi note on signal followed by a note off specified duration in milliseconds (ms).
// Optionally user can pass a preset too for convenience.
func (p *Project) Note(track Track, note Note, velocity int8, duration float64, pre ...Preset) {
	p.mu.RLock()
	if p.on[track] != nil {
		p.mu.RUnlock()

		p.mu.Lock()
		p.on[track].cancel <- true
		p.noteoff(track, p.on[track].Note)
		p.on[track] = nil
		p.mu.Unlock()
	} else {
		p.mu.RUnlock()
	}

	if len(pre) != 0 {
		for i := range pre {
			p.Preset(track, pre[i])
		}
	}

	p.noteon(track, note, velocity)
	t := time.NewTimer(time.Millisecond * time.Duration(duration))
	p.mu.Lock()
	p.on[track] = &active{Note: note, cancel: make(chan bool)}
	p.mu.Unlock()
	go func() {
		select {
		case <-t.C:
			p.noteoff(track, note)

			p.mu.Lock()
			p.on[track] = nil
			p.mu.Unlock()

		case <-p.on[track].cancel:
			t.Stop()
		}
	}()
}

// CC control change.
func (p *Project) CC(track Track, parameter Parameter, value int8) {
	p.cc(track, parameter, value)
}

// PC Project control change.
func (p *Project) PC(t Track, pc int8) {
	p.pc(t, pc)
}

// Close midi connection. Use it with defer after creating a new project.
func (p *Project) Close() {
	p.in.Close()
	p.out.Close()
	p.drv.Close()
}

func (p *Project) noteon(t Track, n Note, vel int8) {
	p.wr.SetChannel(uint8(t))
	writer.NoteOn(p.wr, uint8(n), uint8(vel))
}

func (p *Project) noteoff(t Track, n Note) {
	p.wr.SetChannel(uint8(t))
	writer.NoteOff(p.wr, uint8(n))
}

func (p *Project) cc(t Track, par Parameter, val int8) {
	p.wr.SetChannel(uint8(t))
	writer.ControlChange(p.wr, uint8(par), uint8(val))
}

func (p *Project) pc(t Track, pc int8) {
	p.wr.SetChannel(uint8(t))
	writer.ProgramChange(p.wr, uint8(pc))
}

func PT1() Preset {
	p := make(map[Parameter]int8)
	p[MACHINE] = int8(KICK)
	p[TRACKLEVEL] = int8(120)
	p[MUTE] = int8(0)
	p[PAN] = int8(63)
	p[SWEEP] = int8(16)
	p[CONTOUR] = int8(24)
	p[DELAY] = int8(0)
	p[REVERB] = int8(0)
	p[VOLUMEDIST] = int8(60)
	p[CYCLESPITCH] = int8(64)
	p[DECAY] = int8(29)
	p[COLOR] = int8(10)
	p[SHAPE] = int8(16)
	p[PUNCH] = int8(0)
	p[GATE] = int8(0)
	return p
}

func PT2() Preset {
	p := PT1()
	p[MACHINE] = int8(SNARE)
	p[SWEEP] = int8(8)
	p[CONTOUR] = int8(0)
	p[DECAY] = int8(40)
	p[COLOR] = int8(0)
	p[SHAPE] = int8(127)
	return p
}

func PT3() Preset {
	p := PT1()
	p[MACHINE] = int8(METAL)
	p[SWEEP] = int8(48)
	p[CONTOUR] = int8(0)
	p[DECAY] = int8(20)
	p[COLOR] = int8(16)
	p[SHAPE] = int8(46)
	return p
}

func PT4() Preset {
	p := PT1()
	p[MACHINE] = int8(PERC)
	p[SWEEP] = int8(100)
	p[CONTOUR] = int8(64)
	p[DECAY] = int8(26)
	p[COLOR] = int8(15)
	p[SHAPE] = int8(38)
	return p
}

func PT5() Preset {
	p := PT1()
	p[MACHINE] = int8(TONE)
	p[SWEEP] = int8(38)
	p[CONTOUR] = int8(52)
	p[DECAY] = int8(42)
	p[COLOR] = int8(22)
	p[SHAPE] = int8(40)
	return p
}

func PT6() Preset {
	p := PT1()
	p[MACHINE] = int8(CHORD)
	p[SWEEP] = int8(43)
	p[CONTOUR] = int8(24)
	p[DECAY] = int8(64)
	p[COLOR] = int8(20)
	p[SHAPE] = int8(4)
	return p
}

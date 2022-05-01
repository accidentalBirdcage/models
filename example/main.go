package main

import (
	"time"

	"github.com/bh90210/models"
)

func main() {
	p, err := models.NewProject(models.CYCLES)
	if err != nil {
		panic(err)
	}

	defer p.Close()

	// Note length in ms.
	len := 250.0

	// Get the preset sound for track 1 (kick).
	kick := models.PT1()
	// Apply it to track 1 (kick).
	p.Preset(models.T1, kick)
	// Trigger a note.
	p.Note(models.T1, models.C4, 120, len)

	time.Sleep(time.Duration(len) * time.Millisecond)

	// Trigger a note with a preset sound config.
	p.Preset(models.T2, models.PT2())
	p.Note(models.T2, models.C4, 120, len)

	time.Sleep(time.Duration(len) * time.Millisecond)

	p.Preset(models.T3, models.PT3())
	// Send an individual control change message.
	p.CC(models.T3, models.DELAY, 0)
	p.Note(models.T3, models.C4, 120, len)

	time.Sleep(time.Duration(len) * time.Millisecond)

	defaultPerc := models.PT4()
	p.Preset(models.T4, defaultPerc)
	// Create a new preset from scratch.
	perc := make(map[models.Parameter]int8)
	perc[models.DELAY] = 0
	p.Preset(models.T4, perc)
	p.Note(models.T4, models.C4, 120, len)

	time.Sleep(time.Duration(len) * time.Millisecond)

	p.Preset(models.T5, models.PT5())
	p.Note(models.T5, models.C4, 120, len)

	time.Sleep(time.Duration(len) * time.Millisecond)

	chord := models.PT6()
	// Set a particular chord to be played.
	chord[models.SHAPE] = int8(models.Major)
	p.CC(models.T6, models.SHAPE, chord[models.SHAPE])
	p.Note(models.T6, models.C4, 120, len)
}

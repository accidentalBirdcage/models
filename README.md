<img src="https://user-images.githubusercontent.com/22690219/130872109-150ac61f-ad69-4bfb-8f10-3337abcb6551.png" alt="drawing" width="350"/> <img src="https://i.imgur.com/pJbgSUh.png" alt="drawing" width="350"/>

[![Go Reference](https://pkg.go.dev/badge/github.com/bh90210/models.svg)](https://pkg.go.dev/github.com/bh90210/models)

# elektron:models

Go package to programmatically control [Elektron's](https://www.elektron.se/) **model:cycles** & **model:samples** via midi & a song generator under `/generate` package.

## Prerequisites

### Go

Install Go https://golang.org/doc/install.

### RtMidi

#### Ubuntu 20.04+

```console
apt install librtmidi4 librtmidi-dev
```
For older versions take a look [here](https://launchpad.net/ubuntu/+source/rtmidi).

#### MacOS

```console
brew install rtmidi
```
For more information see the [formulae page](https://formulae.brew.sh/formula/rtmidi).

#### Windows

`Help needed.`

## Quick Use

_A complete example can be found in the [example](https://github.com/bh90210/elektronmodels/tree/master/example/) folder._

_The relevant cycles/samples manuals' part for this library is the `APPENDIX A: MIDI SPECIFICATIONS`._

<img src="https://i.imgur.com/Yrs6YS3.png" alt="drawing" width="350"/> <img src="https://i.imgur.com/cmil9NG.png" alt="drawing" width="350"/>


Code to get a single kick drum hit at C4 key, with velocity set at `120` and length at 200 milliseconds:
```go
package main

import "github.com/bh90210/models"

func main() {
	p, _ := models.NewProject(models.CYCLES)
	defer p.Close()

    // Track, note, velocity, length (ms), preset.
	p.Note(models.T1, models.C4, 120, 200, models.PT1())
}

```
There are four methods to use, `Preset` to set preset on the fly, `Note` to fire a note on/off for given duration, `CC` to send a single control change message && `PC` for program changes. 

# Generator

Generator package allows to generate songs in pseudo random fashion.

It will produce 4/4 songs between around 1m 30sec - 7m long & between 60 to 170 BPM, always in multiples of 8 bars, with a minimum of 32 bars per song.

The API is very simple but allows for the generated song to be exported as a YAML file and subsequently imported and played back.

## API

### generate.NewSong 

Generate a new song. It accepts an initiated project object.

### generate.LoadSong

Load an existing song. It accepts an initiated project object and the path to the YAML file containing the song data.

### Methods

#### Song.Play

Start playing the song. If the song is already playing unlike the machine it will not pause but just ignore it.

Play is a blocking function. It will unblock when the song is over.

#### Song.Stop

Stop and reset the position of the song.

#### Song.Save

Export generated song in YAML format.

some info

a sample

Example code:
```go
package main

import (
	"github.com/bh90210/models"
	"github.com/bh90210/models/generate"
)

func main() {
	p, _ := models.NewProject(models.CYCLES)
	defer p.Close()

	s := generate.NewSong(p)
	// Play generated song.
	s.Play()
    // Export song as YAML file.
	s.Save("/path/to/save/file.yaml")
}

```
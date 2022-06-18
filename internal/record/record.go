package record

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/jfreymuth/pulse"
)

func Start(ready chan bool, filename string) {
	c, err := pulse.NewClient()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer c.Close()

	file := newFile(fmt.Sprintf("%s.wav", filename), 44100, 2)
	stream, err := c.NewRecord(pulse.Float32Writer(file.Write), pulse.RecordStereo)
	if err != nil {
		fmt.Println(err)
		return
	}

	stream.Start()

	fmt.Println("Press enter to stop recording...")
	ready <- true
	os.Stdin.Read([]byte{0})
	fmt.Println("Stopping recording...")
	stream.Stop()
	file.Close()
}

type wavfile struct {
	out io.WriteSeeker
}

func newFile(name string, sampleRate, channels int) *wavfile {
	file, err := os.Create(name)
	if err != nil {
		panic(err)
	}

	io.WriteString(file, "RIFF    WAVEfmt ")
	var buf [20]byte
	binary.LittleEndian.PutUint32(buf[:], 16)
	binary.LittleEndian.PutUint16(buf[4:], 3)                              // format (3 = float)
	binary.LittleEndian.PutUint16(buf[6:], uint16(channels))               // channels
	binary.LittleEndian.PutUint32(buf[8:], uint32(sampleRate))             // sample rate
	binary.LittleEndian.PutUint32(buf[12:], uint32(sampleRate*4*channels)) // bytes/second
	binary.LittleEndian.PutUint16(buf[16:], uint16(4*channels))            // bytes/frame
	binary.LittleEndian.PutUint16(buf[18:], 32)                            // bits/sample
	file.Write(buf[:])
	io.WriteString(file, "data    ")
	return &wavfile{out: file}
}

func (f *wavfile) Write(p []float32) (int, error) {
	return len(p), binary.Write(f.out, binary.LittleEndian, p)
}

func (f *wavfile) Close() {
	pos, _ := f.out.Seek(0, io.SeekCurrent)
	f.out.Seek(4, io.SeekStart)
	binary.Write(f.out, binary.LittleEndian, uint32(pos-8))
	f.out.Seek(40, io.SeekStart)
	binary.Write(f.out, binary.LittleEndian, uint32(pos-44))
	if c, ok := f.out.(io.Closer); ok {
		c.Close()
	}
}

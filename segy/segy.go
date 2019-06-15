package segy

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/sanandak/seg/seg2"
)

// header info borrowed from github.com/asbjorn/segygo

// TraceHeader - segy/su 240 byte trace header
type TraceHeader struct {
	Tracl                                                                int32
	Tracr                                                                int32
	Fldr                                                                 int32
	Tracf                                                                int32
	Ep                                                                   int32
	CDP                                                                  int32
	CDPT                                                                 int32
	Trid                                                                 int16
	Nvs                                                                  int16
	Nhs                                                                  int16
	Duse                                                                 int16
	Offset                                                               int32
	Gelev                                                                int32
	Selev                                                                int32
	Sdepth                                                               int32
	Gdel                                                                 int32
	Sdel                                                                 int32
	SwDep                                                                int32
	GwDep                                                                int32
	Scalel                                                               int16
	Scalco                                                               int16
	Sx                                                                   int32
	Sy                                                                   int32
	Gx                                                                   int32
	Gy                                                                   int32
	CoUnit                                                               int16
	WeVel                                                                int16
	SweVel                                                               int16
	Sut, Gut, Sstat, Gstat, Tstat, Laga, Lagb, Delrt, Muts, Mute         int16
	Ns, Dt                                                               uint16
	Gain, Igc, Igi, Corr, Sfs, Sfe, Slen, Styp, Stas, Stae, Tatyp        int16
	Afilf, Afils, NoFilf, NoFils, Lcf, Hcf, Lcs, Hcs, Year, Day          int16
	Hour, Minute, Sec, Timbas, Trwf, Grnors, Grnofr, Grnlof, Gaps, Otrav int16
	D1, F1, D2, F2, Ungpow, Unscale                                      float32
	Ntr                                                                  int32
	Mark, Shortpad                                                       int16
	Unass                                                                [14]int16 // unassigned short array
}

// Trace is a SEGY trace with Data and TraceHeader
type Trace struct {
	TraceHeader
	Data []float32
}

// ReadSU reads from `fn` (string) and returns an array of segy []Trace
func ReadSU(fn string) []Trace {
	f, err := os.Open(fn)
	if err != nil {
		log.Fatal("can't open", fn)
	}
	defer f.Close()

	trcs := make([]Trace, 0)
	for {
		hdr := TraceHeader{}
		err := binary.Read(f, binary.BigEndian, &hdr)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("err reading from ", fn)
		}
		ns := hdr.Ns
		//dt := hdr.Dt
		data := make([]float32, ns)
		err = binary.Read(f, binary.LittleEndian, &data)
		if err != nil {
			log.Fatal("err reading data from", fn)
		}
		trcs = append(trcs, Trace{TraceHeader: hdr, Data: data})
	}
	return trcs
}

// WriteSU - write to ``fn` (string) write `trcs` (array of segy []Trace)
func WriteSU(fn string, trcs []Trace) int {
	f, err := os.OpenFile(fn, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("err opening out file", fn)
	}
	defer f.Close()

	var buf []byte
	b := bytes.NewBuffer(buf)

	var nwritten int
	for i, trc := range trcs {
		binary.Write(b, binary.BigEndian, trc.TraceHeader)
		binary.Write(b, binary.BigEndian, trc.Data)
		n, err := f.Write(b.Bytes())
		if err != nil {
			log.Fatal("err writing trace #", i)
		}
		nwritten += n
		b.Reset()
	}
	return nwritten
}

// Seg2Segy takes an individual seg2 `Trace` and returns
// a byte slice representing a SEG-Y trace including 240 byte
// header and trace data (big endian)
func Seg2Segy(seg2trc seg2.Trace) []byte {
	trc := &Trace{}
	hdr := &trc.TraceHeader
	hdr.Ns = uint16(len(seg2trc.Data))
	hdr.Dt = uint16(1000)
	if dt, ok := seg2trc.Hdr["SAMPLE_INTERVAL"]; ok {
		v, err := strconv.ParseFloat(dt[0], 32)
		if err == nil {
			hdr.Dt = uint16(v * 1.e6)
		}
	}

	if s, ok := seg2trc.Hdr["SHOT_SEQUENCE_NUMBER"]; ok {
		v, err := strconv.Atoi(s[0])
		if err == nil {
			hdr.Fldr = int32(v)
		}
	}

	if s, ok := seg2trc.Hdr["RECEIVER_LOCATION"]; ok {
		v, err := strconv.ParseFloat(s[0], 32)
		if err == nil {
			hdr.Gx = int32(v)
		}
	}

	if s, ok := seg2trc.Hdr["SOURCE_LOCATION"]; ok {
		v, err := strconv.ParseFloat(s[0], 32)
		if err == nil {
			hdr.Sx = int32(v)
		}
	}
	if s, ok := seg2trc.Hdr["CHANNEL_NUMBER"]; ok {
		v, err := strconv.Atoi(s[0])
		if err == nil {
			hdr.Tracr = int32(v)
			hdr.Tracf = int32(v)
		}
	}

	hdr.Offset = hdr.Sx - hdr.Gx
	trc.Data = seg2trc.Data

	var buf []byte
	b := bytes.NewBuffer(buf)
	binary.Write(b, binary.BigEndian, hdr)
	binary.Write(b, binary.BigEndian, trc.Data)
	return b.Bytes()
}

package seg2

import (
	"bytes"
	"encoding/binary"
	"log"
	"os"
	"strings"
)

// HDRS are the trongly suggested headers
var HDRS = []string{
	"RECEIVER_LOCATION",    // int
	"SAMPLE_INTERVAL",      // float
	"SOURCE_LOCATION",      // int
	"SHOT_SEQUENCE_NUMBER", // int
	"DELAY",                // float
}

// TrcHdr is a map of the strings in the header as key and an array of
// strings in the value
type TrcHdr map[string][]string

// Trace consists of the data and header (a map/dict)
type Trace struct {
	Data []float32
	Hdr  TrcHdr
}

// FileHeader is the contents of the SEG2 File Header
type FileHeader struct {
	FDID        uint16
	RevNum      uint16
	TrcPtrLen   uint16
	NumTrcs     uint16
	StrTermLen  uint8
	StrTerm0    byte
	StrTerm1    byte
	LineTermLen uint8
	LineTerm0   byte
	LineTerm1   byte
	Res         [18]byte
	//Strings     []byte
}

// TrcBlkHdr is the contents of the Trace Block Header
type TrcBlkHdr struct {
	TrcID      uint16
	BlkSiz     uint16
	DataBlkSiz uint32
	NSamps     uint32
	DataFormat uint8
	Res        [19]byte
}

// ReadSEG2 reads a SEG2 formatted file and returns an array of `SEG2Trace`
func ReadSEG2(fn string) []Trace {
	// seg2 file header
	fh := &FileHeader{}
	f, err := os.Open(fn)
	if err != nil {
		log.Fatal("can't open", fn)
	}
	defer f.Close()

	// read file header block
	binary.Read(f, binary.LittleEndian, fh)
	//fmt.Printf("%+v\n", fh)

	// read trace pointer block
	trcptrs := make([]uint32, fh.NumTrcs)
	binary.Read(f, binary.LittleEndian, &trcptrs)

	// read trace header block and data block
	var seg2trcs = make([]Trace, fh.NumTrcs)
	for i, ptr := range trcptrs {
		//fmt.Println("trace", i)
		f.Seek(int64(ptr), os.SEEK_SET)
		tbh := &TrcBlkHdr{}
		binary.Read(f, binary.LittleEndian, tbh)
		//fmt.Println(tbh)
		//trcStrLen := tbh.BlkSiz - 32
		//trcStr := make([]byte, trcStrLen)
		var hdr = make(TrcHdr)
		var strOffset uint16
		for {
			binary.Read(f, binary.LittleEndian, &strOffset)
			//fmt.Println(strOffset)
			if strOffset == 0 {
				break
			}
			var str = make([]byte, strOffset-2)
			binary.Read(f, binary.LittleEndian, &str)

			// assuming only one string terminating char, get rid of it
			str = bytes.Trim(str, string([]byte{fh.StrTerm0}))

			//fmt.Println(string(str))
			comps := strings.Split(string(str), " ")
			if len(comps) > 0 {
				hdr[comps[0]] = comps[1:]
			}
			//fmt.Println(hdr)
		}
		f.Seek(int64(ptr)+int64(tbh.BlkSiz), os.SEEK_SET)
		var data = make([]float32, tbh.NSamps)
		binary.Read(f, binary.LittleEndian, &data)
		//fmt.Println(hdr, data[:10])
		seg2trcs[i] = Trace{Data: data, Hdr: hdr}
	}
	return seg2trcs
}

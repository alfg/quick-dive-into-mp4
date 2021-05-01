package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

const (
	// BoxHeaderSize Size of box header.
	BoxHeaderSize = int64(8)
)

// FtypBox - File Type Box
// Box Type: ftyp
// Container: File
// Mandatory: Yes
// Quantity: Exactly one
type FtypBox struct {
	*Box
	MajorBrand       string   // Brand identifer.
	MinorVersion     uint32   // Informative integer for the minor version of the major brand.
	CompatibleBrands []string // A list, to the end of the box, of brands.
}

func (b *FtypBox) parse() error {
	data := b.ReadBoxData()
	b.MajorBrand = string(data[0:4])
	b.MinorVersion = binary.BigEndian.Uint32(data[4:8])
	if len(data) > 8 {
		for i := 8; i < len(data); i += 4 {
			b.CompatibleBrands = append(b.CompatibleBrands, string(data[i:i+4]))
		}
	}
	return nil
}

type MoovBox struct {
	*Box
	Mvhd *MvhdBox
}

func (b *MoovBox) parse() error {
	boxes := readBoxes(b.Reader, b.Start+BoxHeaderSize, b.Size-BoxHeaderSize)

	for _, box := range boxes {
		switch box.Name {
		case "mvhd":
			b.Mvhd = &MvhdBox{Box: box}
			b.Mvhd.parse()
		}
	}
	return nil
}

type MvhdBox struct {
	*Box
	Flags            uint32
	Version          uint8
	CreationTime     uint32
	ModificationTime uint32
	Timescale        uint32
	Duration         uint32
	Rate             Fixed32
	Volume           Fixed16
}

func (b *MvhdBox) parse() error {
	data := b.ReadBoxData()
	b.Version = data[0]
	b.Timescale = binary.BigEndian.Uint32(data[12:16])
	b.Duration = binary.BigEndian.Uint32(data[16:20])
	b.Rate = fixed32(data[20:24])
	b.Volume = fixed16(data[24:26])
	return nil
}

// Mp4Reader defines an mp4 reader structure.
type Mp4Reader struct {
	Reader io.ReaderAt
	Ftyp   *FtypBox
	Moov   *MoovBox
	Size   int64
}

// Parse reads an MP4 reader for atom boxes.
func (m *Mp4Reader) Parse() error {
	if m.Size == 0 {
		if ofile, ok := m.Reader.(*os.File); ok {
			info, err := ofile.Stat()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return err
			}
			m.Size = info.Size()
		}
	}

	boxes := readBoxes(m, int64(0), m.Size)
	for _, box := range boxes {
		switch box.Name {
		case "ftyp":
			m.Ftyp = &FtypBox{Box: box}
			m.Ftyp.parse()

		case "moov":
			m.Moov = &MoovBox{Box: box}
			m.Moov.parse()
			// Add cases to check for more boxes here.
		}
	}
	return nil
}

// ReadBoxAt reads a box from an offset.
func (m *Mp4Reader) ReadBoxAt(offset int64) (boxSize uint32, boxType string) {
	buf := m.ReadBytesAt(BoxHeaderSize, offset)
	boxSize = binary.BigEndian.Uint32(buf[0:4])
	boxType = string(buf[4:8])
	return boxSize, boxType
}

// ReadBytesAt reads a box at n and offset.
func (m *Mp4Reader) ReadBytesAt(n int64, offset int64) (word []byte) {
	buf := make([]byte, n)
	if _, error := m.Reader.ReadAt(buf, offset); error != nil {
		fmt.Println(error)
		return
	}
	return buf
}

// Box defines an Atom Box structure.
type Box struct {
	Name        string
	Size, Start int64
	Reader      *Mp4Reader
}

// ReadBoxData reads the box data from an atom box.
func (b *Box) ReadBoxData() []byte {
	if b.Size <= BoxHeaderSize {
		return nil
	}
	return b.Reader.ReadBytesAt(b.Size-BoxHeaderSize, b.Start+BoxHeaderSize)
}

func readBoxes(m *Mp4Reader, start int64, n int64) (l []*Box) {
	for offset := start; offset < start+n; {
		size, name := m.ReadBoxAt(offset)

		b := &Box{
			Name:   string(name),
			Size:   int64(size),
			Reader: m,
			Start:  offset,
		}

		l = append(l, b)
		offset += int64(size)
	}
	return l
}

// Open opens a file and returns an &Mp4Reader{}.
func Open(path string) (f *Mp4Reader, err error) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil, err
	}

	f = &Mp4Reader{
		Reader: file,
	}
	return f, f.Parse()
}

// Fixed16 is an 8.8 Fixed Point Decimal notation
type Fixed16 uint16

func (f Fixed16) String() string {
	return fmt.Sprintf("%v", uint16(f)>>8)
}

func fixed16(bytes []byte) Fixed16 {
	return Fixed16(binary.BigEndian.Uint16(bytes))
}

// Fixed32 is a 16.16 Fixed Point Decimal notation
type Fixed32 uint32

func fixed32(bytes []byte) Fixed32 {
	return Fixed32(binary.BigEndian.Uint32(bytes))
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("missing argument, provide an mp4 file!")
		return
	}

	mp4, err := Open(os.Args[1])
	if err != nil {
		fmt.Println("unable to open file")
	}

	fmt.Println("ftyp.name: ", mp4.Ftyp.Name)
	fmt.Println("ftyp.major_brand: ", mp4.Ftyp.MajorBrand)
	fmt.Println("ftyp.minor_version: ", mp4.Ftyp.MinorVersion)
	fmt.Println("ftyp.compatible_brands: ", mp4.Ftyp.CompatibleBrands)

	fmt.Println("moov.name: ", mp4.Moov.Name, mp4.Moov.Size)
	fmt.Println("moov.mvhd.name: ", mp4.Moov.Mvhd.Name)
	fmt.Println("moov.mvhd.version: ", mp4.Moov.Mvhd.Version)
	fmt.Println("moov.mvhd.volume: ", mp4.Moov.Mvhd.Volume)
}

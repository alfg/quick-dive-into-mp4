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

// Mp4Reader defines an mp4 reader structure.
type Mp4Reader struct {
	Reader io.ReaderAt
	Ftyp   *FtypBox
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

			// Add cases to check for more boxes here.
			// case "moov":
			// 	m.Moov = &MoovBox{Box: box}
			// 	m.Moov.parse()
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
}

# A Quick Dive Into MP4

What is an MP4? We all know it as a file format for playing video with sound. It's used for streaming video by Netflix, YouTube, Instagram, and also for capturing video on your iPhone, but how does it work? How is it used? What is the byte structure? What is a container?

This guide is an introduction and a quick dive into the MP4 file format, also known as the ISO Base Media File Format ([ISO-BMFF MPEG-4 Part 14](https://en.wikipedia.org/wiki/MPEG-4_Part_14)). Fancy name, I know.

I won't go into the playback details in this guide, but more of the MP4 byte format commonly known as the MP4 Box Structure.

## Introduction
The MPEG-4 Part 14 (MP4) is one of the most common [container formats](https://en.wikipedia.org/wiki/Comparison_of_video_container_formats) for video and has an extension of `.mp4`. You may already know of other container formats, such as `wav`, `mov`, `mp3` or more recently `webm`. A container just "contains" the video or audio track, or both. It can also support embedded subtitle tracks too.

MP4 is an extension of the ISO Base Media File Format ([ISOBMFF, MPEG-4 Part 12)](https://en.wikipedia.org/wiki/ISO/IEC_base_media_file_format), which is a format designed to contain timed media information.

The ISO-BMFF format is directly based on [QuickTime](https://en.wikipedia.org/wiki/QuickTime_File_Format), therefore the MP4 is essentially identical to the QuickTime file format.

```
     MPEG-4 Part 14 - MP4 File Format
 ┌──────────────────────────────────────┐
 ├────────────────────┐                 │
 │  ISO Base Media    │                 │
 │    File Format     │  MP4 Extension  │
 │  (MPEG-4 Part 12)  │                 │
 └────────────────────┴─────────────────┘
```

In order to fully understand the MP4 structure, you'll need to obtain a copy of the [ISO](https://en.wikipedia.org/wiki/International_Organization_for_Standardization) documents:
* 14496-12 – MPEG-4 Part 12
* 14496-14 - MPEG-4 Part 14

A Google search should result in a few resources to get a copy of the PDF.

## What's in a container?
Since MP4 is a [container](https://en.wikipedia.org/wiki/Comparison_of_video_container_formats) format, it doesn't actually handle the decoding of the video and audio streams, it just contains them as tracks along with their metadata.

A container can store some of the following information:
* General metadata such as file type and compatibility
* Video, audio and subtitle tracks and codec details
* Metadata: duration, timescale, bitrate, width/height, etc
* Progressive and fragmented metadata details
* A series of video frames or audio samples known as "Sample Data"

This is all the information a player needs to decode and play the content.

At a high-level, this is what an MP4 structure typically looks like:

```
video.mp4
├───general file metadata
├───movie data
├───tracks
│   ├───video
│   │   ├───video metadata
│   │   └───video sample data
│   └───audio
│       ├───audio metadata
│       └───audio sample data
└───more metadata
```

## Movie Boxes
The MP4 byte structure is composed of a series of boxes, also known as "atoms", according to the QuickTime specification. Each box describes and contains data to build the MP4 container format.

Boxes typically have a four letter name, also known as a [FourCC](https://en.wikipedia.org/wiki/FourCC). This is the shortened version of the full box name, enough to fit into 4 bytes. This is important for when you are reading and writing boxes into or from the byte format.

Before we jump into the byte structure details, here is a more technical view of the MP4 box tree, compared to the high-level view above:

```
video.mp4
├───ftyp -------------------> FileType Box
├───mdat -------------------> Movie Data Box
├───moov -------------------> Movie Boxes
│   ├───trak ---------------> Track Box
│   │   ├─── tkhd ----------> Track Header
│   │   └─── mdia ----------> Media Box
│   │        └─── ...
│   └───trak
│   │   ├─── tkhd ----------> Track Header
│   │   └─── mdia ----------> Media Box
│   │        └─── ...
└───udta -------------------> Userdata Box
```

This is just a simplified view. However, there are many more boxes defined in the MP4 specification.

## What's in the box?
An MP4 "Box" contains just enough information to read and parse the box name, size and data.

Each of these boxes have a different purpose, containing a bit of information and details on a specific piece of data. Some boxes describe the file type, and others can describe codec detail, picture resolution, frame rate, duration, sample sizes and more. There's also boxes containing the encoded video and audio data too.

A box typically contains the following base information:
* Size of the box (in bytes)
* Box Name (FourCC)
* Box Data

```
┌─────────────────────┐
|      Box Header     |
| Size (4) | Type (4) | Box Header = 8 Bytes
| --------------------|
|     Box Data (N)    | Box Data = N Bytes
└─────────────────────┘
           └─────────── Box Size = 8 + N bytes
```

This is just enough information we need to know how to parse a box, along with the MP4 specification document to understand the box fields.

## Parse a Box
So let's parse our first box!

As mentioned above, the first 8 bytes of each box is known as the "Box Header", where the first 4 bytes are the size of the box, and the next 4 bytes are the box name. These are the two values you need to know to iterate and parse each box, byte by byte.

Here's a box header struct for example:
```go
type Box struct {
  Size    int32
  Name    string
}
```

Reading the box data from each atom requires the box size, name and byte structure of each box you are parsing. You can refer to the MPEG-4 Part 14 specification for the byte structure of each known box, or just refer to some existing MP4 parsing open-source code.

According to the specification, the `ftyp` box has the following structure:
```c
aligned(8) class FileTypeBox
  extends Box(‘ftyp’) {
  unsigned int(32) major_brand;
  unsigned int(32) minor_version;
  unsigned int(32) compatible_brands[]; // to end of the box
}
```

* `major_brand` – is a brand identifier
* `minor_version` – is an informative integer for the minor version of the major brand
* `compatible_brands` – is a list, to the end of the box, of brands

For example, reading the FileTypeBox (`ftyp`) would look something like the following (in Golang):

```go
type FtypBox struct {
  *Box
  MajorBrand       string
  MinorVersion     uint32
  CompatibleBrands []string
}

func (b *FtypBox) parse() {
  data := b.ReadBoxData() // Read box header.
  b.MajorBrand = string(data[0:4])
  b.MinorVersion = binary.BigEndian.Uint32(data[4:8])
  if len(data) > 8 {
    for i := 8; i < len(data); i += 4 {
      b.CompatibleBrands = append(b.CompatibleBrands, string(data[i:i+4]))
    }
  }
}
```

Going over the above:
* Reading the first 4 bytes of the box header as an unsigned 32 bit integer (big endian) gives us the box size: `32 bytes`.
* The next 4 bytes gives us: `0x66747970` in hexidecimal, or `ftyp` as a string.
* Next 4 bytes gives us the Major Brand: `0x69736F6D` in hexidecimal, or `isom` as a string.
* Next 4 bytes gives us the Minor Version: `512`
* The next 16 bytes, read as `uint32be` (into a string) at a time gives us an array of compatible brands: `isom`, `iso2`, `avc1`,
 and `mp41`.
* We have read a total of `32 bytes` as defined in the box header. 

See a minimal example:
https://gist.github.com/alfg/7375aee32fda490de4bf62fbced49d2e#file-mp4_example-go

```
$ go run mp4.go tears-of-steel.mp4
ftyp.name:  ftyp
ftyp.major_brand:  isom
ftyp.minor_version:  512
ftyp.compatible_brands:  [isom iso2 avc1 mp41]
```

If you were to open the mp4 into a hex editor, it would look something like this for the `ftyp` box:

```
0x00 00 00 00 20 66 74 79 70 | 69 73 6F 6D 00 00 02 00 ... ftypisom....
0x10 69 73 6F 6D 69 73 6F 32 | 61 76 63 31 6D 70 34 31 isomiso2avc1mp41
```

## Next Box!

Now that we've read the `ftyp` box, we can move on to the next box header, which happens to be the `moov` box:

```go
type MoovBox struct {
  *Box
  Mvhd  *MvhdBox
}

func (b *MoovBox) parse() {
  boxes := readBoxes(b.Reader, b.Start+BoxHeaderSize, b.Size-BoxHeaderSize)

  for _, box := range boxes {
    switch box.Name {
    case "mvhd":
    b.Mvhd = &MvhdBox{Box: box}
    b.Mvhd.parse()
  }
}
```

The `moov` box contains a nested `mvhd` box, so we also need to define `mvhd` too:

```go
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

func (b *MvhdBox) parse() {
  data := b.ReadBoxData()
  b.Version = data[0]
  b.Timescale = binary.BigEndian.Uint32(data[12:16])
  b.Duration = binary.BigEndian.Uint32(data[16:20])
  b.Rate = fixed32(data[20:24])
  b.Volume = fixed16(data[24:26])
}
```

See the example with `moov` and `mvhd` box included:
https://gist.github.com/alfg/7375aee32fda490de4bf62fbced49d2e#file-mp4_example_2-go

```
$ go run mp4.go tears-of-steel.mp4
ftyp.name:  ftyp
ftyp.major_brand:  isom
ftyp.minor_version:  512
ftyp.compatible_brands:  [isom iso2 avc1 mp41]
moov.name:  moov 3170
moov.mvhd.name:  mvhd
moov.mvhd.version:  0
moov.mvhd.volume:  1
```

## EOF

Now that we've parsed 3 boxes, hopefully you have an idea on how to implement more. The process is iterative when using a reader:
* Read the box header, containing the box size and name.
* Refer to the specification to read and/or skip fields.
* Skip any remaining bytes left in the box size.
* Read the next box (or skip).

##### Some things to keep in mind:
* Some boxes have multiple versions, and therefore can differ in the struct and overall size of the box.
* You can skip properties, but the reader must know how many bytes to skip.
* There are various MP4 specifications beyond `MPEG-4 Part 14` as more boxes are being added throughout the years.
* Fragmented MP4 files (fMP4) are segmented as a series of `moof` and `mdat` boxes. This is more common and optimial for streaming delivery. I'll cover this in a future post.

## Thanks for reading!
For a more complete example of reading MP4 boxes in Go, check out:
https://github.com/alfg/mp4

I also have a more advanced MP4 reader and writer in Rust:
https://github.com/alfg/mp4rs

I highly suggest some of the following tools for inspecting MP4 files:
* [MP4Box](https://github.com/gpac/gpac/wiki/MP4Box)
* [mp4box.js](https://gpac.github.io/mp4box.js)
* [MediaInfo](https://github.com/MediaArea/MediaInfo)
* [FFProbe](https://ffmpeg.org/ffprobe.html)

Find me on GitHub at: https://github.com/alfg

Happy Hacking! 🎥

# References and Resources
* https://developer.apple.com/library/archive/documentation/QuickTime/QTFF
* https://en.wikipedia.org/wiki/MPEG-4_Part_14
* https://en.wikipedia.org/wiki/ISO/IEC_base_media_file_format
* https://en.wikipedia.org/wiki/Comparison_of_video_container_formats
* https://en.wikipedia.org/wiki/FourCC
* https://gist.github.com/alfg/7375aee32fda490de4bf62fbced49d2e
* https://github.com/alfg

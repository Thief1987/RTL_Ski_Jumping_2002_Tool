package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	TOC_buf, name_buf, data_buf bytes.Buffer
	Path                        string
)

func ReadUint64LE(r io.Reader) uint64 {
	var buf bytes.Buffer
	io.CopyN(&buf, r, 8)
	return binary.LittleEndian.Uint64(buf.Bytes())
}

func ReadUint16LE(r io.Reader) uint16 {
	var buf bytes.Buffer
	io.CopyN(&buf, r, 2)
	return binary.LittleEndian.Uint16(buf.Bytes())
}

func ReadUint32LE(r io.Reader) uint32 {
	var buf bytes.Buffer
	io.CopyN(&buf, r, 4)
	return binary.LittleEndian.Uint32(buf.Bytes())
}

func PathCreation(n string, nl int) string {
	for l := 1; l < nl; l++ {
		if string(n[nl-l]) == "\\" {
			Path = n[:nl-l]
			break
		}
	}
	return Path
}

func main() {
	args := os.Args
	arc, _ := os.Open(args[1])
	os.Mkdir(arc.Name()[:len(arc.Name())-4], 0700)
	os.Chdir(arc.Name()[:len(arc.Name())-4])
	arc.Seek(4, 0)
	TOC_size := (ReadUint32LE(arc) * 0x800)
	files := TOC_size / 0x50 // Not exact number
	arc.Seek(0, 0)
	io.CopyN(&TOC_buf, arc, int64(TOC_size))
	for i := 0; i < int(files); i++ {
		Zsize := ReadUint32LE(&TOC_buf)
		if Zsize == 0 {
			break
		}
		offset := ReadUint32LE(&TOC_buf) * 0x800
		_ = ReadUint32LE(&TOC_buf) // ??
		fsize := ReadUint32LE(&TOC_buf)
		io.CopyN(&name_buf, &TOC_buf, 0x40)
		name := strings.Replace(name_buf.String(), "\x00", "", -1)
		fmt.Println(name)
		name_buf.Reset()
		os.MkdirAll(PathCreation(name, len(name)), 0700)
		f, _ := os.Create(name)
		if Zsize == fsize {
			arc.Seek(int64(offset), 0)
			io.CopyN(f, arc, int64(fsize))
		} else {
			arc.Seek(int64(offset), 0)
			io.CopyN(&data_buf, arc, int64(Zsize))
			f.Write(LZdecompress(data_buf))
			data_buf.Reset()
		}
	}
}

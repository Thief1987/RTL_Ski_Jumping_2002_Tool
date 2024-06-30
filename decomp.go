// LZ-like compression algo used in the RTL Ski Jumping 2002 game
package main

import (
	"bytes"
	"encoding/binary"
	"io"
)

var (
	token   = make([]byte, 1)
	dec_buf bytes.Buffer
)

func LZdecompress(lz bytes.Buffer) []byte {
	dec_buf.Reset()
	size := ReadUint64LE(&lz)
	for dec_buf.Len() < int(size) {
		lz.Read(token)
		if token[0] < 0x20 {
			if token[0] == 0 {
				lz.Read(token)
				if token[0] == 0 {
					size := ReadUint16LE(&lz)
					if size == 0 {
						break
					}
					io.CopyN(&dec_buf, &lz, int64(size))
				} else {
					size := int(token[0]) + 0x1F
					io.CopyN(&dec_buf, &lz, int64(size))
				}
			} else {
				io.CopyN(&dec_buf, &lz, int64(token[0]))
			}
		} else if token[0] >= 0x20 && token[0] < 0x40 {
			tmp := int(token[0]) - 0x20
			if tmp == 0 {
				lz.Read(token)
				tmp = tmp + 0x20 + int(token[0])
			}
			dw := tmp / 4
			b := tmp % 4
			for j := 0; j < int(dw); j++ {
				binary.Write(&dec_buf, binary.LittleEndian, uint32(0))
			}
			for j := 0; j < int(b); j++ {
				binary.Write(&dec_buf, binary.LittleEndian, uint8(0))
			}
		} else if token[0] >= 0x80 {
			check := token[0] & 0x40
			if int(check) != 0 {
				io.CopyN(&dec_buf, &lz, 2)
			}
			offset := (token[0]&0x3F)*2 + 2
			copy := dec_buf.Bytes()[dec_buf.Len()-int(offset) : dec_buf.Len()-int(offset)+2]
			dec_buf.Write(copy)
		} else if token[0] >= 0x40 && token[0] < 0x80 {
			check := (token[0] & 0x0F) + 2
			if check == 2 {
				length := ReadUint16LE(&lz)
				offset := ReadUint16LE(&lz)
				for token[0]&0x30 != 0 {
					io.CopyN(&dec_buf, &lz, 1)
					token[0] = token[0] - 0x10
				}
				copy := dec_buf.Bytes()[dec_buf.Len()-int(length)-int(offset)+1 : dec_buf.Len()-int(offset)+1]
				dec_buf.Write(copy)
			} else {
				offset := ReadUint16LE(&lz)
				for token[0]&0x30 != 0 {
					io.CopyN(&dec_buf, &lz, 1)
					token[0] = token[0] - 0x10
				}
				if check != 0 {
					copy := dec_buf.Bytes()[dec_buf.Len()-int(offset)-int(check)+1 : dec_buf.Len()-int(offset)+1]
					dec_buf.Write(copy)
				}
			}
		}
	}
	return dec_buf.Bytes()
}

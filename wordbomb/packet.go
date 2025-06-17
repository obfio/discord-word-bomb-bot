package wordbomb

import (
	"errors"
	"fmt"
	"unicode/utf16"
)

var (
	packetIDs = map[int]string{
		9:  "HANDSHAKE",
		10: "JOIN_ROOM",
		11: "ERROR",
		12: "LEAVE_ROOM",
		13: "ROOM_DATA",
		14: "ROOM_STATE",
		15: "ROOM_STATE_PATCH",
		16: "ROOM_DATA_SCHEMA",
		17: "ROOM_DATA_BYTES",
	}

	packetIDReverse = map[string]int{
		"HANDSHAKE":        9,
		"JOIN_ROOM":        10,
		"ERROR":            11,
		"LEAVE_ROOM":       12,
		"ROOM_DATA":        13,
		"ROOM_STATE":       14,
		"ROOM_STATE_PATCH": 15,
		"ROOM_DATA_SCHEMA": 16,
		"ROOM_DATA_BYTES":  17,
	}
)

func getPacketID(msg []byte) string {
	pID := msg[0]

	return packetIDs[int(pID)]
}

func (c *Client) handleJoinRoomPacket(msg []byte) {
	offset := 1

	roomID, err := utf8Read(msg, offset)
	if err != nil {
		return
	}

	offset += len(roomID)

	c.ReconnectionToken = fmt.Sprintf("%s:%s", c.RoomID, roomID)

	c.Send([]byte{byte(packetIDReverse["JOIN_ROOM"])})

}

type RoomDataPacket struct {
	MessageType    interface{}
	MessagePayload interface{}
}

func (c *Client) handleRoomDataPacket(msg []byte) (RoomDataPacket, error) {
	decodeCtx := decodeCtx{offset: 1}

	var messageType interface{}

	if decodeStringCheck(msg, decodeCtx.offset) {
		messageType = decodeString(msg, &decodeCtx)
	} else {
		messageType = decodeNumber(msg, &decodeCtx)
	}

	if !(len(msg) > decodeCtx.offset) {
		return RoomDataPacket{}, errors.New("offset out of range")
	}
	messagePayload, err := decodePayload(msg, decodeCtx.offset)
	if err != nil {
		return RoomDataPacket{}, err
	}

	return RoomDataPacket{
		MessageType:    messageType,
		MessagePayload: messagePayload,
	}, nil

}

func utf8Read(buffer []byte, offset int) (string, error) {
	if offset >= len(buffer) {
		return "", errors.New("offset out of range")
	}

	length := int(buffer[offset])
	offset++

	end := offset + length
	if end > len(buffer) {
		return "", errors.New("buffer too small")
	}

	result := make([]rune, 0)

	for offset < end {
		b := buffer[offset]
		offset++

		if b&0x80 == 0 {
			// 1-byte
			result = append(result, rune(b))
		} else if b&0xE0 == 0xC0 {
			if offset >= end {
				return "", errors.New("unexpected end for 2-byte character")
			}
			b2 := buffer[offset]
			offset++
			r := rune(b&0x1F)<<6 | rune(b2&0x3F)
			result = append(result, r)
		} else if b&0xF0 == 0xE0 {
			if offset+1 >= end {
				return "", errors.New("unexpected end for 3-byte character")
			}
			b2 := buffer[offset]
			b3 := buffer[offset+1]
			offset += 2
			r := rune(b&0x0F)<<12 | rune(b2&0x3F)<<6 | rune(b3&0x3F)
			result = append(result, r)
		} else if b&0xF8 == 0xF0 {
			if offset+2 >= end {
				return "", errors.New("unexpected end for 4-byte character")
			}
			b2 := buffer[offset]
			b3 := buffer[offset+1]
			b4 := buffer[offset+2]
			offset += 3
			cp := rune(b&0x07)<<18 | rune(b2&0x3F)<<12 | rune(b3&0x3F)<<6 | rune(b4&0x3F)
			if cp >= 0x10000 {
				cp -= 0x10000
				hi, lo := utf16.EncodeRune(cp)
				result = append(result, hi, lo)
			} else {
				result = append(result, cp)
			}
		} else {
			return "", fmt.Errorf("invalid UTF-8 byte: 0x%X", b)
		}
	}

	return string(result), nil
}

/*
	h.decode = function(e, t=0) {
	    var n = new c(e,t)
	      , s = n._parse();
	    if (n._offset !== e.byteLength)
	        throw new Error(e.byteLength - n._offset + " trailing bytes");
	    return s
	}
*/
func decodePayload(buffer []byte, offset int) (interface{}, error) {
	parser := &Parser{View: buffer, Offset: offset}
	payload, err := parser.Parse()
	if err != nil {
		return nil, err
	}

	return payload, nil
}

func decodeStringCheck(buffer []byte, offset int) bool {
	n := buffer[offset]
	return n < 192 && n > 160 || 217 == n || 218 == n || 219 == n
}

type decodeCtx struct {
	offset int
}

/*
	function ye(e, t) {
	        var n = e[t.offset++];
	        return n < 128 ? n : 202 === n ? ue(e, t) : 203 === n ? he(e, t) : 204 === n ? Y(e, t) : 205 === n ? ee(e, t) : 206 === n ? ne(e, t) : 207 === n ? se(e, t) : 208 === n ? Q(e, t) : 209 === n ? Z(e, t) : 210 === n ? te(e, t) : 211 === n ? oe(e, t) : n > 223 ? -1 * (255 - n + 1) : void 0
	    }
*/
func decodeNumber(buffer []byte, ctx *decodeCtx) int {
	n := buffer[ctx.offset]
	ctx.offset++

	if n < 128 {
		return int(n)
	}

	switch n {
	case 202:
		panic("unhandled case 202")
		// return ue(buffer, ctx)
	case 203:
		panic("unhandled case 203")
		// return he(buffer, ctx)
	case 204:
		return Y(buffer, ctx)
	case 205:
		return ee(buffer, ctx)
	case 206:
		return ne(buffer, ctx)
	case 207:
		panic("unhandled case 207")
		// return se(buffer, ctx)
	case 208:
		panic("unhandled case 208")
		// return Q(buffer, ctx)
	case 209:
		panic("unhandled case 209")
		// return Z(buffer, ctx)
	case 210:
		return te(buffer, ctx)
	case 211:
		panic("unhandled case 211")
		// return oe(buffer, ctx)
	default:
		return -1 * (255 - int(n) + 1)
	}
}

func decodeString(buffer []byte, ctx *decodeCtx) string {
	r := buffer[ctx.offset]
	ctx.offset++
	n := 0
	if r < 192 {
		n = int(31 & r)
	} else if r == 217 {
		n = Y(buffer, ctx)
	} else if r == 218 {
		n = ee(buffer, ctx)
	} else if r == 219 {
		n = ne(buffer, ctx)
	}

	i, err := decodeUtf8Range(buffer, ctx.offset, n)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	ctx.offset += n
	return i
}

func decodeUtf8Range(buffer []byte, start int, length int) (string, error) {
	var result []rune
	var codePoint rune
	offset := start
	end := start + length

	if end > len(buffer) {
		return "", errors.New("range exceeds buffer size")
	}

	for offset < end {
		byte1 := buffer[offset]
		offset++

		switch {
		case byte1&0x80 == 0x00:
			// 1-byte (ASCII)
			result = append(result, rune(byte1))

		case byte1&0xE0 == 0xC0:
			if offset >= end {
				return "", errors.New("incomplete 2-byte sequence")
			}
			byte2 := buffer[offset]
			offset++
			codePoint = rune(byte1&0x1F)<<6 | rune(byte2&0x3F)
			result = append(result, codePoint)

		case byte1&0xF0 == 0xE0:
			if offset+1 >= end {
				return "", errors.New("incomplete 3-byte sequence")
			}
			byte2 := buffer[offset]
			byte3 := buffer[offset+1]
			offset += 2
			codePoint = rune(byte1&0x0F)<<12 | rune(byte2&0x3F)<<6 | rune(byte3&0x3F)
			result = append(result, codePoint)

		case byte1&0xF8 == 0xF0:
			if offset+2 >= end {
				return "", errors.New("incomplete 4-byte sequence")
			}
			byte2 := buffer[offset]
			byte3 := buffer[offset+1]
			byte4 := buffer[offset+2]
			offset += 3
			codePoint = rune(byte1&0x07)<<18 |
				rune(byte2&0x3F)<<12 |
				rune(byte3&0x3F)<<6 |
				rune(byte4&0x3F)

			if codePoint >= 0x10000 {
				codePoint -= 0x10000
				hi, lo := utf16.EncodeRune(codePoint)
				result = append(result, hi, lo)
			} else {
				result = append(result, codePoint)
			}

		default:
			return "", fmt.Errorf("invalid UTF-8 byte: 0x%X", byte1)
		}
	}

	return string(result), nil
}

func Y(buffer []byte, ctx *decodeCtx) int {
	o := buffer[ctx.offset]
	ctx.offset++

	return int(o)
}

func ee(buffer []byte, ctx *decodeCtx) int {
	o := buffer[ctx.offset]
	ctx.offset++
	o1 := buffer[ctx.offset]
	ctx.offset++

	return int(o | o1<<8)
}

func ne(buffer []byte, ctx *decodeCtx) int {
	return trippleShift(te(buffer, ctx), 0)
}
func te(buffer []byte, ctx *decodeCtx) int {
	o := buffer[ctx.offset]
	ctx.offset++
	o1 := buffer[ctx.offset]
	ctx.offset++
	o2 := buffer[ctx.offset]
	ctx.offset++
	o3 := buffer[ctx.offset]
	ctx.offset++

	return int(o) | int(o1)<<8 | int(o2)<<16 | int(o3)<<24
}

func trippleShift(num, t int) int {
	overflow := int32(num)
	return int(uint32(overflow) >> t)
}

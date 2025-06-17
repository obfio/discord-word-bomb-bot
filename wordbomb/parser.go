package wordbomb

import (
	"encoding/binary"
	"errors"
	"math"
	"time"
)

type Parser struct {
	View   []byte
	Offset int
}

func (p *Parser) readUint8() byte {
	val := p.View[p.Offset]
	p.Offset++
	return val
}

func (p *Parser) readUint16() uint16 {
	val := binary.BigEndian.Uint16(p.View[p.Offset:])
	p.Offset += 2
	return val
}

func (p *Parser) readUint32() uint32 {
	val := binary.BigEndian.Uint32(p.View[p.Offset:])
	p.Offset += 4
	return val
}

func (p *Parser) readInt8() int8 {
	val := int8(p.View[p.Offset])
	p.Offset++
	return val
}

func (p *Parser) readInt16() int16 {
	val := int16(binary.BigEndian.Uint16(p.View[p.Offset:]))
	p.Offset += 2
	return val
}

func (p *Parser) readInt32() int32 {
	val := int32(binary.BigEndian.Uint32(p.View[p.Offset:]))
	p.Offset += 4
	return val
}

func (p *Parser) readFloat32() float32 {
	bits := binary.BigEndian.Uint32(p.View[p.Offset:])
	p.Offset += 4
	return math.Float32frombits(bits)
}

func (p *Parser) readFloat64() float64 {
	bits := binary.BigEndian.Uint64(p.View[p.Offset:])
	p.Offset += 8
	return math.Float64frombits(bits)
}

func (p *Parser) Parse() (any, error) {
	if p.Offset >= len(p.View) {
		return nil, errors.New("out of bounds")
	}

	t := p.readUint8()
	var length int
	var extType int8
	var intPart, fracPart uint32

	switch {
	case t < 128:
		return int(t), nil
	case t < 144:
		return p.readMap(int(t & 0x0F))
	case t < 160:
		return p.readArray(int(t & 0x0F))
	case t < 192:
		return p.readString(int(t & 0x1F))
	case t > 223:
		return -1 * (255 - int(t) + 1), nil
	}

	switch t {
	case 192:
		return nil, nil
	case 194:
		return false, nil
	case 195:
		return true, nil
	case 196:
		length = int(p.readUint8())
		return p.readBin(length), nil
	case 197:
		length = int(p.readUint16())
		return p.readBin(length), nil
	case 198:
		length = int(p.readUint32())
		return p.readBin(length), nil
	case 199:
		length = int(p.readUint8())
		extType = p.readInt8()
		if extType == -1 {
			nano := p.readUint32()
			i := int64(p.readInt32())
			o := int64(p.readUint32())
			return time.UnixMilli(i*int64(4294967296) + o).Add(time.Duration(nano) * time.Nanosecond), nil
		}
		return []any{extType, p.readBin(length)}, nil
	case 200:
		length = int(p.readUint16())
		extType = p.readInt8()
		return []any{extType, p.readBin(length)}, nil
	case 201:
		length = int(p.readUint32())
		extType = p.readInt8()
		return []any{extType, p.readBin(length)}, nil
	case 202:
		return p.readFloat32(), nil
	case 203:
		return p.readFloat64(), nil
	case 204:
		return int(p.readUint8()), nil
	case 205:
		return int(p.readUint16()), nil
	case 206:
		return int(p.readUint32()), nil
	case 207:
		intPart = p.readUint32()
		fracPart = p.readUint32()
		return float64(uint64(intPart))*math.Pow(2, 32) + float64(fracPart), nil
	case 208:
		return int(p.readInt8()), nil
	case 209:
		return int(p.readInt16()), nil
	case 210:
		return int(p.readInt32()), nil
	case 211:
		i := int64(p.readInt32())
		o := int64(p.readUint32())
		return i*4294967296 + o, nil
	case 212:
		extType = p.readInt8()
		if extType == 0 {
			p.Offset++
			return nil, nil
		}
		return []any{extType, p.readBin(1)}, nil
	case 213:
		extType = p.readInt8()
		return []any{extType, p.readBin(2)}, nil
	case 214:
		extType = p.readInt8()
		if extType == -1 {
			secs := p.readUint32()
			return time.Unix(int64(secs), 0), nil
		}
		return []any{extType, p.readBin(4)}, nil
	case 215:
		extType = p.readInt8()
		if extType == 0 {
			i := int64(p.readInt32())
			o := int64(p.readUint32())
			return time.UnixMilli(i*4294967296 + o), nil
		} else if extType == -1 {
			i := p.readUint32()
			o := p.readUint32()
			nano := i >> 2
			secs := uint64(i&3)*uint64(4294967296) + uint64(o)
			return time.Unix(int64(secs), int64(nano)*1e3), nil
		}
		return []any{extType, p.readBin(8)}, nil
	case 216:
		extType = p.readInt8()
		return []any{extType, p.readBin(16)}, nil
	case 217:
		length = int(p.readUint8())
		return p.readString(length)
	case 218:
		length = int(p.readUint16())
		return p.readString(length)
	case 219:
		length = int(p.readUint32())
		return p.readString(length)
	case 220:
		length = int(p.readUint16())
		return p.readArray(length)
	case 221:
		length = int(p.readUint32())
		return p.readArray(length)
	case 222:
		length = int(p.readUint16())
		return p.readMap(length)
	case 223:
		length = int(p.readUint32())
		return p.readMap(length)
	}

	return nil, errors.New("unsupported type")
}

func (p *Parser) readBin(length int) []byte {
	data := p.View[p.Offset : p.Offset+length]
	p.Offset += length
	return data
}

func (p *Parser) readString(length int) (string, error) {
	data := p.readBin(length)
	return string(data), nil
}

func (p *Parser) readArray(length int) ([]any, error) {
	result := make([]any, length)
	for i := 0; i < length; i++ {
		val, err := p.Parse()
		if err != nil {
			return nil, err
		}
		result[i] = val
	}
	return result, nil
}

func (p *Parser) readMap(length int) (map[string]any, error) {
	result := make(map[string]any)
	for i := 0; i < length; i++ {
		keyRaw, err := p.Parse()
		if err != nil {
			return nil, err
		}
		key, ok := keyRaw.(string)
		if !ok {
			return nil, errors.New("non-string map key")
		}
		val, err := p.Parse()
		if err != nil {
			return nil, err
		}
		result[key] = val
	}
	return result, nil
}

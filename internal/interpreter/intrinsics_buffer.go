package interpreter

import (
	"bytes"
	"encoding/base64"
	encbinary "encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func executeBufferIntrinsic(inst ir.Instruction, env map[string]Value) (Value, error) {
	callee := inst.Callee
	args := inst.Args

	switch callee {
	case "__buffer.alloc":
		if len(args) < 4 {
			return Value{}, fmt.Errorf("__buffer.alloc requires 4 arguments")
		}
		sizeVal := env[args[0]]
		fillVal := env[args[1]]
		hasFillVal := env[args[2]]
		isStrVal := env[args[3]]

		size := int(sizeVal.Number)
		if size < 0 {
			return Value{}, fmt.Errorf("Buffer size must be non-negative")
		}
		data := make([]byte, size)

		if hasFillVal.Bool {
			if isStrVal.Bool {
				str := fillVal.String
				if len(str) > 0 {
					for i := 0; i < size; i++ {
						data[i] = str[i%len(str)]
					}
				}
			} else {
				b := byte(uint32(fillVal.Number) & 0xFF)
				for i := range data {
					data[i] = b
				}
			}
		}

		return Value{
			Type: ir.TypeBuffer,
			TypedArray: &TypedArray{
				Kind:       ir.TypeBuffer,
				Buffer:     &ArrayBuffer{Data: data},
				ByteOffset: 0,
				Length:     size,
			},
		}, nil

	case "__buffer.from_string":
		if len(args) < 2 {
			return Value{}, fmt.Errorf("__buffer.from_string requires 2 arguments")
		}
		str := env[args[0]].String
		enc := strings.ToLower(env[args[1]].String)

		var data []byte
		switch enc {
		case "hex":
			decoded, err := hex.DecodeString(str)
			if err == nil {
				data = decoded
			}
		case "base64", "base64url":
			decoded, err := base64.StdEncoding.DecodeString(str)
			if err != nil {
				decoded, _ = base64.RawStdEncoding.DecodeString(str)
			}
			data = decoded
		case "ascii":
			data = make([]byte, len(str))
			for i := 0; i < len(str); i++ {
				data[i] = str[i] & 0x7F
			}
		default: // utf8, utf-8, latin1, binary
			data = []byte(str)
		}

		return Value{
			Type: ir.TypeBuffer,
			TypedArray: &TypedArray{
				Kind:       ir.TypeBuffer,
				Buffer:     &ArrayBuffer{Data: data},
				ByteOffset: 0,
				Length:     len(data),
			},
		}, nil

	case "__buffer.from_array":
		if len(args) < 1 {
			return Value{}, fmt.Errorf("__buffer.from_array requires 1 argument")
		}
		val := env[args[0]]
		var data []byte
		if val.TypedArray != nil && val.TypedArray.Buffer != nil {
			src := val.TypedArray
			start := src.ByteOffset
			end := start + src.ByteLength()
			if end > len(src.Buffer.Data) {
				end = len(src.Buffer.Data)
			}
			data = make([]byte, end-start)
			copy(data, src.Buffer.Data[start:end])
		} else if val.Buffer != nil {
			data = make([]byte, len(val.Buffer.Data))
			copy(data, val.Buffer.Data)
		} else if val.Array != nil {
			data = make([]byte, len(val.Array))
			for i, v := range val.Array {
				data[i] = byte(uint32(v.Number) & 0xFF)
			}
		}

		return Value{
			Type: ir.TypeBuffer,
			TypedArray: &TypedArray{
				Kind:       ir.TypeBuffer,
				Buffer:     &ArrayBuffer{Data: data},
				ByteOffset: 0,
				Length:     len(data),
			},
		}, nil

	case "__buffer.concat":
		if len(args) < 2 {
			return Value{}, fmt.Errorf("__buffer.concat requires 2 arguments")
		}
		listVal := env[args[0]]
		totLenVal := env[args[1]]

		var buffers [][]byte
		totalLen := 0
		if listVal.Array != nil {
			for _, item := range listVal.Array {
				if item.TypedArray != nil && item.TypedArray.Buffer != nil {
					ta := item.TypedArray
					start := ta.ByteOffset
					end := start + ta.ByteLength()
					if end > len(ta.Buffer.Data) {
						end = len(ta.Buffer.Data)
					}
					slice := ta.Buffer.Data[start:end]
					buffers = append(buffers, slice)
					totalLen += len(slice)
				}
			}
		}

		if totLenVal.Number >= 0 {
			totalLen = int(totLenVal.Number)
		}

		outData := make([]byte, totalLen)
		off := 0
		for _, b := range buffers {
			if off >= totalLen {
				break
			}
			n := copy(outData[off:], b)
			off += n
		}

		return Value{
			Type: ir.TypeBuffer,
			TypedArray: &TypedArray{
				Kind:       ir.TypeBuffer,
				Buffer:     &ArrayBuffer{Data: outData},
				ByteOffset: 0,
				Length:     totalLen,
			},
		}, nil

	case "__buffer.isBuffer":
		if len(args) < 1 {
			return Value{}, fmt.Errorf("__buffer.isBuffer requires 1 argument")
		}
		val := env[args[0]]
		if val.Type == ir.TypeUnknown && val.Boxed != nil {
			val = *val.Boxed
		}
		isBuf := val.Type == ir.TypeBuffer || (val.TypedArray != nil && val.TypedArray.Kind == ir.TypeBuffer)
		return Value{Type: ir.TypeBool, Bool: isBuf}, nil

	case "__buffer.byteLength":
		if len(args) < 2 {
			return Value{}, fmt.Errorf("__buffer.byteLength requires 2 arguments")
		}
		str := env[args[0]].String
		enc := strings.ToLower(env[args[1]].String)
		length := len(str)
		switch enc {
		case "hex":
			length = len(str) / 2
		case "base64", "base64url":
			decoded, err := base64.StdEncoding.DecodeString(str)
			if err == nil {
				length = len(decoded)
			}
		}
		return Value{Type: ir.TypeNumber, Number: float64(length)}, nil

	case "__buffer.toString":
		if len(args) < 6 {
			return Value{}, fmt.Errorf("__buffer.toString requires 6 arguments")
		}
		bufVal := env[args[0]]
		enc := strings.ToLower(env[args[1]].String)
		start := int(env[args[2]].Number)
		end := int(env[args[3]].Number)
		hasStart := env[args[4]].Bool
		hasEnd := env[args[5]].Bool

		if bufVal.TypedArray == nil || bufVal.TypedArray.Buffer == nil {
			return Value{Type: ir.TypeString, String: ""}, nil
		}
		ta := bufVal.TypedArray
		fullLen := ta.Length
		if !hasStart || start < 0 {
			start = 0
		}
		if !hasEnd || end > fullLen {
			end = fullLen
		}
		if start > fullLen {
			start = fullLen
		}
		if end < start {
			end = start
		}

		offset := ta.ByteOffset + start
		slice := ta.Buffer.Data[offset : offset+(end-start)]

		var resultStr string
		switch enc {
		case "hex":
			resultStr = hex.EncodeToString(slice)
		case "base64":
			resultStr = base64.StdEncoding.EncodeToString(slice)
		case "ascii":
			var sb strings.Builder
			for _, b := range slice {
				sb.WriteByte(b & 0x7F)
			}
			resultStr = sb.String()
		default: // utf8, latin1, etc.
			resultStr = string(slice)
		}

		return Value{Type: ir.TypeString, String: resultStr}, nil

	case "__buffer.copy":
		if len(args) < 8 {
			return Value{}, fmt.Errorf("__buffer.copy requires 8 arguments")
		}
		srcVal := env[args[0]]
		dstVal := env[args[1]]
		targetStart := int(env[args[2]].Number)
		sourceStart := int(env[args[3]].Number)
		sourceEnd := int(env[args[4]].Number)
		hasTS := env[args[5]].Bool
		hasSS := env[args[6]].Bool
		hasSE := env[args[7]].Bool

		if srcVal.TypedArray == nil || dstVal.TypedArray == nil {
			return Value{Type: ir.TypeNumber, Number: 0}, nil
		}
		src := srcVal.TypedArray
		dst := dstVal.TypedArray

		tstart := 0
		if hasTS && targetStart > 0 {
			tstart = targetStart
		}
		sstart := 0
		if hasSS && sourceStart > 0 {
			sstart = sourceStart
		}
		send := src.Length
		if hasSE && sourceEnd < send {
			send = sourceEnd
		}

		if sstart >= send || tstart >= dst.Length {
			return Value{Type: ir.TypeNumber, Number: 0}, nil
		}
		toCopy := send - sstart
		if tstart+toCopy > dst.Length {
			toCopy = dst.Length - tstart
		}

		if toCopy > 0 && src.Buffer != nil && dst.Buffer != nil {
			copy(dst.Buffer.Data[dst.ByteOffset+tstart:], src.Buffer.Data[src.ByteOffset+sstart:src.ByteOffset+sstart+toCopy])
		}
		return Value{Type: ir.TypeNumber, Number: float64(toCopy)}, nil

	case "__buffer.equals":
		if len(args) < 2 {
			return Value{}, fmt.Errorf("__buffer.equals requires 2 arguments")
		}
		a := env[args[0]].TypedArray
		b := env[args[1]].TypedArray
		if a == nil || b == nil || a.Buffer == nil || b.Buffer == nil {
			return Value{Type: ir.TypeBool, Bool: false}, nil
		}
		aData := a.Buffer.Data[a.ByteOffset : a.ByteOffset+a.Length]
		bData := b.Buffer.Data[b.ByteOffset : b.ByteOffset+b.Length]
		return Value{Type: ir.TypeBool, Bool: bytes.Equal(aData, bData)}, nil

	case "__buffer.compare":
		if len(args) < 2 {
			return Value{}, fmt.Errorf("__buffer.compare requires 2 arguments")
		}
		a := env[args[0]].TypedArray
		b := env[args[1]].TypedArray
		if a == nil || b == nil || a.Buffer == nil || b.Buffer == nil {
			return Value{Type: ir.TypeNumber, Number: 0}, nil
		}
		aData := a.Buffer.Data[a.ByteOffset : a.ByteOffset+a.Length]
		bData := b.Buffer.Data[b.ByteOffset : b.ByteOffset+b.Length]
		return Value{Type: ir.TypeNumber, Number: float64(bytes.Compare(aData, bData))}, nil

	case "__buffer.indexOf":
		if len(args) < 5 {
			return Value{}, fmt.Errorf("__buffer.indexOf requires 5 arguments")
		}
		buf := env[args[0]].TypedArray
		val := env[args[1]]
		isStr := env[args[2]].Bool
		offset := int(env[args[3]].Number)
		hasOffset := env[args[4]].Bool

		if buf == nil || buf.Buffer == nil {
			return Value{Type: ir.TypeNumber, Number: -1}, nil
		}
		if !hasOffset || offset < 0 {
			offset = 0
		}
		if offset >= buf.Length {
			return Value{Type: ir.TypeNumber, Number: -1}, nil
		}
		data := buf.Buffer.Data[buf.ByteOffset+offset : buf.ByteOffset+buf.Length]

		idx := -1
		if isStr {
			target := []byte(val.String)
			i := bytes.Index(data, target)
			if i >= 0 {
				idx = offset + i
			}
		} else {
			target := byte(uint32(val.Number) & 0xFF)
			i := bytes.IndexByte(data, target)
			if i >= 0 {
				idx = offset + i
			}
		}
		return Value{Type: ir.TypeNumber, Number: float64(idx)}, nil

	// Binary Read operations
	case "__buffer.readUInt8":
		p, err := bufferSlice(env, args, 1)
		if err != nil {
			return Value{}, err
		}
		return Value{Type: ir.TypeNumber, Number: float64(p[0])}, nil

	case "__buffer.readInt8":
		p, err := bufferSlice(env, args, 1)
		if err != nil {
			return Value{}, err
		}
		return Value{Type: ir.TypeNumber, Number: float64(int8(p[0]))}, nil

	case "__buffer.readUInt16LE":
		p, err := bufferSlice(env, args, 2)
		if err != nil {
			return Value{}, err
		}
		return Value{Type: ir.TypeNumber, Number: float64(encbinary.LittleEndian.Uint16(p))}, nil

	case "__buffer.readUInt16BE":
		p, err := bufferSlice(env, args, 2)
		if err != nil {
			return Value{}, err
		}
		return Value{Type: ir.TypeNumber, Number: float64(encbinary.BigEndian.Uint16(p))}, nil

	case "__buffer.readInt16LE":
		p, err := bufferSlice(env, args, 2)
		if err != nil {
			return Value{}, err
		}
		return Value{Type: ir.TypeNumber, Number: float64(int16(encbinary.LittleEndian.Uint16(p)))}, nil

	case "__buffer.readInt16BE":
		p, err := bufferSlice(env, args, 2)
		if err != nil {
			return Value{}, err
		}
		return Value{Type: ir.TypeNumber, Number: float64(int16(encbinary.BigEndian.Uint16(p)))}, nil

	case "__buffer.readUInt32LE":
		p, err := bufferSlice(env, args, 4)
		if err != nil {
			return Value{}, err
		}
		return Value{Type: ir.TypeNumber, Number: float64(encbinary.LittleEndian.Uint32(p))}, nil

	case "__buffer.readUInt32BE":
		p, err := bufferSlice(env, args, 4)
		if err != nil {
			return Value{}, err
		}
		return Value{Type: ir.TypeNumber, Number: float64(encbinary.BigEndian.Uint32(p))}, nil

	case "__buffer.readInt32LE":
		p, err := bufferSlice(env, args, 4)
		if err != nil {
			return Value{}, err
		}
		return Value{Type: ir.TypeNumber, Number: float64(int32(encbinary.LittleEndian.Uint32(p)))}, nil

	case "__buffer.readInt32BE":
		p, err := bufferSlice(env, args, 4)
		if err != nil {
			return Value{}, err
		}
		return Value{Type: ir.TypeNumber, Number: float64(int32(encbinary.BigEndian.Uint32(p)))}, nil

	case "__buffer.readFloatLE":
		p, err := bufferSlice(env, args, 4)
		if err != nil {
			return Value{}, err
		}
		u := encbinary.LittleEndian.Uint32(p)
		return Value{Type: ir.TypeNumber, Number: float64(math.Float32frombits(u))}, nil

	case "__buffer.readFloatBE":
		p, err := bufferSlice(env, args, 4)
		if err != nil {
			return Value{}, err
		}
		u := encbinary.BigEndian.Uint32(p)
		return Value{Type: ir.TypeNumber, Number: float64(math.Float32frombits(u))}, nil

	case "__buffer.readDoubleLE":
		p, err := bufferSlice(env, args, 8)
		if err != nil {
			return Value{}, err
		}
		u := encbinary.LittleEndian.Uint64(p)
		return Value{Type: ir.TypeNumber, Number: math.Float64frombits(u)}, nil

	case "__buffer.readDoubleBE":
		p, err := bufferSlice(env, args, 8)
		if err != nil {
			return Value{}, err
		}
		u := encbinary.BigEndian.Uint64(p)
		return Value{Type: ir.TypeNumber, Number: math.Float64frombits(u)}, nil

	// Binary Write operations
	case "__buffer.writeUInt8":
		p, err := bufferSlice(env, []string{args[0], args[2]}, 1)
		if err != nil {
			return Value{}, err
		}
		p[0] = byte(uint32(env[args[1]].Number) & 0xFF)
		return Value{Type: ir.TypeNumber, Number: env[args[2]].Number + 1}, nil

	case "__buffer.writeInt8":
		p, err := bufferSlice(env, []string{args[0], args[2]}, 1)
		if err != nil {
			return Value{}, err
		}
		p[0] = byte(int8(int32(env[args[1]].Number)))
		return Value{Type: ir.TypeNumber, Number: env[args[2]].Number + 1}, nil

	case "__buffer.writeUInt16LE":
		p, err := bufferSlice(env, []string{args[0], args[2]}, 2)
		if err != nil {
			return Value{}, err
		}
		encbinary.LittleEndian.PutUint16(p, uint16(uint32(env[args[1]].Number)))
		return Value{Type: ir.TypeNumber, Number: env[args[2]].Number + 2}, nil

	case "__buffer.writeUInt16BE":
		p, err := bufferSlice(env, []string{args[0], args[2]}, 2)
		if err != nil {
			return Value{}, err
		}
		encbinary.BigEndian.PutUint16(p, uint16(uint32(env[args[1]].Number)))
		return Value{Type: ir.TypeNumber, Number: env[args[2]].Number + 2}, nil

	case "__buffer.writeInt16LE":
		p, err := bufferSlice(env, []string{args[0], args[2]}, 2)
		if err != nil {
			return Value{}, err
		}
		encbinary.LittleEndian.PutUint16(p, uint16(int16(int32(env[args[1]].Number))))
		return Value{Type: ir.TypeNumber, Number: env[args[2]].Number + 2}, nil

	case "__buffer.writeInt16BE":
		p, err := bufferSlice(env, []string{args[0], args[2]}, 2)
		if err != nil {
			return Value{}, err
		}
		encbinary.BigEndian.PutUint16(p, uint16(int16(int32(env[args[1]].Number))))
		return Value{Type: ir.TypeNumber, Number: env[args[2]].Number + 2}, nil

	case "__buffer.writeUInt32LE":
		p, err := bufferSlice(env, []string{args[0], args[2]}, 4)
		if err != nil {
			return Value{}, err
		}
		encbinary.LittleEndian.PutUint32(p, uint32(env[args[1]].Number))
		return Value{Type: ir.TypeNumber, Number: env[args[2]].Number + 4}, nil

	case "__buffer.writeUInt32BE":
		p, err := bufferSlice(env, []string{args[0], args[2]}, 4)
		if err != nil {
			return Value{}, err
		}
		encbinary.BigEndian.PutUint32(p, uint32(env[args[1]].Number))
		return Value{Type: ir.TypeNumber, Number: env[args[2]].Number + 4}, nil

	case "__buffer.writeInt32LE":
		p, err := bufferSlice(env, []string{args[0], args[2]}, 4)
		if err != nil {
			return Value{}, err
		}
		encbinary.LittleEndian.PutUint32(p, uint32(int32(env[args[1]].Number)))
		return Value{Type: ir.TypeNumber, Number: env[args[2]].Number + 4}, nil

	case "__buffer.writeInt32BE":
		p, err := bufferSlice(env, []string{args[0], args[2]}, 4)
		if err != nil {
			return Value{}, err
		}
		encbinary.BigEndian.PutUint32(p, uint32(int32(env[args[1]].Number)))
		return Value{Type: ir.TypeNumber, Number: env[args[2]].Number + 4}, nil

	case "__buffer.writeFloatLE":
		p, err := bufferSlice(env, []string{args[0], args[2]}, 4)
		if err != nil {
			return Value{}, err
		}
		encbinary.LittleEndian.PutUint32(p, math.Float32bits(float32(env[args[1]].Number)))
		return Value{Type: ir.TypeNumber, Number: env[args[2]].Number + 4}, nil

	case "__buffer.writeFloatBE":
		p, err := bufferSlice(env, []string{args[0], args[2]}, 4)
		if err != nil {
			return Value{}, err
		}
		encbinary.BigEndian.PutUint32(p, math.Float32bits(float32(env[args[1]].Number)))
		return Value{Type: ir.TypeNumber, Number: env[args[2]].Number + 4}, nil

	case "__buffer.writeDoubleLE":
		p, err := bufferSlice(env, []string{args[0], args[2]}, 8)
		if err != nil {
			return Value{}, err
		}
		encbinary.LittleEndian.PutUint64(p, math.Float64bits(env[args[1]].Number))
		return Value{Type: ir.TypeNumber, Number: env[args[2]].Number + 8}, nil

	case "__buffer.writeDoubleBE":
		p, err := bufferSlice(env, []string{args[0], args[2]}, 8)
		if err != nil {
			return Value{}, err
		}
		encbinary.BigEndian.PutUint64(p, math.Float64bits(env[args[1]].Number))
		return Value{Type: ir.TypeNumber, Number: env[args[2]].Number + 8}, nil

	default:
		return Value{}, fmt.Errorf("unknown buffer intrinsic %q", callee)
	}
}

func bufferSlice(env map[string]Value, args []string, size int) ([]byte, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("buffer method requires target and offset")
	}
	ta := env[args[0]].TypedArray
	if ta == nil || ta.Buffer == nil {
		return nil, fmt.Errorf("buffer target is invalid")
	}
	offset := int(env[args[1]].Number)
	if offset < 0 || offset+size > ta.Length || ta.ByteOffset+offset+size > len(ta.Buffer.Data) {
		return nil, fmt.Errorf("buffer offset %d out of bounds (length %d)", offset, ta.Length)
	}
	start := ta.ByteOffset + offset
	return ta.Buffer.Data[start : start+size], nil
}

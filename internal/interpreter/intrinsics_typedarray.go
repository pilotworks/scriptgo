package interpreter

import (
	encbinary "encoding/binary"
	"fmt"
	"math"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func executeTypedArrayIntrinsic(inst ir.Instruction, env map[string]Value) (Value, error) {
	callee := inst.Callee
	args := inst.Args
	kindVal := inst.Type

	switch callee {
	case "__arraybuffer.new":
		if len(args) != 1 {
			return Value{}, fmt.Errorf("ArrayBuffer constructor requires 1 argument")
		}
		lenVal, ok := env[args[0]]
		if !ok || lenVal.Type != ir.TypeNumber || lenVal.Number < 0 {
			return Value{}, fmt.Errorf("ArrayBuffer byteLength must be a non-negative number")
		}
		byteLen := int(lenVal.Number)
		return Value{
			Type:   ir.TypeArrayBuffer,
			Buffer: &ArrayBuffer{Data: make([]byte, byteLen)},
		}, nil

	case "__arraybuffer.byteLength":
		if len(args) != 1 {
			return Value{}, fmt.Errorf("ArrayBuffer.byteLength requires 1 argument")
		}
		bufVal, ok := env[args[0]]
		if !ok || bufVal.Buffer == nil {
			return Value{}, fmt.Errorf("ArrayBuffer.byteLength requires an ArrayBuffer")
		}
		return Value{Type: ir.TypeNumber, Number: float64(len(bufVal.Buffer.Data))}, nil

	case "__arraybuffer.slice":
		if len(args) < 1 {
			return Value{}, fmt.Errorf("ArrayBuffer.slice requires at least 1 argument")
		}
		bufVal, ok := env[args[0]]
		if !ok || bufVal.Buffer == nil {
			return Value{}, fmt.Errorf("ArrayBuffer.slice requires an ArrayBuffer")
		}
		data := bufVal.Buffer.Data
		dataLen := len(data)
		startIdx := 0
		endIdx := dataLen
		if len(args) > 1 {
			sVal := env[args[1]]
			if sVal.Type == ir.TypeNumber {
				startIdx = int(sVal.Number)
				if startIdx < 0 {
					startIdx += dataLen
					if startIdx < 0 {
						startIdx = 0
					}
				} else if startIdx > dataLen {
					startIdx = dataLen
				}
			}
		}
		if len(args) > 2 {
			eVal := env[args[2]]
			if eVal.Type == ir.TypeNumber {
				endIdx = int(eVal.Number)
				if endIdx < 0 {
					endIdx += dataLen
					if endIdx < 0 {
						endIdx = 0
					}
				} else if endIdx > dataLen {
					endIdx = dataLen
				}
			}
		}
		newLen := 0
		if endIdx > startIdx {
			newLen = endIdx - startIdx
		}
		newData := make([]byte, newLen)
		if newLen > 0 {
			copy(newData, data[startIdx:endIdx])
		}
		return Value{
			Type:   ir.TypeArrayBuffer,
			Buffer: &ArrayBuffer{Data: newData},
		}, nil

	case "__arraybuffer.isView":
		if len(args) != 1 {
			return Value{}, fmt.Errorf("ArrayBuffer.isView requires 1 argument")
		}
		argVal, ok := env[args[0]]
		return Value{Type: ir.TypeBool, Bool: ok && (argVal.TypedArray != nil || argVal.DataView != nil)}, nil

	case "__typedarray.new_length":
		if len(args) != 1 {
			return Value{}, fmt.Errorf("__typedarray.new_length requires 1 argument")
		}
		lenVal := int(env[args[0]].Number)
		if lenVal < 0 {
			return Value{}, fmt.Errorf("TypedArray length must be non-negative")
		}
		ta := &TypedArray{
			Kind:   kindVal,
			Length: lenVal,
		}
		totalBytes := lenVal * ta.ElementSize()
		ta.Buffer = &ArrayBuffer{Data: make([]byte, totalBytes)}
		return Value{Type: kindVal, TypedArray: ta}, nil

	case "__typedarray.new_buffer":
		if len(args) < 1 {
			return Value{}, fmt.Errorf("__typedarray.new_buffer requires at least buffer argument")
		}
		bufVal := env[args[0]]
		if bufVal.Buffer == nil {
			return Value{}, fmt.Errorf("TypedArray buffer constructor requires an ArrayBuffer")
		}
		byteOffset := 0
		if len(args) > 1 {
			byteOffset = int(env[args[1]].Number)
		}
		ta := &TypedArray{
			Kind:       kindVal,
			Buffer:     bufVal.Buffer,
			ByteOffset: byteOffset,
		}
		if len(args) > 2 {
			l := int(env[args[2]].Number)
			if l > 0 {
				ta.Length = l
			} else {
				available := len(bufVal.Buffer.Data) - byteOffset
				if available < 0 {
					available = 0
				}
				ta.Length = available / ta.ElementSize()
			}
		} else {
			available := len(bufVal.Buffer.Data) - byteOffset
			if available < 0 {
				available = 0
			}
			ta.Length = available / ta.ElementSize()
		}
		return Value{Type: kindVal, TypedArray: ta}, nil

	case "__typedarray.new_array":
		if len(args) != 1 {
			return Value{}, fmt.Errorf("__typedarray.new_array requires 1 argument")
		}
		src := env[args[0]]
		var lenVal int
		if src.TypedArray != nil {
			lenVal = src.TypedArray.Length
		} else {
			lenVal = len(src.GetArray())
		}
		ta := &TypedArray{
			Kind:   kindVal,
			Length: lenVal,
		}
		totalBytes := lenVal * ta.ElementSize()
		ta.Buffer = &ArrayBuffer{Data: make([]byte, totalBytes)}
		if src.TypedArray != nil {
			for i := 0; i < lenVal; i++ {
				if ta.Kind == ir.TypeBigInt64Array || ta.Kind == ir.TypeBigUint64Array {
					ta.SetBigInt(i, src.TypedArray.GetBigInt(i))
				} else {
					ta.Set(i, src.TypedArray.Get(i))
				}
			}
		} else {
			arr := src.GetArray()
			for i, it := range arr {
				if ta.Kind == ir.TypeBigInt64Array || ta.Kind == ir.TypeBigUint64Array {
					ta.SetBigInt(i, it.BigInt)
				} else {
					ta.Set(i, it.Number)
				}
			}
		}
		return Value{Type: kindVal, TypedArray: ta}, nil

	case "__typedarray.length":
		if len(args) != 1 {
			return Value{}, fmt.Errorf("TypedArray.length requires 1 argument")
		}
		ta := env[args[0]].TypedArray
		if ta == nil {
			return Value{}, fmt.Errorf("TypedArray.length requires a TypedArray")
		}
		return Value{Type: ir.TypeNumber, Number: float64(ta.Length)}, nil

	case "__typedarray.byteLength":
		if len(args) != 1 {
			return Value{}, fmt.Errorf("TypedArray.byteLength requires 1 argument")
		}
		ta := env[args[0]].TypedArray
		if ta == nil {
			return Value{}, fmt.Errorf("TypedArray.byteLength requires a TypedArray")
		}
		return Value{Type: ir.TypeNumber, Number: float64(ta.ByteLength())}, nil

	case "__typedarray.byteOffset":
		if len(args) != 1 {
			return Value{}, fmt.Errorf("TypedArray.byteOffset requires 1 argument")
		}
		ta := env[args[0]].TypedArray
		if ta == nil {
			return Value{}, fmt.Errorf("TypedArray.byteOffset requires a TypedArray")
		}
		return Value{Type: ir.TypeNumber, Number: float64(ta.ByteOffset)}, nil

	case "__typedarray.buffer":
		if len(args) != 1 {
			return Value{}, fmt.Errorf("TypedArray.buffer requires 1 argument")
		}
		ta := env[args[0]].TypedArray
		if ta == nil {
			return Value{}, fmt.Errorf("TypedArray.buffer requires a TypedArray")
		}
		return Value{Type: ir.TypeArrayBuffer, Buffer: ta.Buffer}, nil

	case "__typedarray.subarray":
		if len(args) < 1 {
			return Value{}, fmt.Errorf("TypedArray.subarray requires at least 1 argument")
		}
		ta := env[args[0]].TypedArray
		if ta == nil {
			return Value{}, fmt.Errorf("TypedArray.subarray requires a TypedArray")
		}
		startIdx := 0
		endIdx := ta.Length
		if len(args) > 1 {
			s := int(env[args[1]].Number)
			if s < 0 {
				s += ta.Length
				if s < 0 {
					s = 0
				}
			} else if s > ta.Length {
				s = ta.Length
			}
			startIdx = s
		}
		if len(args) > 2 {
			e := int(env[args[2]].Number)
			if e < 0 {
				e += ta.Length
				if e < 0 {
					e = 0
				}
			} else if e > ta.Length {
				e = ta.Length
			}
			endIdx = e
		}
		subLen := 0
		if endIdx > startIdx {
			subLen = endIdx - startIdx
		}
		newOffset := ta.ByteOffset + startIdx*ta.ElementSize()
		sub := &TypedArray{
			Kind:       ta.Kind,
			Buffer:     ta.Buffer,
			ByteOffset: newOffset,
			Length:     subLen,
		}
		return Value{Type: ta.Kind, TypedArray: sub}, nil

	case "__typedarray.slice":
		if len(args) < 1 {
			return Value{}, fmt.Errorf("TypedArray.slice requires at least 1 argument")
		}
		ta := env[args[0]].TypedArray
		if ta == nil {
			return Value{}, fmt.Errorf("TypedArray.slice requires a TypedArray")
		}
		startIdx := 0
		endIdx := ta.Length
		if len(args) > 1 {
			s := int(env[args[1]].Number)
			if s < 0 {
				s += ta.Length
				if s < 0 {
					s = 0
				}
			} else if s > ta.Length {
				s = ta.Length
			}
			startIdx = s
		}
		if len(args) > 2 {
			e := int(env[args[2]].Number)
			if e < 0 {
				e += ta.Length
				if e < 0 {
					e = 0
				}
			} else if e > ta.Length {
				e = ta.Length
			}
			endIdx = e
		}
		sliceLen := 0
		if endIdx > startIdx {
			sliceLen = endIdx - startIdx
		}
		newTa := &TypedArray{
			Kind:   ta.Kind,
			Length: sliceLen,
			Buffer: &ArrayBuffer{Data: make([]byte, sliceLen*ta.ElementSize())},
		}
		for i := 0; i < sliceLen; i++ {
			if ta.Kind == ir.TypeBigInt64Array || ta.Kind == ir.TypeBigUint64Array {
				newTa.SetBigInt(i, ta.GetBigInt(startIdx+i))
			} else {
				newTa.Set(i, ta.Get(startIdx+i))
			}
		}
		return Value{Type: ta.Kind, TypedArray: newTa}, nil

	case "__typedarray.set":
		if len(args) < 2 {
			return Value{}, fmt.Errorf("TypedArray.set requires target and source")
		}
		target := env[args[0]].TypedArray
		if target == nil {
			return Value{}, fmt.Errorf("TypedArray.set requires a target TypedArray")
		}
		offset := 0
		if len(args) > 2 {
			offset = int(env[args[2]].Number)
		}
		src := env[args[1]]
		if src.TypedArray != nil {
			for i := 0; i < src.TypedArray.Length; i++ {
				if target.Kind == ir.TypeBigInt64Array || target.Kind == ir.TypeBigUint64Array {
					target.SetBigInt(offset+i, src.TypedArray.GetBigInt(i))
				} else {
					target.Set(offset+i, src.TypedArray.Get(i))
				}
			}
		} else {
			arr := src.GetArray()
			for i, it := range arr {
				if target.Kind == ir.TypeBigInt64Array || target.Kind == ir.TypeBigUint64Array {
					target.SetBigInt(offset+i, it.BigInt)
				} else {
					target.Set(offset+i, it.Number)
				}
			}
		}
		return Value{Type: ir.TypeVoid}, nil

	case "__typedarray.fill":
		if len(args) < 2 {
			return Value{}, fmt.Errorf("TypedArray.fill requires target and value")
		}
		ta := env[args[0]].TypedArray
		if ta == nil {
			return Value{}, fmt.Errorf("TypedArray.fill requires a TypedArray")
		}
		val := env[args[1]].Number
		bigVal := env[args[1]].BigInt
		startIdx := 0
		endIdx := ta.Length
		if len(args) > 2 {
			s := int(env[args[2]].Number)
			if s < 0 {
				s += ta.Length
				if s < 0 {
					s = 0
				}
			} else if s > ta.Length {
				s = ta.Length
			}
			startIdx = s
		}
		if len(args) > 3 {
			e := int(env[args[3]].Number)
			if e < 0 {
				e += ta.Length
				if e < 0 {
					e = 0
				}
			} else if e > ta.Length {
				e = ta.Length
			}
			endIdx = e
		}
		for i := startIdx; i < endIdx; i++ {
			if ta.Kind == ir.TypeBigInt64Array || ta.Kind == ir.TypeBigUint64Array {
				ta.SetBigInt(i, bigVal)
			} else {
				ta.Set(i, val)
			}
		}
		return env[args[0]], nil

	case "__typedarray.toString":
		if len(args) != 1 {
			return Value{}, fmt.Errorf("TypedArray.toString requires 1 argument")
		}
		taVal := env[args[0]]
		return Value{Type: ir.TypeString, String: format(taVal)}, nil

	case "__arraybuffer.toString":
		if len(args) != 1 {
			return Value{}, fmt.Errorf("ArrayBuffer.toString requires 1 argument")
		}
		bufVal := env[args[0]]
		return Value{Type: ir.TypeString, String: format(bufVal)}, nil

	// -------------------------------------------------------------------------
	// DataView intrinsics
	// -------------------------------------------------------------------------
	case "__dataview.new":
		if len(args) < 1 {
			return Value{}, fmt.Errorf("DataView constructor requires at least 1 argument")
		}
		bufVal := env[args[0]]
		if bufVal.Buffer == nil {
			return Value{}, fmt.Errorf("DataView constructor requires an ArrayBuffer")
		}
		byteOffset := 0
		if len(args) > 1 {
			byteOffset = int(env[args[1]].Number)
		}
		if byteOffset < 0 || byteOffset > len(bufVal.Buffer.Data) {
			return Value{}, fmt.Errorf("DataView byteOffset out of bounds")
		}
		byteLength := len(bufVal.Buffer.Data) - byteOffset
		if len(args) > 2 {
			l := int(env[args[2]].Number)
			if l > 0 {
				byteLength = l
			}
		}
		if byteLength < 0 || byteOffset+byteLength > len(bufVal.Buffer.Data) {
			return Value{}, fmt.Errorf("DataView byteLength out of bounds")
		}
		dv := &DataView{
			Buffer:     bufVal.Buffer,
			ByteOffset: byteOffset,
			ByteLength: byteLength,
		}
		return Value{Type: ir.TypeDataView, DataView: dv}, nil

	case "__dataview.byteLength":
		if len(args) != 1 {
			return Value{}, fmt.Errorf("DataView.byteLength requires 1 argument")
		}
		dv := env[args[0]].DataView
		if dv == nil {
			return Value{}, fmt.Errorf("DataView.byteLength requires a DataView")
		}
		return Value{Type: ir.TypeNumber, Number: float64(dv.ByteLength)}, nil

	case "__dataview.byteOffset":
		if len(args) != 1 {
			return Value{}, fmt.Errorf("DataView.byteOffset requires 1 argument")
		}
		dv := env[args[0]].DataView
		if dv == nil {
			return Value{}, fmt.Errorf("DataView.byteOffset requires a DataView")
		}
		return Value{Type: ir.TypeNumber, Number: float64(dv.ByteOffset)}, nil

	case "__dataview.buffer":
		if len(args) != 1 {
			return Value{}, fmt.Errorf("DataView.buffer requires 1 argument")
		}
		dv := env[args[0]].DataView
		if dv == nil {
			return Value{}, fmt.Errorf("DataView.buffer requires a DataView")
		}
		return Value{Type: ir.TypeArrayBuffer, Buffer: dv.Buffer}, nil

	case "__dataview.getInt8":
		p, err := dataviewSlice(env, args, 1)
		if err != nil {
			return Value{}, err
		}
		return Value{Type: ir.TypeNumber, Number: float64(int8(p[0]))}, nil

	case "__dataview.setUint8":
		p, err := dataviewSlice(env, args, 1)
		if err != nil {
			return Value{}, err
		}
		p[0] = byte(uint32(env[args[2]].Number) & 0xff)
		return Value{Type: ir.TypeVoid}, nil

	case "__dataview.getUint8":
		p, err := dataviewSlice(env, args, 1)
		if err != nil {
			return Value{}, err
		}
		return Value{Type: ir.TypeNumber, Number: float64(p[0])}, nil

	case "__dataview.setInt8":
		p, err := dataviewSlice(env, args, 1)
		if err != nil {
			return Value{}, err
		}
		p[0] = byte(int8(int32(env[args[2]].Number)))
		return Value{Type: ir.TypeVoid}, nil

	case "__dataview.getInt16":
		p, err := dataviewSlice(env, args, 2)
		if err != nil {
			return Value{}, err
		}
		order := getByteOrder(env, args, 2)
		return Value{Type: ir.TypeNumber, Number: float64(int16(order.Uint16(p)))}, nil

	case "__dataview.setUint16":
		p, err := dataviewSlice(env, args, 2)
		if err != nil {
			return Value{}, err
		}
		order := getByteOrder(env, args, 3)
		order.PutUint16(p, uint16(uint32(env[args[2]].Number)))
		return Value{Type: ir.TypeVoid}, nil

	case "__dataview.getUint16":
		p, err := dataviewSlice(env, args, 2)
		if err != nil {
			return Value{}, err
		}
		order := getByteOrder(env, args, 2)
		return Value{Type: ir.TypeNumber, Number: float64(order.Uint16(p))}, nil

	case "__dataview.setInt16":
		p, err := dataviewSlice(env, args, 2)
		if err != nil {
			return Value{}, err
		}
		order := getByteOrder(env, args, 3)
		order.PutUint16(p, uint16(int16(int32(env[args[2]].Number))))
		return Value{Type: ir.TypeVoid}, nil

	case "__dataview.getInt32":
		p, err := dataviewSlice(env, args, 4)
		if err != nil {
			return Value{}, err
		}
		order := getByteOrder(env, args, 2)
		return Value{Type: ir.TypeNumber, Number: float64(int32(order.Uint32(p)))}, nil

	case "__dataview.setUint32":
		p, err := dataviewSlice(env, args, 4)
		if err != nil {
			return Value{}, err
		}
		order := getByteOrder(env, args, 3)
		order.PutUint32(p, uint32(env[args[2]].Number))
		return Value{Type: ir.TypeVoid}, nil

	case "__dataview.getUint32":
		p, err := dataviewSlice(env, args, 4)
		if err != nil {
			return Value{}, err
		}
		order := getByteOrder(env, args, 2)
		return Value{Type: ir.TypeNumber, Number: float64(order.Uint32(p))}, nil

	case "__dataview.setInt32":
		p, err := dataviewSlice(env, args, 4)
		if err != nil {
			return Value{}, err
		}
		order := getByteOrder(env, args, 3)
		order.PutUint32(p, uint32(int32(env[args[2]].Number)))
		return Value{Type: ir.TypeVoid}, nil

	case "__dataview.getFloat32":
		p, err := dataviewSlice(env, args, 4)
		if err != nil {
			return Value{}, err
		}
		order := getByteOrder(env, args, 2)
		u := order.Uint32(p)
		return Value{Type: ir.TypeNumber, Number: float64(math.Float32frombits(u))}, nil

	case "__dataview.setFloat32":
		p, err := dataviewSlice(env, args, 4)
		if err != nil {
			return Value{}, err
		}
		order := getByteOrder(env, args, 3)
		order.PutUint32(p, math.Float32bits(float32(env[args[2]].Number)))
		return Value{Type: ir.TypeVoid}, nil

	case "__dataview.getFloat64":
		p, err := dataviewSlice(env, args, 8)
		if err != nil {
			return Value{}, err
		}
		order := getByteOrder(env, args, 2)
		u := order.Uint64(p)
		return Value{Type: ir.TypeNumber, Number: math.Float64frombits(u)}, nil

	case "__dataview.setFloat64":
		p, err := dataviewSlice(env, args, 8)
		if err != nil {
			return Value{}, err
		}
		order := getByteOrder(env, args, 3)
		order.PutUint64(p, math.Float64bits(env[args[2]].Number))
		return Value{Type: ir.TypeVoid}, nil

	case "__dataview.getBigInt64", "__dataview.getBigUint64":
		p, err := dataviewSlice(env, args, 8)
		if err != nil {
			return Value{}, err
		}
		order := getByteOrder(env, args, 2)
		return Value{Type: ir.TypeBigInt, BigInt: int64(order.Uint64(p))}, nil

	case "__dataview.setBigInt64", "__dataview.setBigUint64":
		p, err := dataviewSlice(env, args, 8)
		if err != nil {
			return Value{}, err
		}
		order := getByteOrder(env, args, 3)
		order.PutUint64(p, uint64(env[args[2]].BigInt))
		return Value{Type: ir.TypeVoid}, nil

	case "__dataview.toString":
		if len(args) != 1 {
			return Value{}, fmt.Errorf("DataView.toString requires 1 argument")
		}
		dvVal := env[args[0]]
		return Value{Type: ir.TypeString, String: format(dvVal)}, nil

	default:
		return Value{}, fmt.Errorf("unknown typedarray intrinsic %q", callee)
	}
}

func dataviewSlice(env map[string]Value, args []string, size int) ([]byte, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("DataView method requires DataView target")
	}
	dv := env[args[0]].DataView
	if dv == nil || dv.Buffer == nil {
		return nil, fmt.Errorf("DataView method target is invalid")
	}
	byteOffset := 0
	if len(args) > 1 {
		byteOffset = int(env[args[1]].Number)
	}
	if byteOffset < 0 || byteOffset+size > dv.ByteLength || dv.ByteOffset+byteOffset+size > len(dv.Buffer.Data) {
		return nil, fmt.Errorf("DataView offset %d out of bounds (length %d)", byteOffset, dv.ByteLength)
	}
	start := dv.ByteOffset + byteOffset
	return dv.Buffer.Data[start : start+size], nil
}

func getByteOrder(env map[string]Value, args []string, idx int) encbinary.ByteOrder {
	if len(args) > idx {
		leVal := env[args[idx]]
		if (leVal.Type == ir.TypeBool && leVal.Bool) || (leVal.Type == ir.TypeNumber && leVal.Number != 0) {
			return encbinary.LittleEndian
		}
	}
	return encbinary.BigEndian
}

package interpreter

import (
	"fmt"

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
		return Value{Type: ir.TypeBool, Bool: ok && argVal.TypedArray != nil}, nil

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
		srcArr := env[args[0]]
		var elements []float64
		if srcArr.TypedArray != nil {
			for i := 0; i < srcArr.TypedArray.Length; i++ {
				elements = append(elements, srcArr.TypedArray.Get(i))
			}
		} else {
			arr := srcArr.GetArray()
			for _, item := range arr {
				elements = append(elements, item.Number)
			}
		}
		ta := &TypedArray{
			Kind:   kindVal,
			Length: len(elements),
		}
		ta.Buffer = &ArrayBuffer{Data: make([]byte, len(elements)*ta.ElementSize())}
		for i, el := range elements {
			ta.Set(i, el)
		}
		return Value{Type: kindVal, TypedArray: ta}, nil

	case "__typedarray.length":
		if len(args) != 1 {
			return Value{}, fmt.Errorf("TypedArray.length requires 1 argument")
		}
		taVal := env[args[0]]
		if taVal.TypedArray == nil {
			return Value{}, fmt.Errorf("TypedArray.length requires a TypedArray")
		}
		return Value{Type: ir.TypeNumber, Number: float64(taVal.TypedArray.Length)}, nil

	case "__typedarray.byteLength":
		if len(args) != 1 {
			return Value{}, fmt.Errorf("TypedArray.byteLength requires 1 argument")
		}
		taVal := env[args[0]]
		if taVal.TypedArray == nil {
			return Value{}, fmt.Errorf("TypedArray.byteLength requires a TypedArray")
		}
		return Value{Type: ir.TypeNumber, Number: float64(taVal.TypedArray.ByteLength())}, nil

	case "__typedarray.byteOffset":
		if len(args) != 1 {
			return Value{}, fmt.Errorf("TypedArray.byteOffset requires 1 argument")
		}
		taVal := env[args[0]]
		if taVal.TypedArray == nil {
			return Value{}, fmt.Errorf("TypedArray.byteOffset requires a TypedArray")
		}
		return Value{Type: ir.TypeNumber, Number: float64(taVal.TypedArray.ByteOffset)}, nil

	case "__typedarray.buffer":
		if len(args) != 1 {
			return Value{}, fmt.Errorf("TypedArray.buffer requires 1 argument")
		}
		taVal := env[args[0]]
		if taVal.TypedArray == nil {
			return Value{}, fmt.Errorf("TypedArray.buffer requires a TypedArray")
		}
		return Value{Type: ir.TypeArrayBuffer, Buffer: taVal.TypedArray.Buffer}, nil

	case "__typedarray.subarray":
		if len(args) < 1 {
			return Value{}, fmt.Errorf("TypedArray.subarray requires at least 1 argument")
		}
		taVal := env[args[0]]
		if taVal.TypedArray == nil {
			return Value{}, fmt.Errorf("TypedArray.subarray requires a TypedArray")
		}
		ta := taVal.TypedArray
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
		subTa := &TypedArray{
			Kind:       ta.Kind,
			Buffer:     ta.Buffer,
			ByteOffset: newOffset,
			Length:     subLen,
		}
		return Value{Type: ta.Kind, TypedArray: subTa}, nil

	case "__typedarray.slice":
		if len(args) < 1 {
			return Value{}, fmt.Errorf("TypedArray.slice requires at least 1 argument")
		}
		taVal := env[args[0]]
		if taVal.TypedArray == nil {
			return Value{}, fmt.Errorf("TypedArray.slice requires a TypedArray")
		}
		ta := taVal.TypedArray
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
			newTa.Set(i, ta.Get(startIdx+i))
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
				target.Set(offset+i, src.TypedArray.Get(i))
			}
		} else {
			arr := src.GetArray()
			for i, it := range arr {
				target.Set(offset+i, it.Number)
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
			ta.Set(i, val)
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

	default:
		return Value{}, fmt.Errorf("unknown typedarray intrinsic %q", callee)
	}
}

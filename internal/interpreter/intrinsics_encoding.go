package interpreter

import (
	"bytes"
	"fmt"
	"unicode/utf8"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func executeTextEncodingIntrinsic(instruction ir.Instruction, env map[string]Value) (Value, error) {
	switch instruction.Callee {
	case "__text_encoder.new":
		return Value{
			Type: ir.TypeTextEncoder,
			TextEncoder: &TextEncoderValue{
				Encoding: "utf-8",
			},
		}, nil

	case "__text_encoder.encoding":
		return Value{
			Type:   ir.TypeString,
			String: "utf-8",
		}, nil

	case "__text_encoder.encode":
		str := ""
		if len(instruction.Args) > 0 {
			argVal := env[instruction.Args[0]]
			if argVal.Type == ir.TypeString {
				str = argVal.String
			}
		}
		data := []byte(str)
		buf := &ArrayBuffer{Data: data}
		ta := &TypedArray{
			Kind:       ir.TypeUint8Array,
			Buffer:     buf,
			ByteOffset: 0,
			Length:     len(data),
		}
		return Value{
			Type:       ir.TypeUint8Array,
			TypedArray: ta,
		}, nil

	case "__text_encoder.encode_into":
		if len(instruction.Args) < 2 {
			return Value{}, fmt.Errorf("encodeInto requires source and destination arguments")
		}
		sourceVal := env[instruction.Args[0]]
		destVal := env[instruction.Args[1]]

		if destVal.TypedArray == nil || destVal.TypedArray.Buffer == nil {
			return Value{}, fmt.Errorf("encodeInto destination must be a TypedArray")
		}

		sourceStr := sourceVal.String
		dest := destVal.TypedArray
		destLen := dest.Length
		destBuf := dest.Buffer.Data
		destOffset := dest.ByteOffset

		readCodeUnits := 0
		writtenBytes := 0

		r := []rune(sourceStr)
		for _, ru := range r {
			var buf [4]byte
			n := utf8.EncodeRune(buf[:], ru)

			if destOffset+writtenBytes+n > len(destBuf) || writtenBytes+n > destLen {
				break
			}

			copy(destBuf[destOffset+writtenBytes:], buf[:n])
			writtenBytes += n

			if ru > 0xFFFF {
				readCodeUnits += 2
			} else {
				readCodeUnits++
			}
		}

		resObj := map[string]Value{
			"read":    {Type: ir.TypeNumber, Number: float64(readCodeUnits)},
			"written": {Type: ir.TypeNumber, Number: float64(writtenBytes)},
		}

		return Value{
			Type:   ir.Type("object:TextEncoderEncodeIntoResult"),
			Object: resObj,
		}, nil

	case "__text_decoder.new":
		fatal := false
		ignoreBOM := false

		if len(instruction.Args) > 1 {
			optsVal := env[instruction.Args[1]]
			if optsVal.Object != nil {
				if f, ok := optsVal.Object["fatal"]; ok && f.Bool {
					fatal = true
				}
				if b, ok := optsVal.Object["ignoreBOM"]; ok && b.Bool {
					ignoreBOM = true
				}
			}
		}

		return Value{
			Type: ir.TypeTextDecoder,
			TextDecoder: &TextDecoderValue{
				Encoding:  "utf-8",
				Fatal:     fatal,
				IgnoreBOM: ignoreBOM,
			},
		}, nil

	case "__text_decoder.encoding":
		return Value{
			Type:   ir.TypeString,
			String: "utf-8",
		}, nil

	case "__text_decoder.fatal":
		dec := env[instruction.Args[0]].TextDecoder
		val := false
		if dec != nil {
			val = dec.Fatal
		}
		return Value{
			Type: ir.TypeBool,
			Bool: val,
		}, nil

	case "__text_decoder.ignore_bom":
		dec := env[instruction.Args[0]].TextDecoder
		val := false
		if dec != nil {
			val = dec.IgnoreBOM
		}
		return Value{
			Type: ir.TypeBool,
			Bool: val,
		}, nil

	case "__text_decoder.decode":
		if len(instruction.Args) == 0 {
			return Value{Type: ir.TypeString, String: ""}, nil
		}
		decVal := env[instruction.Args[0]]
		fatal := false
		ignoreBOM := false
		if decVal.TextDecoder != nil {
			fatal = decVal.TextDecoder.Fatal
			ignoreBOM = decVal.TextDecoder.IgnoreBOM
		}

		var data []byte
		if len(instruction.Args) > 1 {
			inputVal := env[instruction.Args[1]]
			if inputVal.TypedArray != nil && inputVal.TypedArray.Buffer != nil {
				ta := inputVal.TypedArray
				start := ta.ByteOffset
				end := start + ta.ByteLength()
				if end <= len(ta.Buffer.Data) {
					data = ta.Buffer.Data[start:end]
				}
			} else if inputVal.DataView != nil && inputVal.DataView.Buffer != nil {
				dv := inputVal.DataView
				start := dv.ByteOffset
				end := start + dv.ByteLength
				if end <= len(dv.Buffer.Data) {
					data = dv.Buffer.Data[start:end]
				}
			} else if inputVal.Buffer != nil {
				data = inputVal.Buffer.Data
			}
		}

		if len(data) == 0 {
			return Value{Type: ir.TypeString, String: ""}, nil
		}

		if !ignoreBOM && bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}) {
			data = data[3:]
		}

		if fatal && !utf8.Valid(data) {
			return Value{}, fmt.Errorf("TypeError: The encoded data was not valid.")
		}

		return Value{
			Type:   ir.TypeString,
			String: string(data),
		}, nil

	default:
		return Value{}, fmt.Errorf("unknown text encoding intrinsic %s", instruction.Callee)
	}
}

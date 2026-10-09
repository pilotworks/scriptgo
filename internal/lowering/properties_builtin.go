package lowering

import (
	"strconv"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func tryLowerArrayBufferViewProperty(path string, expression *frontend.SyntaxExpression, result *string, function *ir.Function, counter *int, object string, objectType ir.Type) (string, ir.Type, bool, error) {
	viewIntrinsic := "__typedarray."
	if objectType == "ArrayBufferView" || objectType == "object:ArrayBufferView" {
		viewIntrinsic = "__arraybuffer_view."
	}
	if expression.Text == "length" {
		if *result == "" {
			*result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeNumber,
			Result: *result,
			Callee: "__typedarray.length",
			Args:   []string{object},
			Span:   toIRSpan(path, expression.Span),
		})
		return *result, ir.TypeNumber, true, nil
	}
	if expression.Text == "byteLength" {
		if *result == "" {
			*result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeNumber,
			Result: *result,
			Callee: viewIntrinsic + "byteLength",
			Args:   []string{object},
			Span:   toIRSpan(path, expression.Span),
		})
		return *result, ir.TypeNumber, true, nil
	}
	if expression.Text == "byteOffset" {
		if *result == "" {
			*result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeNumber,
			Result: *result,
			Callee: viewIntrinsic + "byteOffset",
			Args:   []string{object},
			Span:   toIRSpan(path, expression.Span),
		})
		return *result, ir.TypeNumber, true, nil
	}
	if expression.Text == "buffer" || expression.Text == "parent" {
		if *result == "" {
			*result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeArrayBuffer,
			Result: *result,
			Callee: viewIntrinsic + "buffer",
			Args:   []string{object},
			Span:   toIRSpan(path, expression.Span),
		})
		return *result, ir.TypeArrayBuffer, true, nil
	}
	if expression.Text == "BYTES_PER_ELEMENT" {
		elemSize := "1"
		switch objectType {
		case ir.TypeInt16Array, ir.TypeUint16Array:
			elemSize = "2"
		case ir.TypeInt32Array, ir.TypeUint32Array, ir.TypeFloat32Array:
			elemSize = "4"
		case ir.TypeFloat64Array, ir.TypeBigInt64Array, ir.TypeBigUint64Array:
			elemSize = "8"
		}
		if *result == "" {
			*result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   ir.TypeNumber,
			Result: *result,
			Value:  elemSize,
			Span:   toIRSpan(path, expression.Span),
		})
		return *result, ir.TypeNumber, true, nil
	}

	return "", "", false, nil
}

func tryLowerDataViewProperty(path string, expression *frontend.SyntaxExpression, result *string, function *ir.Function, counter *int, object string) (string, ir.Type, bool, error) {
	if expression.Text == "byteLength" {
		if *result == "" {
			*result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeNumber,
			Result: *result,
			Callee: "__dataview.byteLength",
			Args:   []string{object},
			Span:   toIRSpan(path, expression.Span),
		})
		return *result, ir.TypeNumber, true, nil
	}
	if expression.Text == "byteOffset" {
		if *result == "" {
			*result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeNumber,
			Result: *result,
			Callee: "__dataview.byteOffset",
			Args:   []string{object},
			Span:   toIRSpan(path, expression.Span),
		})
		return *result, ir.TypeNumber, true, nil
	}
	if expression.Text == "buffer" {
		if *result == "" {
			*result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeArrayBuffer,
			Result: *result,
			Callee: "__dataview.buffer",
			Args:   []string{object},
			Span:   toIRSpan(path, expression.Span),
		})
		return *result, ir.TypeArrayBuffer, true, nil
	}

	return "", "", false, nil
}

func tryLowerTextDecoderProperty(path string, expression *frontend.SyntaxExpression, result *string, function *ir.Function, counter *int, object string) (string, ir.Type, bool, error) {
	switch expression.Text {
	case "encoding":
		if *result == "" {
			*result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeString,
			Result: *result,
			Callee: "__text_decoder.encoding",
			Args:   []string{object},
			Span:   toIRSpan(path, expression.Span),
		})
		return *result, ir.TypeString, true, nil
	case "fatal":
		if *result == "" {
			*result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeBool,
			Result: *result,
			Callee: "__text_decoder.fatal",
			Args:   []string{object},
			Span:   toIRSpan(path, expression.Span),
		})
		return *result, ir.TypeBool, true, nil
	case "ignoreBOM":
		if *result == "" {
			*result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeBool,
			Result: *result,
			Callee: "__text_decoder.ignore_bom",
			Args:   []string{object},
			Span:   toIRSpan(path, expression.Span),
		})
		return *result, ir.TypeBool, true, nil
	}

	return "", "", false, nil
}

var regExpFlagLetters = map[string]string{
	"global": "g", "ignoreCase": "i", "multiline": "m", "dotAll": "s",
	"unicode": "u", "sticky": "y", "hasIndices": "d", "unicodeSets": "v",
}

func tryLowerRegExpProperty(path string, expression *frontend.SyntaxExpression, result *string, function *ir.Function, counter *int, object string) (string, ir.Type, bool, error) {
	switch expression.Text {
	case "global", "ignoreCase", "multiline", "dotAll", "unicode", "sticky", "hasIndices", "unicodeSets":
		if *result == "" {
			*result = nextTemp(counter)
		}
		// Each flag accessor reports whether its letter is in the stored
		// flags string.
		span := toIRSpan(path, expression.Span)
		flags := nextTemp(counter)
		letter := nextTemp(counter)
		function.Body = append(function.Body,
			ir.Instruction{Op: ir.OpFieldGet, Type: ir.TypeString, Result: flags, Callee: "RegExp", Field: "flags", FieldIndex: 1, Args: []string{object}, Span: span},
			ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: letter, Value: regExpFlagLetters[expression.Text], Span: span},
			ir.Instruction{Op: ir.OpCall, Type: ir.TypeBool, Result: *result, Callee: "__string.includes", Args: []string{flags, letter}, Span: span},
		)
		return *result, ir.TypeBool, true, nil
	case "lastIndex":
		if *result == "" {
			*result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:         ir.OpFieldGet,
			Type:       ir.TypeNumber,
			Result:     *result,
			Callee:     "RegExp",
			Field:      "lastIndex",
			FieldIndex: 2,
			Args:       []string{object},
			Span:       toIRSpan(path, expression.Span),
		})
		return *result, ir.TypeNumber, true, nil
	}

	return "", "", false, nil
}

func tryLowerLengthProperty(path string, expression *frontend.SyntaxExpression, result *string, function *ir.Function, counter *int, object string, objectType ir.Type) (string, ir.Type, bool, error) {
	if objectType == ir.TypeString {
		if *result == "" {
			*result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeNumber,
			Result: *result,
			Callee: "__string.length",
			Args:   []string{object},
			Span:   toIRSpan(path, expression.Span),
		})
		return *result, ir.TypeNumber, true, nil
	}
	if strings.HasSuffix(string(objectType), "[]") || objectType == ir.TypeStringArray || objectType == ir.TypeNumberArray || objectType == ir.TypeBoolArray || objectType == ir.Type("symbol[]") {
		if *result == "" {
			*result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeNumber,
			Result: *result,
			Callee: "__array.length",
			Args:   []string{object},
			Span:   toIRSpan(path, expression.Span),
		})
		return *result, ir.TypeNumber, true, nil
	}
	if strings.HasPrefix(string(objectType), "object:") {
		shapeName := strings.TrimPrefix(string(objectType), "object:")
		if s, ok := anonymousShapes[shapeName]; ok && len(s.Fields) > 0 && s.Fields[0].Name == "0" {
			if *result == "" {
				*result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   ir.TypeNumber,
				Result: *result,
				Value:  strconv.Itoa(len(s.Fields)),
				Span:   toIRSpan(path, expression.Span),
			})
			return *result, ir.TypeNumber, true, nil
		}
	}

	return "", "", false, nil
}

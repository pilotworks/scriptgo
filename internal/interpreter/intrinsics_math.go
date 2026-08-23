package interpreter

import (
	"fmt"
	"math"
	"math/bits"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func executeMathIntrinsic(name string, arguments []string, env map[string]Value) (Value, error) {
	values := make([]float64, 0, len(arguments))
	for _, arg := range arguments {
		val, ok := env[arg]
		if !ok || val.Type != ir.TypeNumber {
			return Value{}, fmt.Errorf("%s requires number operands", name)
		}
		values = append(values, val.Number)
	}
	switch name {
	case "__Math.abs":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Abs(values[0])}, nil
	case "__Math.ceil":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Ceil(values[0])}, nil
	case "__Math.floor":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Floor(values[0])}, nil
	case "__Math.trunc":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Trunc(values[0])}, nil
	case "__Math.sqrt":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Sqrt(values[0])}, nil
	case "__Math.round":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Round(values[0])}, nil
	case "__Math.sin":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Sin(values[0])}, nil
	case "__Math.cos":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Cos(values[0])}, nil
	case "__Math.tan":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Tan(values[0])}, nil
	case "__Math.atan":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Atan(values[0])}, nil
	case "__Math.asin":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Asin(values[0])}, nil
	case "__Math.acos":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Acos(values[0])}, nil
	case "__Math.cbrt":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Cbrt(values[0])}, nil
	case "__Math.fround":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: float64(float32(values[0]))}, nil
	case "__Math.f16round":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: float16round(values[0])}, nil
	case "__Math.clz32":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		u := uint32(int64(values[0]))
		return Value{Type: ir.TypeNumber, Number: float64(bits.LeadingZeros32(u))}, nil
	case "__Math.imul":
		if len(values) != 2 {
			return Value{}, fmt.Errorf("%s requires 2 arguments", name)
		}
		a := int32(uint32(int64(values[0])))
		b := int32(uint32(int64(values[1])))
		return Value{Type: ir.TypeNumber, Number: float64(a * b)}, nil
	case "__Math.sinh":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Sinh(values[0])}, nil
	case "__Math.cosh":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Cosh(values[0])}, nil
	case "__Math.tanh":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Tanh(values[0])}, nil
	case "__Math.asinh":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Asinh(values[0])}, nil
	case "__Math.acosh":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Acosh(values[0])}, nil
	case "__Math.atanh":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Atanh(values[0])}, nil
	case "__Math.expm1":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Expm1(values[0])}, nil
	case "__Math.log1p":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Log1p(values[0])}, nil
	case "__Math.atan2":
		if len(values) != 2 {
			return Value{}, fmt.Errorf("%s requires 2 arguments", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Atan2(values[0], values[1])}, nil
	case "__Math.log":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Log(values[0])}, nil
	case "__Math.log2":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Log2(values[0])}, nil
	case "__Math.log10":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Log10(values[0])}, nil
	case "__Math.exp":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Exp(values[0])}, nil
	case "__Math.sign":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		if math.IsNaN(values[0]) {
			return Value{Type: ir.TypeNumber, Number: math.NaN()}, nil
		}
		if values[0] > 0 {
			return Value{Type: ir.TypeNumber, Number: 1}, nil
		} else if values[0] < 0 {
			return Value{Type: ir.TypeNumber, Number: -1}, nil
		}
		return Value{Type: ir.TypeNumber, Number: values[0]}, nil
	case "__Math.hypot":
		if len(values) != 2 {
			return Value{}, fmt.Errorf("%s requires 2 arguments", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Hypot(values[0], values[1])}, nil
	case "__Math.random":
		return Value{Type: ir.TypeNumber, Number: 0.5}, nil
	case "__Math.min":
		if len(values) != 2 {
			return Value{}, fmt.Errorf("%s requires 2 arguments", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Min(values[0], values[1])}, nil
	case "__Math.max":
		if len(values) != 2 {
			return Value{}, fmt.Errorf("%s requires 2 arguments", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Max(values[0], values[1])}, nil
	case "__Math.pow":
		if len(values) != 2 {
			return Value{}, fmt.Errorf("%s requires 2 arguments", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Pow(values[0], values[1])}, nil
	default:
		return Value{}, fmt.Errorf("unknown math intrinsic %q", name)
	}
}

func float16round(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) || v == 0 {
		return v
	}
	f32 := float32(v)
	bits := math.Float32bits(f32)
	sign := uint16((bits >> 16) & 0x8000)
	exp := int((bits >> 23) & 0xff) - 127 + 15
	mant := bits & 0x7fffff

	var h uint16
	if exp <= 0 {
		if exp < -10 {
			h = sign
		} else {
			mant = (mant | 0x800000) >> uint(1-exp)
			h = sign | uint16((mant+0x1000)>>13)
		}
	} else if exp >= 31 {
		h = sign | 0x7c00
	} else {
		h = sign | uint16(exp<<10) | uint16((mant+0x1000)>>13)
	}

	hSign := uint32(h&0x8000) << 16
	hExp := int((h >> 10) & 0x1f)
	hMant := uint32(h & 0x3ff)
	var outBits uint32
	if hExp == 0 {
		if hMant == 0 {
			outBits = hSign
		} else {
			for (hMant & 0x400) == 0 {
				hMant <<= 1
				hExp--
			}
			hExp++
			hMant &= 0x3ff
			outBits = hSign | uint32((hExp+127-15)<<23) | (hMant << 13)
		}
	} else if hExp == 31 {
		if hMant == 0 {
			outBits = hSign | 0x7f800000
		} else {
			outBits = hSign | 0x7f800000 | (hMant << 13)
		}
	} else {
		outBits = hSign | uint32((hExp+127-15)<<23) | (hMant << 13)
	}
	return float64(math.Float32frombits(outBits))
}

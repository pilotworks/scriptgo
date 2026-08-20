package interpreter

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/pilotworks/scriptgo/internal/ir"
)

var (
	symbolMu         sync.Mutex
	symbolCounter    uint64 = 1000
	symbolRegistry          = make(map[string]Value)
	wellKnownSymbols        = make(map[string]Value)
)

func nextSymbol(desc string) Value {
	id := atomic.AddUint64(&symbolCounter, 1)
	return Value{
		Type:       ir.TypeSymbol,
		SymbolID:   id,
		SymbolDesc: desc,
	}
}

func getOrCreateWellKnownSymbol(name string) Value {
	symbolMu.Lock()
	defer symbolMu.Unlock()
	if s, ok := wellKnownSymbols[name]; ok {
		return s
	}
	desc := name
	s := nextSymbol(desc)
	wellKnownSymbols[name] = s
	return s
}

func executeSymbolIntrinsic(callee string, args []string, env map[string]Value) (Value, error) {
	symbolMu.Lock()
	defer symbolMu.Unlock()

	switch callee {
	case "__symbol.create":
		desc := ""
		if len(args) > 0 {
			v, err := lookup(env, args, 0)
			if err != nil {
				return Value{}, err
			}
			desc = v.String
		}
		return nextSymbol(desc), nil
	case "__symbol.for":
		if len(args) != 1 {
			return Value{}, fmt.Errorf("Symbol.for requires 1 argument")
		}
		keyVal, err := lookup(env, args, 0)
		if err != nil {
			return Value{}, err
		}
		key := keyVal.String
		if s, ok := symbolRegistry[key]; ok {
			return s, nil
		}
		s := nextSymbol(key)
		symbolRegistry[key] = s
		return s, nil
	case "__symbol.keyFor":
		if len(args) != 1 {
			return Value{}, fmt.Errorf("Symbol.keyFor requires 1 argument")
		}
		symVal, err := lookup(env, args, 0)
		if err != nil {
			return Value{}, err
		}
		for k, s := range symbolRegistry {
			if s.SymbolID == symVal.SymbolID {
				return Value{Type: ir.TypeString, String: k}, nil
			}
		}
		return Value{Type: ir.TypeString, String: "undefined"}, nil
	case "__symbol.toString":
		if len(args) != 1 {
			return Value{}, fmt.Errorf("Symbol.toString requires 1 argument")
		}
		symVal, err := lookup(env, args, 0)
		if err != nil {
			return Value{}, err
		}
		return Value{Type: ir.TypeString, String: format(symVal)}, nil
	case "__symbol.description":
		if len(args) != 1 {
			return Value{}, fmt.Errorf("Symbol.description requires 1 argument")
		}
		symVal, err := lookup(env, args, 0)
		if err != nil {
			return Value{}, err
		}
		if symVal.SymbolDesc == "" {
			return Value{Type: ir.TypeString, String: "undefined"}, nil
		}
		return Value{Type: ir.TypeString, String: symVal.SymbolDesc}, nil
	default:
		return Value{}, fmt.Errorf("unsupported symbol intrinsic %q", callee)
	}
}

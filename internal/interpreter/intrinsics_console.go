package interpreter

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/pilotworks/scriptgo/internal/ir"
)

var (
	consoleMu          sync.Mutex
	consoleIndentLevel int
	consoleCounters    = map[string]int{}
	consoleTimers      = map[string]time.Time{}
)

// ResetConsoleState resets any active indentation, counts, and timers.
func ResetConsoleState() {
	consoleMu.Lock()
	defer consoleMu.Unlock()
	consoleIndentLevel = 0
	consoleCounters = map[string]int{}
	consoleTimers = map[string]time.Time{}
}

func getConsoleIndent() string {
	consoleMu.Lock()
	defer consoleMu.Unlock()
	return strings.Repeat("  ", consoleIndentLevel)
}

func executeConsoleIntrinsic(name string, arguments []string, env map[string]Value, output io.Writer) (Value, error) {
	consoleMu.Lock()
	defer consoleMu.Unlock()

	switch name {
	case "__console.clear":
		fmt.Fprint(output, "\x1b[2J\x1b[0f")
		return Value{Type: ir.TypeVoid}, nil
	case "__console.group", "__console.groupCollapsed":
		consoleIndentLevel++
		return Value{Type: ir.TypeVoid}, nil
	case "__console.groupEnd":
		if consoleIndentLevel > 0 {
			consoleIndentLevel--
		}
		return Value{Type: ir.TypeVoid}, nil
	case "__console.new":
		return Value{Type: ir.Type("object:Console"), Object: map[string]Value{}}, nil
	case "__console.profile", "__console.profileEnd", "__console.timeStamp":
		return Value{Type: ir.TypeVoid}, nil
	case "__console.count":
		label := "default"
		if len(arguments) > 0 {
			val, ok := env[arguments[0]]
			if ok && val.Type == ir.TypeString && val.String != "" {
				label = val.String
			}
		}
		consoleCounters[label]++
		indent := strings.Repeat("  ", consoleIndentLevel)
		fmt.Fprintf(output, "%s%s: %d\n", indent, label, consoleCounters[label])
		return Value{Type: ir.TypeVoid}, nil
	case "__console.countReset":
		label := "default"
		if len(arguments) > 0 {
			val, ok := env[arguments[0]]
			if ok && val.Type == ir.TypeString && val.String != "" {
				label = val.String
			}
		}
		if _, exists := consoleCounters[label]; exists {
			consoleCounters[label] = 0
		} else {
			fmt.Fprintf(output, "Count for '%s' does not exist\n", label)
		}
		return Value{Type: ir.TypeVoid}, nil
	case "__console.time":
		label := "default"
		if len(arguments) > 0 {
			val, ok := env[arguments[0]]
			if ok && val.Type == ir.TypeString && val.String != "" {
				label = val.String
			}
		}
		if _, active := consoleTimers[label]; active {
			fmt.Fprintf(output, "Timer '%s' already exists\n", label)
		} else {
			consoleTimers[label] = time.Now()
		}
		return Value{Type: ir.TypeVoid}, nil
	case "__console.timeLog":
		label := "default"
		if len(arguments) > 0 {
			val, ok := env[arguments[0]]
			if ok && val.Type == ir.TypeString && val.String != "" {
				label = val.String
			}
		}
		startTime, active := consoleTimers[label]
		if !active {
			fmt.Fprintf(output, "Timer '%s' does not exist\n", label)
			return Value{Type: ir.TypeVoid}, nil
		}
		elapsed := time.Since(startTime).Seconds() * 1000.0
		indent := strings.Repeat("  ", consoleIndentLevel)
		if len(arguments) > 1 {
			val, ok := env[arguments[1]]
			if ok {
				fmt.Fprintf(output, "%s%s: %.3fms %s\n", indent, label, elapsed, val.String)
				return Value{Type: ir.TypeVoid}, nil
			}
		}
		fmt.Fprintf(output, "%s%s: %.3fms\n", indent, label, elapsed)
		return Value{Type: ir.TypeVoid}, nil
	case "__console.timeEnd":
		label := "default"
		if len(arguments) > 0 {
			val, ok := env[arguments[0]]
			if ok && val.Type == ir.TypeString && val.String != "" {
				label = val.String
			}
		}
		startTime, active := consoleTimers[label]
		if !active {
			fmt.Fprintf(output, "Timer '%s' does not exist\n", label)
			return Value{Type: ir.TypeVoid}, nil
		}
		elapsed := time.Since(startTime).Seconds() * 1000.0
		delete(consoleTimers, label)
		indent := strings.Repeat("  ", consoleIndentLevel)
		fmt.Fprintf(output, "%s%s: %.3fms\n", indent, label, elapsed)
		return Value{Type: ir.TypeVoid}, nil
	case "__console.trace":
		indent := strings.Repeat("  ", consoleIndentLevel)
		if len(arguments) > 0 {
			val, ok := env[arguments[0]]
			if ok && val.String != "" {
				fmt.Fprintf(output, "%sTrace: %s\n", indent, val.String)
				return Value{Type: ir.TypeVoid}, nil
			}
		}
		fmt.Fprintf(output, "%sTrace\n", indent)
		return Value{Type: ir.TypeVoid}, nil
	default:
		return Value{}, fmt.Errorf("unknown console intrinsic %q", name)
	}
}

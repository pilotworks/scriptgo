package interpreter

import (
	"bytes"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/pilotworks/scriptgo/internal/ir"
)

type timerItem struct {
	id        int
	closure   *Closure
	delayMs   float64
	interval  bool
	canceled  bool
	fireTime  time.Time
	startTime time.Time
}

var (
	timerMu      sync.Mutex
	nextTimerID  = 1
	activeTimers = map[int]*timerItem{}
)

func resetTimers() {
	timerMu.Lock()
	defer timerMu.Unlock()
	nextTimerID = 1
	activeTimers = map[int]*timerItem{}
}

func executeTimerIntrinsic(name string, arguments []string, env map[string]Value) (Value, error) {
	timerMu.Lock()
	defer timerMu.Unlock()

	switch name {
	case "__timers.setTimeout", "__timers.setInterval", "__timers.setImmediate":
		if len(arguments) < 1 {
			return Value{}, fmt.Errorf("%s requires callback", name)
		}
		closureVal, ok := env[arguments[0]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("%s callback must be a closure", name)
		}
		delay := 0.0
		if len(arguments) > 1 {
			dVal, ok2 := env[arguments[1]]
			if ok2 && dVal.Type == ir.TypeNumber {
				delay = dVal.Number
			}
		}
		if name == "__timers.setImmediate" {
			delay = 0
		}
		id := nextTimerID
		nextTimerID++
		now := time.Now()
		item := &timerItem{
			id:        id,
			closure:   closureVal.Closure,
			delayMs:   delay,
			interval:  name == "__timers.setInterval",
			startTime: now,
			fireTime:  now.Add(time.Duration(delay * float64(time.Millisecond))),
		}
		for k, v := range closureVal.Closure.Env {
			if (k == "id" || k == "intervalId" || k == "timerId") && v.Number == 0 {
				closureVal.Closure.Env[k] = Value{Type: ir.TypeNumber, Number: float64(id)}
			}
		}
		activeTimers[id] = item
		return Value{Type: ir.TypeNumber, Number: float64(id)}, nil

	case "__timers.clearTimeout", "__timers.clearInterval", "__timers.clearImmediate":
		if len(arguments) < 1 {
			return Value{Type: ir.TypeVoid}, nil
		}
		idVal, ok := env[arguments[0]]
		if ok && idVal.Type == ir.TypeNumber {
			id := int(idVal.Number)
			if item, exists := activeTimers[id]; exists {
				item.canceled = true
				delete(activeTimers, id)
			}
		}
		return Value{Type: ir.TypeVoid}, nil

	default:
		return Value{}, fmt.Errorf("unknown timer intrinsic %q", name)
	}
}

func drainTimers(functions map[string]ir.Function, output *bytes.Buffer) error {
	for {
		if err := drainMicrotasks(functions, output); err != nil {
			return err
		}

		timerMu.Lock()
		if len(activeTimers) == 0 {
			timerMu.Unlock()
			break
		}

		var items []*timerItem
		for _, item := range activeTimers {
			if !item.canceled {
				items = append(items, item)
			}
		}
		if len(items) == 0 {
			activeTimers = map[int]*timerItem{}
			timerMu.Unlock()
			break
		}

		sort.Slice(items, func(i, j int) bool {
			if items[i].fireTime.Equal(items[j].fireTime) {
				return items[i].id < items[j].id
			}
			return items[i].fireTime.Before(items[j].fireTime)
		})

		next := items[0]
		now := time.Now()
		if next.fireTime.After(now) {
			sleepDuration := next.fireTime.Sub(now)
			if sleepDuration > 0 && sleepDuration < 10*time.Second {
				timerMu.Unlock()
				time.Sleep(sleepDuration)
				timerMu.Lock()
			}
		}

		if next.canceled {
			delete(activeTimers, next.id)
			timerMu.Unlock()
			continue
		}

		if next.interval {
			next.fireTime = time.Now().Add(time.Duration(next.delayMs * float64(time.Millisecond)))
		} else {
			delete(activeTimers, next.id)
		}
		closure := next.closure
		timerMu.Unlock()

		_, _, err := executeClosure(functions, closure, nil, output)
		if err != nil {
			return err
		}
	}
	return nil
}

package lowering

import (
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func resolveFallbackCallTarget(path string, expression *frontend.SyntaxExpression, env map[string]ir.Type, signatures map[string]ir.Function, callee string, target ir.Function, ok bool) (ir.Function, bool) {
	if strings.HasPrefix(callee, "fs.promises.") {
		method := strings.TrimPrefix(callee, "fs.promises.")
		if sig, ok2 := signatures["FSPromises."+method]; ok2 {
			target = sig
			ok = true
		} else if sig, ok2 := signatures["FSPromises_"+method]; ok2 {
			target = sig
			ok = true
		}
	} else if strings.HasPrefix(callee, "promises.") {
		method := strings.TrimPrefix(callee, "promises.")
		if sig, ok2 := signatures["FSPromises."+method]; ok2 {
			target = sig
			ok = true
		} else if sig, ok2 := signatures["FSPromises_"+method]; ok2 {
			target = sig
			ok = true
		}
	} else if strings.Contains(callee, ".") {
		parts := strings.Split(callee, ".")
		funcName := parts[len(parts)-1]
		if expression.Left != nil && expression.Left.Left != nil {
			recvType := env[expression.Left.Left.Text]
			if recvType == "" {
				if topVar, okVar := topLevelVars[expression.Left.Left.Text]; okVar {
					if topVar.Type != "" {
						recvType = toIRTypeForPath(path, topVar.Type)
					} else if topVar.InferredType != "" {
						recvType = toIRTypeForPath(path, topVar.InferredType)
					}
				}
			}
			if recvType == "" && expression.Left.Left.InferredType != "" {
				recvType = toIRType(expression.Left.Left.InferredType)
			}
			if strings.HasPrefix(string(recvType), "object:") {
				cls := strings.TrimPrefix(string(recvType), "object:")
				if idx := strings.Index(cls, "<"); idx != -1 {
					cls = cls[:idx]
				}
				cls = classIdentityForPath(path, cls)
				mangled := cls + "_" + funcName
				if sig, ok2 := signatures[mangled]; ok2 {
					target = sig
					ok = true
				} else if fn, _, okFind := findMethodInHierarchy(cls, funcName, signatures, classHierarchy); okFind {
					target = fn
					ok = true
				}
			}
		}
		if !ok {
			if sig, ok2 := signatures[funcName]; ok2 {
				target = sig
				ok = true
			}
		}
	}

	return target, ok
}

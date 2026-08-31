package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitSqliteIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	resolvedArgs := make([]string, len(instruction.Args))
	for i, arg := range instruction.Args {
		resolvedArgs[i] = e.resolveArg(out, arg)
	}

	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++

	switch instruction.Callee {
	case "__sqlite.open":
		slot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		flagsArg := "6.0"
		if len(resolvedArgs) > 1 {
			flagsArg = fmt.Sprintf("%%%s", resolvedArgs[1])
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_sqlite_open(ptr %%%s, double %s, ptr %%%s)\n", status, resolvedArgs[0], flagsArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__sqlite.close":
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_sqlite_close(ptr %%%s)\n", status, resolvedArgs[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__sqlite.exec":
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_sqlite_exec(ptr %%%s, ptr %%%s)\n", status, resolvedArgs[0], resolvedArgs[1])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__sqlite.prepare":
		slot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_sqlite_prepare(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, resolvedArgs[0], resolvedArgs[1], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__sqlite.run":
		slot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		firstArg := "null"
		if len(resolvedArgs) > 1 {
			p, err := e.emitSqliteBoxedArg(out, instruction.Result+".p0", resolvedArgs[1])
			if err != nil {
				return err
			}
			firstArg = p
		}
		restArg := "null"
		if len(resolvedArgs) > 2 {
			restArg = fmt.Sprintf("%%%s", resolvedArgs[2])
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_sqlite_run(ptr %%%s, ptr %s, ptr %s, ptr %%%s)\n", status, resolvedArgs[0], firstArg, restArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__sqlite.get":
		slot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		firstArg := "null"
		if len(resolvedArgs) > 1 {
			p, err := e.emitSqliteBoxedArg(out, instruction.Result+".p0", resolvedArgs[1])
			if err != nil {
				return err
			}
			firstArg = p
		}
		restArg := "null"
		if len(resolvedArgs) > 2 {
			restArg = fmt.Sprintf("%%%s", resolvedArgs[2])
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_sqlite_get(ptr %%%s, ptr %s, ptr %s, ptr %%%s)\n", status, resolvedArgs[0], firstArg, restArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__sqlite.all":
		slot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		firstArg := "null"
		if len(resolvedArgs) > 1 {
			p, err := e.emitSqliteBoxedArg(out, instruction.Result+".p0", resolvedArgs[1])
			if err != nil {
				return err
			}
			firstArg = p
		}
		restArg := "null"
		if len(resolvedArgs) > 2 {
			restArg = fmt.Sprintf("%%%s", resolvedArgs[2])
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_sqlite_all(ptr %%%s, ptr %s, ptr %s, ptr %%%s)\n", status, resolvedArgs[0], firstArg, restArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__sqlite.columns":
		slot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_sqlite_columns(ptr %%%s, ptr %%%s)\n", status, resolvedArgs[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__sqlite.expandedSQL":
		slot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_sqlite_expanded_sql(ptr %%%s, ptr %%%s)\n", status, resolvedArgs[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__sqlite.finalize":
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_sqlite_finalize(ptr %%%s)\n", status, resolvedArgs[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__sqlite.stmtConfig":
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_sqlite_stmt_config(ptr %%%s, double %%%s, double %%%s, double %%%s, double %%%s)\n",
			status, resolvedArgs[0], resolvedArgs[1], resolvedArgs[2], resolvedArgs[3], resolvedArgs[4])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__sqlite.enableLoadExtension":
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_sqlite_enable_load_extension(ptr %%%s, double %%%s)\n", status, resolvedArgs[0], resolvedArgs[1])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__sqlite.loadExtension":
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_sqlite_load_extension(ptr %%%s, ptr %%%s)\n", status, resolvedArgs[0], resolvedArgs[1])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__sqlite.createSession":
		slot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		tableArg := "null"
		if len(resolvedArgs) > 1 {
			tableArg = fmt.Sprintf("%%%s", resolvedArgs[1])
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_sqlite_session_create(ptr %%%s, ptr %s, ptr %%%s)\n", status, resolvedArgs[0], tableArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__sqlite.sessionChangeset":
		slot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_sqlite_session_changeset(ptr %%%s, ptr %%%s)\n", status, resolvedArgs[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__sqlite.sessionPatchset":
		slot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_sqlite_session_patchset(ptr %%%s, ptr %%%s)\n", status, resolvedArgs[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__sqlite.sessionClose":
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_sqlite_session_close(ptr %%%s)\n", status, resolvedArgs[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__sqlite.applyChangeset":
		slot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca i32\n", slot)
		onConflictArg := "0.0"
		if len(resolvedArgs) > 2 {
			onConflictArg = fmt.Sprintf("%%%s", resolvedArgs[2])
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_sqlite_apply_changeset(ptr %%%s, ptr %%%s, double %s, ptr %%%s)\n",
			status, resolvedArgs[0], resolvedArgs[1], onConflictArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s.i32 = load i32, ptr %%%s\n", instruction.Result, slot)
		fmt.Fprintf(out, "  %%%s = icmp ne i32 %%%s.i32, 0\n", instruction.Result, instruction.Result)
		return nil

	case "__sqlite.location":
		slot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		dbNameArg := "null"
		if len(resolvedArgs) > 1 {
			dbNameArg = fmt.Sprintf("%%%s", resolvedArgs[1])
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_sqlite_location(ptr %%%s, ptr %s, ptr %%%s)\n", status, resolvedArgs[0], dbNameArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__sqlite.isTransaction":
		slot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_sqlite_is_transaction(ptr %%%s, ptr %%%s)\n", status, resolvedArgs[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s.dbl = load double, ptr %%%s\n", instruction.Result, slot)
		fmt.Fprintf(out, "  %%%s = fcmp one double %%%s.dbl, 0.0\n", instruction.Result, instruction.Result)
		return nil

	case "__sqlite.createFunction":
		fnArg := "null"
		if len(resolvedArgs) > 4 {
			p, err := e.emitSqliteBoxedArg(out, instruction.Result+".fn", resolvedArgs[4])
			if err != nil {
				return err
			}
			fnArg = p
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_sqlite_create_function(ptr %%%s, ptr %%%s, double %%%s, double %%%s, ptr %s)\n",
			status, resolvedArgs[0], resolvedArgs[1], resolvedArgs[2], resolvedArgs[3], fnArg)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__sqlite.createAggregate":
		optArg := "null"
		if len(resolvedArgs) > 4 {
			p, err := e.emitSqliteBoxedArg(out, instruction.Result+".opt", resolvedArgs[4])
			if err != nil {
				return err
			}
			optArg = p
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_sqlite_create_aggregate(ptr %%%s, ptr %%%s, double %%%s, double %%%s, ptr %s)\n",
			status, resolvedArgs[0], resolvedArgs[1], resolvedArgs[2], resolvedArgs[3], optArg)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__sqlite.backup":
		slot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_sqlite_backup(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, resolvedArgs[0], resolvedArgs[1], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil

	default:
		return fmt.Errorf("unknown sqlite intrinsic %q", instruction.Callee)
	}
}

func (e *functionEmitter) emitSqliteBoxedArg(out *strings.Builder, resultName string, argName string) (string, error) {
	if argName == "" {
		return "null", nil
	}
	argType := e.types[argName]
	boxedVar := fmt.Sprintf("%s.boxed", resultName)
	if argType == ir.TypeUnknown {
		boxedVar = argName
	} else {
		if err := e.emitBoxValue(out, argName, argType, boxedVar); err != nil {
			return "", err
		}
	}
	slot := fmt.Sprintf("%s.box_slot", resultName)
	fmt.Fprintf(out, "  %%%s = alloca { i32, i32, i64 }\n", slot)
	fmt.Fprintf(out, "  store { i32, i32, i64 } %%%s, ptr %%%s\n", boxedVar, slot)
	return fmt.Sprintf("%%%s", slot), nil
}

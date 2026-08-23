package interpreter

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	goruntime "runtime"
	"strings"
	"time"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func executeFsIntrinsic(name string, arguments []string, env map[string]Value) (Value, error) {
	switch name {
	case "__fs.readFileSync":
		if len(arguments) < 1 || len(arguments) > 2 {
			return Value{}, fmt.Errorf("fs.readFileSync requires 1 or 2 arguments")
		}
		pathVal, ok := env[arguments[0]]
		if !ok || pathVal.Type != ir.TypeString {
			return Value{}, fmt.Errorf("fs.readFileSync requires a string path")
		}
		content, err := os.ReadFile(pathVal.String)
		if err != nil {
			return Value{}, err
		}
		return Value{Type: ir.TypeString, String: string(content)}, nil
	case "__fs.writeFileSync":
		if len(arguments) != 2 {
			return Value{}, fmt.Errorf("fs.writeFileSync requires 2 arguments")
		}
		pathVal, ok := env[arguments[0]]
		if !ok || pathVal.Type != ir.TypeString {
			return Value{}, fmt.Errorf("fs.writeFileSync requires a string path")
		}
		contentVal, ok := env[arguments[1]]
		if !ok || contentVal.Type != ir.TypeString {
			return Value{}, fmt.Errorf("fs.writeFileSync requires a string content")
		}
		err := os.WriteFile(pathVal.String, []byte(contentVal.String), 0644)
		if err != nil {
			return Value{}, err
		}
		return Value{Type: ir.TypeVoid}, nil
	case "__fs.existsSync":
		if len(arguments) != 1 {
			return Value{}, fmt.Errorf("fs.existsSync requires 1 argument")
		}
		pathVal, ok := env[arguments[0]]
		if !ok || pathVal.Type != ir.TypeString {
			return Value{}, fmt.Errorf("fs.existsSync requires a string path")
		}
		_, err := os.Stat(pathVal.String)
		return Value{Type: ir.TypeBool, Bool: err == nil}, nil
	case "__fs.unlinkSync":
		if len(arguments) != 1 {
			return Value{}, fmt.Errorf("fs.unlinkSync requires 1 argument")
		}
		pathVal, ok := env[arguments[0]]
		if !ok || pathVal.Type != ir.TypeString {
			return Value{}, fmt.Errorf("fs.unlinkSync requires a string path")
		}
		_ = os.Remove(pathVal.String)
		return Value{Type: ir.TypeVoid}, nil
	case "__fs.readdirSync":
		if len(arguments) < 1 {
			return Value{}, fmt.Errorf("fs.readdirSync requires 1 argument")
		}
		pathVal := env[arguments[0]].String
		entries, err := os.ReadDir(pathVal)
		if err != nil {
			return Value{}, err
		}
		var arr []Value
		for _, entry := range entries {
			arr = append(arr, Value{Type: ir.TypeString, String: entry.Name()})
		}
		return Value{Type: ir.TypeStringArray, Array: arr}, nil
	case "__fs.copyFileSync":
		if len(arguments) != 2 {
			return Value{}, fmt.Errorf("fs.copyFileSync requires 2 arguments")
		}
		src := env[arguments[0]].String
		dest := env[arguments[1]].String
		data, err := os.ReadFile(src)
		if err != nil {
			return Value{}, err
		}
		err = os.WriteFile(dest, data, 0644)
		if err != nil {
			return Value{}, err
		}
		return Value{Type: ir.TypeVoid}, nil
	case "__fs.renameSync":
		if len(arguments) != 2 {
			return Value{}, fmt.Errorf("fs.renameSync requires 2 arguments")
		}
		oldPath := env[arguments[0]].String
		newPath := env[arguments[1]].String
		err := os.Rename(oldPath, newPath)
		if err != nil {
			return Value{}, err
		}
		return Value{Type: ir.TypeVoid}, nil
	case "__fs.appendFileSync":
		if len(arguments) != 2 {
			return Value{}, fmt.Errorf("fs.appendFileSync requires 2 arguments")
		}
		pathVal := env[arguments[0]].String
		content := env[arguments[1]].String
		f, err := os.OpenFile(pathVal, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return Value{}, err
		}
		defer f.Close()
		_, err = f.WriteString(content)
		if err != nil {
			return Value{}, err
		}
		return Value{Type: ir.TypeVoid}, nil
	case "__fs.mkdirSync":
		if len(arguments) < 1 {
			return Value{}, fmt.Errorf("fs.mkdirSync requires 1 argument")
		}
		pathVal := env[arguments[0]].String
		isRec := false
		if len(arguments) > 1 && (env[arguments[1]].Bool || env[arguments[1]].Number > 0) {
			isRec = true
		}
		var err error
		if isRec {
			err = os.MkdirAll(pathVal, 0755)
		} else {
			err = os.Mkdir(pathVal, 0755)
		}
		if err != nil && !os.IsExist(err) {
			return Value{}, err
		}
		return Value{Type: ir.TypeVoid}, nil
	case "__fs.rmSync":
		if len(arguments) < 1 {
			return Value{}, fmt.Errorf("fs.rmSync requires 1 argument")
		}
		pathVal := env[arguments[0]].String
		isRec := false
		if len(arguments) > 1 && (env[arguments[1]].Bool || env[arguments[1]].Number > 0) {
			isRec = true
		}
		var err error
		if isRec {
			err = os.RemoveAll(pathVal)
		} else {
			err = os.Remove(pathVal)
		}
		if err != nil {
			isForce := false
			if len(arguments) > 2 && (env[arguments[2]].Bool || env[arguments[2]].Number > 0) {
				isForce = true
			}
			if !isForce && !os.IsNotExist(err) {
				return Value{}, err
			}
		}
		return Value{Type: ir.TypeVoid}, nil
	case "__fs.statSync":
		if len(arguments) != 1 {
			return Value{}, fmt.Errorf("fs.statSync requires 1 argument")
		}
		pathVal := env[arguments[0]].String
		info, err := os.Stat(pathVal)
		if err != nil {
			return Value{}, err
		}
		size := float64(info.Size())
		mtimeMs := float64(info.ModTime().UnixMilli())
		birthtimeMs := mtimeMs
		var mode float64
		if info.IsDir() {
			mode = float64(0040755)
		} else {
			mode = float64(0100644)
		}
		return Value{
			Type: "object:Stats",
			Object: map[string]Value{
				"size":        {Type: ir.TypeNumber, Number: size},
				"mtimeMs":     {Type: ir.TypeNumber, Number: mtimeMs},
				"birthtimeMs": {Type: ir.TypeNumber, Number: birthtimeMs},
				"mode":        {Type: ir.TypeNumber, Number: mode},
			},
		}, nil
	default:
		return Value{}, fmt.Errorf("unknown fs intrinsic %q", name)
	}
}

func executeProcessIntrinsic(name string, arguments []string, env map[string]Value) (Value, error) {
	switch name {
	case "__process.exit":
		return Value{Type: ir.TypeVoid}, nil
	case "__process.cwd":
		cwd, err := os.Getwd()
		if err != nil {
			return Value{}, err
		}
		return Value{Type: ir.TypeString, String: cwd}, nil
	case "__process.argv":
		args := os.Args
		arr := make([]Value, len(args))
		for i, a := range args {
			arr[i] = Value{Type: ir.TypeString, String: a}
		}
		return Value{Type: ir.TypeStringArray, Array: arr}, nil
	case "__process.env":
		if len(arguments) != 1 {
			return Value{}, fmt.Errorf("process.env requires 1 argument")
		}
		keyVal, ok := env[arguments[0]]
		if !ok || keyVal.Type != ir.TypeString {
			return Value{}, fmt.Errorf("process.env requires a string key")
		}
		return Value{Type: ir.TypeString, String: os.Getenv(keyVal.String)}, nil
	default:
		return Value{}, fmt.Errorf("unknown process intrinsic %q", name)
	}
}

func executeOsIntrinsic(name string, arguments []string, env map[string]Value) (Value, error) {
	switch name {
	case "__os.platform":
		platform := goruntime.GOOS
		if platform == "windows" {
			platform = "win32"
		}
		return Value{Type: ir.TypeString, String: platform}, nil
	case "__os.arch":
		arch := goruntime.GOARCH
		if arch == "amd64" {
			arch = "x64"
		} else if arch == "386" {
			arch = "ia32"
		}
		return Value{Type: ir.TypeString, String: arch}, nil
	case "__os.homedir":
		dir, _ := os.UserHomeDir()
		return Value{Type: ir.TypeString, String: dir}, nil
	case "__os.type":
		typ := "Darwin"
		if goruntime.GOOS == "linux" {
			typ = "Linux"
		} else if goruntime.GOOS == "windows" {
			typ = "Windows_NT"
		}
		return Value{Type: ir.TypeString, String: typ}, nil
	case "__os.release":
		return Value{Type: ir.TypeString, String: "1.0.0"}, nil
	case "__os.uptime":
		return Value{Type: ir.TypeNumber, Number: 3600.0}, nil
	case "__os.totalmem":
		return Value{Type: ir.TypeNumber, Number: 16.0 * 1024 * 1024 * 1024}, nil
	case "__os.freemem":
		return Value{Type: ir.TypeNumber, Number: 8.0 * 1024 * 1024 * 1024}, nil
	case "__os.tmpdir":
		return Value{Type: ir.TypeString, String: os.TempDir()}, nil
	default:
		return Value{}, fmt.Errorf("unknown os intrinsic %q", name)
	}
}

func executePerformanceIntrinsic(name string, arguments []string, env map[string]Value) (Value, error) {
	switch name {
	case "__performance.now":
		ms := float64(time.Now().UnixNano()) / 1e6
		return Value{Type: ir.TypeNumber, Number: ms}, nil
	default:
		return Value{}, fmt.Errorf("unknown performance intrinsic %q", name)
	}
}

func executeChildProcessIntrinsic(instruction ir.Instruction, env map[string]Value) (Value, error) {
	switch instruction.Callee {
	case "__child_process.execSync":
		if len(instruction.Args) < 1 {
			return Value{}, fmt.Errorf("child_process.execSync requires at least 1 argument")
		}
		command := env[instruction.Args[0]].String
		var cwd string
		var input string
		if len(instruction.Args) > 1 && instruction.Args[1] != "" {
			cwd = env[instruction.Args[1]].String
		}
		if len(instruction.Args) > 2 && instruction.Args[2] != "" {
			input = env[instruction.Args[2]].String
		}

		cmd := exec.Command("/bin/sh", "-c", command)
		if cwd != "" {
			cmd.Dir = cwd
		}
		if input != "" {
			cmd.Stdin = strings.NewReader(input)
		}
		var stdoutBuf, stderrBuf bytes.Buffer
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf
		err := cmd.Run()
		var exitCode float64 = 0
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = float64(exitErr.ExitCode())
			} else {
				exitCode = 1
			}
		}
		if instruction.Type == ir.TypeString {
			return Value{Type: ir.TypeString, String: stdoutBuf.String()}, nil
		}
		return Value{
			Type: "object:SpawnSyncReturns",
			Object: map[string]Value{
				"stdout": {Type: ir.TypeString, String: stdoutBuf.String()},
				"stderr": {Type: ir.TypeString, String: stderrBuf.String()},
				"status": {Type: ir.TypeNumber, Number: exitCode},
			},
		}, nil
	case "__child_process.spawnSync":
		if len(instruction.Args) < 1 {
			return Value{}, fmt.Errorf("child_process.spawnSync requires at least 1 argument")
		}
		command := env[instruction.Args[0]].String
		var args []string
		if len(instruction.Args) > 1 && instruction.Args[1] != "" {
			arrVal := env[instruction.Args[1]]
			for _, elem := range arrVal.Array {
				args = append(args, elem.String)
			}
		}
		var cwd string
		var input string
		if len(instruction.Args) > 2 && instruction.Args[2] != "" {
			cwd = env[instruction.Args[2]].String
		}
		if len(instruction.Args) > 3 && instruction.Args[3] != "" {
			input = env[instruction.Args[3]].String
		}

		cmd := exec.Command(command, args...)
		if cwd != "" {
			cmd.Dir = cwd
		}
		if input != "" {
			cmd.Stdin = strings.NewReader(input)
		}
		var stdoutBuf, stderrBuf bytes.Buffer
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf
		err := cmd.Run()
		var exitCode float64 = 0
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = float64(exitErr.ExitCode())
			} else {
				exitCode = 1
			}
		}
		return Value{
			Type: "object:SpawnSyncReturns",
			Object: map[string]Value{
				"stdout": {Type: ir.TypeString, String: stdoutBuf.String()},
				"stderr": {Type: ir.TypeString, String: stderrBuf.String()},
				"status": {Type: ir.TypeNumber, Number: exitCode},
			},
		}, nil
	default:
		return Value{}, fmt.Errorf("unknown child_process intrinsic %q", instruction.Callee)
	}
}

func executeHttpIntrinsic(instruction ir.Instruction, env map[string]Value) (Value, error) {
	switch instruction.Callee {
	case "__http.fetchSync":
		if len(instruction.Args) < 1 {
			return Value{}, fmt.Errorf("fetchSync requires at least 1 argument (url)")
		}
		url := env[instruction.Args[0]].String
		method := "GET"
		if len(instruction.Args) > 1 && instruction.Args[1] != "" {
			if mVal, ok := env[instruction.Args[1]]; ok && mVal.String != "" {
				method = strings.ToUpper(mVal.String)
			}
		}
		var headerPairs []string
		if len(instruction.Args) > 2 && instruction.Args[2] != "" {
			if hVal, ok := env[instruction.Args[2]]; ok {
				for _, elem := range hVal.Array {
					headerPairs = append(headerPairs, elem.String)
				}
			}
		}
		var body string
		if len(instruction.Args) > 3 && instruction.Args[3] != "" {
			if bVal, ok := env[instruction.Args[3]]; ok {
				body = bVal.String
			}
		}

		var reqBody io.Reader
		if body != "" {
			reqBody = strings.NewReader(body)
		}
		req, err := http.NewRequest(method, url, reqBody)
		if err != nil {
			return Value{}, fmt.Errorf("fetch error creating request: %w", err)
		}
		for i := 0; i+1 < len(headerPairs); i += 2 {
			req.Header.Add(headerPairs[i], headerPairs[i+1])
		}

		client := &http.Client{
			Timeout: 30 * time.Second,
		}
		resp, err := client.Do(req)
		var statusCode float64
		statusText := ""
		var respHeaders []Value
		respBodyStr := ""

		if err != nil {
			statusText = err.Error()
			statusCode = 0
		} else {
			defer resp.Body.Close()
			statusCode = float64(resp.StatusCode)
			statusText = resp.Status
			if idx := strings.Index(statusText, " "); idx != -1 {
				statusText = strings.TrimSpace(statusText[idx:])
			}
			respBytes, _ := io.ReadAll(resp.Body)
			respBodyStr = string(respBytes)

			for k, vList := range resp.Header {
				for _, v := range vList {
					respHeaders = append(respHeaders, Value{Type: ir.TypeString, String: strings.ToLower(k)})
					respHeaders = append(respHeaders, Value{Type: ir.TypeString, String: v})
				}
			}
		}

		return Value{
			Type: "object:FetchResponseData",
			Object: map[string]Value{
				"status":     {Type: ir.TypeNumber, Number: statusCode},
				"statusText": {Type: ir.TypeString, String: statusText},
				"headers":    {Type: ir.TypeStringArray, Array: respHeaders},
				"body":       {Type: ir.TypeString, String: respBodyStr},
			},
		}, nil
	default:
		return Value{}, fmt.Errorf("unknown http intrinsic %q", instruction.Callee)
	}
}

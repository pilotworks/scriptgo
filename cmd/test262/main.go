// Command test262 runs a subset of the TC39 test262 suite through scriptgo.
//
//	go run ./cmd/test262 -root ../test262 -paths language/expressions/addition,built-ins/Math
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"

	"github.com/pilotworks/scriptgo/internal/compiler"
	"github.com/pilotworks/scriptgo/internal/test262"
)

func main() {
	root := flag.String("root", "", "path to a test262 checkout (required)")
	paths := flag.String("paths", "language/expressions,language/statements", "comma-separated directories under test/")
	parallel := flag.Int("j", runtime.NumCPU(), "number of concurrent tests")
	timeout := flag.Duration("timeout", 10*time.Second, "per-test execution timeout")
	depth := flag.Int("depth", 3, "path segments used to group the summary")
	list := flag.String("list", "", "file of test paths (relative to test/, one per line) to run instead of -paths")
	jsonOut := flag.String("json", "", "write per-test results as JSON to this path")
	reasons := flag.Int("reasons", 15, "number of top unsupported/fail reasons to print")
	flag.Parse()
	if *root == "" {
		fmt.Fprintln(os.Stderr, "test262: -root is required")
		os.Exit(2)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	selected := strings.Split(*paths, ",")
	if *list != "" {
		data, err := os.ReadFile(*list)
		if err != nil {
			fmt.Fprintln(os.Stderr, "test262:", err)
			os.Exit(2)
		}
		selected = strings.Fields(string(data))
	}
	cfg := test262.Config{
		Root:     *root,
		Paths:    selected,
		Parallel: *parallel,
		Timeout:  *timeout,
		Build: func(entry, output string) error {
			return compiler.BuildWithOptions(entry, output, compiler.BuildOptions{})
		},
	}
	results, err := test262.Run(ctx, cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "test262:", err)
		os.Exit(1)
	}
	test262.WriteText(os.Stdout, test262.Summarize(results, *depth))
	for _, outcome := range []test262.Outcome{test262.Unsupported, test262.Fail} {
		if top := test262.TopReasons(results, outcome, *reasons); len(top) > 0 {
			fmt.Printf("\nTop %s reasons:\n%s\n", outcome, strings.Join(top, "\n"))
		}
	}
	if *jsonOut != "" {
		data, err := json.MarshalIndent(results, "", "  ")
		if err == nil {
			err = os.WriteFile(*jsonOut, data, 0o644)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "test262:", err)
			os.Exit(1)
		}
	}
}

package runtime

import (
	"fmt"
	"strings"
	"sync"
)

// The ECMAScript regular expression engine is QuickJS-ng's libregexp, taken
// from the vendored amalgam (native/dynamic/quickjs-amalgam.c) so it always
// matches the engine QuickJS itself runs. It is compiled as its own
// translation unit and linked into every executable; the regex runtime
// (native/regex/runtime.c) calls it through scriptgo_re_lre_compile and
// scriptgo_re_lre_exec.
//
// The amalgam concatenates QuickJS-ng's sources. The engine needs cutils,
// libunicode (with its tables) and libregexp, which sit in two runs:
//
//   - the start of the amalgam up to xsum.h: cutils.h, list.h, libunicode.h,
//     libregexp.h and libunicode-table.h;
//   - from cutils.c (the second "C utilities" header) up to xsum.c: cutils.c,
//     libregexp.c and libunicode.c.
const (
	regexpEngineHeadersEnd = "/* INTERFACE TO FUNCTIONS FOR EXACT SUMMATION. */"
	regexpEngineCutilsMark = "/*\n * C utilities\n"
	regexpEngineSourcesEnd = "\n/* FUNCTIONS FOR EXACT SUMMATION. */"
)

// regexpEngineSymbols are the external symbols the engine sources define.
// Each is renamed with a scriptgo_re_ prefix so the engine links alongside
// the full QuickJS-ng object of the dynamic runtime, which defines the same
// names. TestRegExpEngineSymbolsArePrefixed fails when an amalgam update adds
// a symbol missing here.
var regexpEngineSymbols = []string{
	"cr_copy", "cr_free", "cr_init", "cr_invert", "cr_op", "cr_realloc", "cr_regexp_canonicalize", "cr_union1",
	"dbuf_free", "dbuf_init", "dbuf_init2", "dbuf_printf", "dbuf_put", "dbuf_put_self", "dbuf_putc", "dbuf_putstr",
	"dbuf_realloc", "dbuf_write", "digits36", "i32toa", "i32toa_radix", "i64toa", "i64toa_radix",
	"js__gettimeofday_us", "js__has_suffix", "js__hrtime_ns", "js__pstrcat", "js__pstrcpy", "js__strstart",
	"js_cond_broadcast", "js_cond_destroy", "js_cond_init", "js_cond_signal", "js_cond_timedwait", "js_cond_wait",
	"js_exepath", "js_mutex_destroy", "js_mutex_init", "js_mutex_lock", "js_mutex_unlock", "js_once",
	"js_thread_create", "js_thread_join",
	"lre_byte_swap", "lre_canonicalize", "lre_case_conv", "lre_check_stack_overflow", "lre_check_timeout",
	"lre_compile", "lre_exec", "lre_get_capture_count", "lre_get_flags", "lre_get_groupnames",
	"lre_id_continue_table_ascii", "lre_id_start_table_ascii", "lre_is_case_ignorable", "lre_is_cased",
	"lre_is_id_continue", "lre_is_id_start", "lre_is_space", "lre_is_white_space", "lre_parse_escape",
	"lre_realloc", "rqsort", "u32toa", "u32toa_radix", "u64toa", "u64toa_radix",
	"unicode_general_category", "unicode_normalize", "unicode_prop", "unicode_script",
	"utf8_decode", "utf8_decode_buf16", "utf8_decode_buf8", "utf8_decode_len", "utf8_encode",
	"utf8_encode_buf16", "utf8_encode_buf8", "utf8_encode_len", "utf8_scan",
}

// regexpEngineCallbacks are the hooks libregexp requires from its embedder.
// The runtime passes the address of a stack limit as the opaque pointer.
const regexpEngineCallbacks = `
/* ScriptGo embedding: opaque is a pointer to the lowest stack address the
 * engine may use (NULL disables the check); there is no interrupt handler;
 * memory comes from the C allocator. */
bool lre_check_stack_overflow(void *opaque, size_t alloca_size) {
    char probe;
    if (opaque == NULL) return false;
    return (uintptr_t)&probe < *(const uintptr_t *)opaque + alloca_size;
}

int lre_check_timeout(void *opaque) {
    (void)opaque;
    return 0;
}

void *lre_realloc(void *opaque, void *ptr, size_t size) {
    (void)opaque;
    if (size == 0) {
        free(ptr);
        return NULL;
    }
    return realloc(ptr, size);
}
`

var (
	regexpEngineOnce   sync.Once
	regexpEngineSource []byte
	regexpEngineErr    error
)

// RegExpEngineSource is the C translation unit of the regular expression
// engine.
func RegExpEngineSource() ([]byte, error) {
	regexpEngineOnce.Do(func() {
		regexpEngineSource, regexpEngineErr = buildRegExpEngineSource(quickJSAmalgamSource)
	})
	return regexpEngineSource, regexpEngineErr
}

func buildRegExpEngineSource(amalgam string) ([]byte, error) {
	headersEnd := strings.Index(amalgam, regexpEngineHeadersEnd)
	if headersEnd < 0 {
		return nil, fmt.Errorf("regexp engine: QuickJS amalgam has no xsum.h marker")
	}
	first := strings.Index(amalgam, regexpEngineCutilsMark)
	cutilsStart := -1
	if first >= 0 {
		if next := strings.Index(amalgam[first+1:], regexpEngineCutilsMark); next >= 0 {
			cutilsStart = first + 1 + next
		}
	}
	if cutilsStart < headersEnd {
		return nil, fmt.Errorf("regexp engine: QuickJS amalgam has no cutils.c marker")
	}
	sourcesEnd := strings.Index(amalgam[cutilsStart:], regexpEngineSourcesEnd)
	if sourcesEnd < 0 {
		return nil, fmt.Errorf("regexp engine: QuickJS amalgam has no xsum.c marker")
	}
	var b strings.Builder
	b.WriteString("/* Generated from the QuickJS-ng amalgam by internal/runtime/regexp_engine.go. */\n")
	for _, name := range regexpEngineSymbols {
		fmt.Fprintf(&b, "#define %s scriptgo_re_%s\n", name, name)
	}
	b.WriteString(amalgam[:headersEnd])
	b.WriteString("\n")
	b.WriteString(amalgam[cutilsStart : cutilsStart+sourcesEnd])
	b.WriteString("\n")
	b.WriteString(regexpEngineCallbacks)
	return []byte(b.String()), nil
}

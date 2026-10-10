#include <stddef.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

/* JavaScript strings are indexed by UTF-16 code unit; ScriptGo stores them
 * as UTF-8. These helpers translate between the two for every string
 * operation that takes or returns a position. A code point above U+FFFF is
 * two code units; when an operation splits one, the lone surrogate is
 * written in its 3-byte generalized UTF-8 (WTF-8) form. */

/* utf8_step is the byte length of the UTF-8 sequence at p (1 for a stray or
 * truncated byte), and *units its UTF-16 length. */
static size_t utf8_step(const unsigned char *p, size_t *units) {
    *units = 1;
    if (p[0] < 0x80) return 1;
    if ((p[0] & 0xE0) == 0xC0 && p[1] != 0) return 2;
    if ((p[0] & 0xF0) == 0xE0 && p[1] != 0 && p[2] != 0) return 3;
    if ((p[0] & 0xF8) == 0xF0 && p[1] != 0 && p[2] != 0 && p[3] != 0) {
        *units = 2;
        return 4;
    }
    return 1;
}

static int utf8_is_ascii(const char *value) {
    for (const unsigned char *p = (const unsigned char *)value; *p != 0; p++) {
        if (*p >= 0x80) return 0;
    }
    return 1;
}

/* utf16_byte_offset is the byte offset where code unit `unit` starts; a
 * unit inside a 4-byte sequence maps to the sequence start. A unit past the
 * end maps to the end. */
static size_t utf16_byte_offset(const char *value, size_t unit) {
    const unsigned char *p = (const unsigned char *)value;
    size_t current = 0;
    while (*p != 0 && current < unit) {
        size_t units;
        size_t step = utf8_step(p, &units);
        if (current + units > unit) break;
        current += units;
        p += step;
    }
    return (size_t)(p - (const unsigned char *)value);
}

/* utf16_unit_offset is the code-unit index of byte offset `byte`. */
static size_t utf16_unit_offset(const char *value, size_t byte) {
    const unsigned char *p = (const unsigned char *)value;
    const unsigned char *end = p + byte;
    size_t units_total = 0;
    while (*p != 0 && p < end) {
        size_t units;
        p += utf8_step(p, &units);
        units_total += units;
    }
    return units_total;
}

static size_t utf16_length(const char *value) {
    return utf16_unit_offset(value, strlen(value));
}

/* utf8_write3 writes a code unit below 0x10000 (a lone surrogate included)
 * as 3 bytes. */
static size_t utf8_write3(char *out, uint32_t unit) {
    out[0] = (char)(0xE0 | (unit >> 12));
    out[1] = (char)(0x80 | ((unit >> 6) & 0x3F));
    out[2] = (char)(0x80 | (unit & 0x3F));
    return 3;
}

/* utf16_slice copies code units [start, end) of value into a new string. */
static int utf16_slice(const char *value, size_t start, size_t end, char **out) {
    const unsigned char *p = (const unsigned char *)value;
    size_t length = strlen(value);
    char *result = malloc(length + 4);
    size_t written = 0, current = 0;
    if (result == NULL) return -1;
    while (*p != 0 && current < end) {
        size_t units;
        size_t step = utf8_step(p, &units);
        if (current >= start && current + units <= end) {
            memcpy(result + written, p, step);
            written += step;
        } else if (units == 2) {
            /* Split astral character: emit the half that is in range. */
            uint32_t cp = ((uint32_t)(p[0] & 0x07) << 18) | ((uint32_t)(p[1] & 0x3F) << 12) |
                          ((uint32_t)(p[2] & 0x3F) << 6) | (uint32_t)(p[3] & 0x3F);
            uint32_t high = 0xD800 + ((cp - 0x10000) >> 10);
            uint32_t low = 0xDC00 + ((cp - 0x10000) & 0x3FF);
            if (current >= start && current < end) written += utf8_write3(result + written, high);
            if (current + 1 >= start && current + 1 < end) written += utf8_write3(result + written, low);
        }
        current += units;
        p += step;
    }
    result[written] = '\0';
    *out = result;
    return 0;
}

/* utf16_encode writes code units as UTF-8, pairing surrogates. */
static int utf16_encode(const uint32_t *units, size_t count, char **out) {
    char *result = malloc(count * 3 + 1);
    size_t written = 0;
    if (result == NULL) return -1;
    for (size_t i = 0; i < count; i++) {
        uint32_t cp = units[i] & 0xFFFF;
        if (cp >= 0xD800 && cp <= 0xDBFF && i + 1 < count && (units[i + 1] & 0xFFFF) >= 0xDC00 && (units[i + 1] & 0xFFFF) <= 0xDFFF) {
            cp = 0x10000 + ((cp - 0xD800) << 10) + ((units[i + 1] & 0xFFFF) - 0xDC00);
            i++;
            result[written++] = (char)(0xF0 | (cp >> 18));
            result[written++] = (char)(0x80 | ((cp >> 12) & 0x3F));
            result[written++] = (char)(0x80 | ((cp >> 6) & 0x3F));
            result[written++] = (char)(0x80 | (cp & 0x3F));
        } else if (cp < 0x80) {
            result[written++] = (char)cp;
        } else if (cp < 0x800) {
            result[written++] = (char)(0xC0 | (cp >> 6));
            result[written++] = (char)(0x80 | (cp & 0x3F));
        } else {
            written += utf8_write3(result + written, cp);
        }
    }
    result[written] = '\0';
    *out = result;
    return 0;
}

/* utf8_decode reads the code point at p (a stray byte decodes as itself). */
static uint32_t utf8_decode(const unsigned char *p, size_t *step) {
    size_t units;
    *step = utf8_step(p, &units);
    switch (*step) {
    case 2: return ((uint32_t)(p[0] & 0x1F) << 6) | (uint32_t)(p[1] & 0x3F);
    case 3: return ((uint32_t)(p[0] & 0x0F) << 12) | ((uint32_t)(p[1] & 0x3F) << 6) | (uint32_t)(p[2] & 0x3F);
    case 4: return ((uint32_t)(p[0] & 0x07) << 18) | ((uint32_t)(p[1] & 0x3F) << 12) | ((uint32_t)(p[2] & 0x3F) << 6) | (uint32_t)(p[3] & 0x3F);
    default: return p[0];
    }
}

static size_t utf8_encode(uint32_t cp, char *out) {
    if (cp < 0x80) {
        out[0] = (char)cp;
        return 1;
    }
    if (cp < 0x800) {
        out[0] = (char)(0xC0 | (cp >> 6));
        out[1] = (char)(0x80 | (cp & 0x3F));
        return 2;
    }
    if (cp < 0x10000) return utf8_write3(out, cp);
    out[0] = (char)(0xF0 | (cp >> 18));
    out[1] = (char)(0x80 | ((cp >> 12) & 0x3F));
    out[2] = (char)(0x80 | ((cp >> 6) & 0x3F));
    out[3] = (char)(0x80 | (cp & 0x3F));
    return 4;
}

/* Simple case mapping for the scripts most text uses: ASCII, Latin-1,
 * Latin Extended-A, the Vietnamese letters of Latin Extended-B and
 * Latin Extended Additional, Greek and Cyrillic. Mappings that change the
 * length (ß to SS) or depend on context are not applied. */
static int case_pair_even_upper(uint32_t cp) {
    return (cp >= 0x100 && cp <= 0x12F) || (cp >= 0x132 && cp <= 0x137) || (cp >= 0x14A && cp <= 0x177) ||
           (cp >= 0x1E00 && cp <= 0x1E95) || (cp >= 0x1EA0 && cp <= 0x1EFF);
}

static int case_pair_odd_upper(uint32_t cp) {
    return (cp >= 0x139 && cp <= 0x148) || (cp >= 0x179 && cp <= 0x17E);
}

static uint32_t unicode_to_lower(uint32_t cp) {
    if (cp >= 'A' && cp <= 'Z') return cp + 0x20;
    if (cp < 0x80) return cp;
    if (cp >= 0xC0 && cp <= 0xDE && cp != 0xD7) return cp + 0x20;
    if (case_pair_even_upper(cp)) return (cp % 2 == 0) ? cp + 1 : cp;
    if (case_pair_odd_upper(cp)) return (cp % 2 == 1) ? cp + 1 : cp;
    switch (cp) {
    case 0x178: return 0xFF;
    case 0x1A0: case 0x1AF: return cp + 1;
    case 0x386: return 0x3AC;
    case 0x38C: return 0x3CC;
    case 0x38E: case 0x38F: return cp + 0x3F;
    }
    if (cp >= 0x388 && cp <= 0x38A) return cp + 0x25;
    if (cp >= 0x391 && cp <= 0x3A9 && cp != 0x3A2) return cp + 0x20;
    if (cp >= 0x410 && cp <= 0x42F) return cp + 0x20;
    if (cp >= 0x400 && cp <= 0x40F) return cp + 0x50;
    return cp;
}

static uint32_t unicode_to_upper(uint32_t cp) {
    if (cp >= 'a' && cp <= 'z') return cp - 0x20;
    if (cp < 0x80) return cp;
    if (cp >= 0xE0 && cp <= 0xFE && cp != 0xF7) return cp - 0x20;
    if (case_pair_even_upper(cp - 1) && cp % 2 == 1) return cp - 1;
    if (case_pair_odd_upper(cp - 1) && cp % 2 == 0) return cp - 1;
    switch (cp) {
    case 0xB5: return 0x39C;
    case 0xFF: return 0x178;
    case 0x131: return 'I';
    case 0x1A1: case 0x1B0: return cp - 1;
    case 0x3AC: return 0x386;
    case 0x3CC: return 0x38C;
    case 0x3CD: case 0x3CE: return cp - 0x3F;
    case 0x3C2: return 0x3A3;
    }
    if (cp >= 0x3AD && cp <= 0x3AF) return cp - 0x25;
    if (cp >= 0x3B1 && cp <= 0x3C9) return cp - 0x20;
    if (cp >= 0x430 && cp <= 0x44F) return cp - 0x20;
    if (cp >= 0x450 && cp <= 0x45F) return cp - 0x50;
    return cp;
}

/* utf8_map_case maps every code point of value with map. */
static int utf8_map_case(const char *value, uint32_t (*map)(uint32_t), char **out) {
    size_t length = strlen(value);
    char *result = malloc(length * 2 + 4);
    size_t written = 0;
    if (result == NULL) return -1;
    for (const unsigned char *p = (const unsigned char *)value; *p != 0;) {
        size_t step;
        uint32_t cp = utf8_decode(p, &step);
        if (step == 1 && cp >= 0x80) {
            result[written++] = (char)cp; /* stray byte kept as is */
        } else {
            written += utf8_encode(map(cp), result + written);
        }
        p += step;
    }
    result[written] = '\0';
    *out = result;
    return 0;
}

/* Base letters of Latin-1 (U+00C0..U+00FF) and Latin Extended-A
 * (U+0100..U+017F), lowercase, accents removed; '.' keeps the code point. */
static const char collation_latin1[] = "aaaaaaaceeeeiiiidnooooo.ouuuuy.saaaaaaaceeeeiiiidnooooo.ouuuuy.y";
static const char collation_latin_ext_a[] =
    "aaaaaaccccccccddddeeeeeeeeeegggggggghhhhiiiiiiiiiiiijjkkk"
    "llllllllllnnnnnnnnnoooooooorrrrrrssssssssttttttuuuuuuuuuuuuwwyyyzzzzzzs";

/* collation_base is the primary collation key of a code point: its
 * lowercase base letter for accented Latin (Vietnamese included),
 * otherwise its lowercase form. */
static uint32_t collation_base(uint32_t cp) {
    if (cp >= 0xC0 && cp <= 0xFF && collation_latin1[cp - 0xC0] != '.') return (uint32_t)collation_latin1[cp - 0xC0];
    if (cp >= 0x100 && cp <= 0x17F) return (uint32_t)collation_latin_ext_a[cp - 0x100];
    if (cp == 0x1A0 || cp == 0x1A1) return 'o';
    if (cp == 0x1AF || cp == 0x1B0) return 'u';
    if (cp >= 0x1EA0 && cp <= 0x1EB7) return 'a';
    if (cp >= 0x1EB8 && cp <= 0x1EC7) return 'e';
    if (cp >= 0x1EC8 && cp <= 0x1ECB) return 'i';
    if (cp >= 0x1ECC && cp <= 0x1EE3) return 'o';
    if (cp >= 0x1EE4 && cp <= 0x1EF1) return 'u';
    if (cp >= 0x1EF2 && cp <= 0x1EF9) return 'y';
    return unicode_to_lower(cp);
}

/* collation_compare orders two strings like a root-locale collator, in
 * three levels: base letters, then accents (unaccented first), then case
 * (lowercase first). It returns -1, 0 or 1. */
static int collation_compare(const char *a, const char *b) {
    for (int level = 0; level < 3; level++) {
        const unsigned char *p = (const unsigned char *)a;
        const unsigned char *q = (const unsigned char *)b;
        while (*p != 0 && *q != 0) {
            size_t step_p, step_q;
            uint32_t x = utf8_decode(p, &step_p);
            uint32_t y = utf8_decode(q, &step_q);
            uint32_t kx, ky;
            if (level == 0) {
                kx = collation_base(x);
                ky = collation_base(y);
            } else if (level == 1) {
                kx = unicode_to_lower(x);
                ky = unicode_to_lower(y);
            } else {
                /* lowercase sorts before uppercase */
                kx = unicode_to_lower(x) == x ? 0 : 1;
                ky = unicode_to_lower(y) == y ? 0 : 1;
            }
            if (kx != ky) return kx < ky ? -1 : 1;
            p += step_p;
            q += step_q;
        }
        if (*p != 0 || *q != 0) return *p == 0 ? -1 : 1;
    }
    return 0;
}

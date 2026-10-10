#include <stddef.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int scriptgo_runtime_set_error(const char *message);

typedef struct scriptgo_symbol {
    uint64_t id;
    char *description;
    /* The string an object stores this symbol's property under, made on
     * first use (scriptgo_symbol_property_key). */
    char *property_key;
} scriptgo_symbol_t;

/* Symbols that key a property, for mapping a stored key back to its symbol
 * (Object.getOwnPropertySymbols, console output). They stay GC roots. */
typedef struct scriptgo_keyed_symbol {
    scriptgo_symbol_t *symbol;
    struct scriptgo_keyed_symbol *next;
} scriptgo_keyed_symbol_t;

static scriptgo_keyed_symbol_t *g_keyed_symbols = NULL;

typedef struct scriptgo_symbol_entry {
    char *key;
    scriptgo_symbol_t *symbol;
    struct scriptgo_symbol_entry *next;
} scriptgo_symbol_entry_t;

static uint64_t g_symbol_id_counter = 1000;
static scriptgo_symbol_entry_t *g_symbol_registry = NULL;

int scriptgo_gc_register(void *ptr, int tag, uint32_t field_count);

extern const char scriptgo_undefined_sentinel;

/* A symbol's description is absent (NULL) for Symbol() and Symbol(undefined),
 * and kept as given otherwise, including "". */
static scriptgo_symbol_t *create_symbol_internal(const char *description) {
    scriptgo_symbol_t *sym = (scriptgo_symbol_t *)malloc(sizeof(scriptgo_symbol_t));
    if (sym == NULL) return NULL;
    sym->id = ++g_symbol_id_counter;
    sym->description = (description != NULL && description != &scriptgo_undefined_sentinel) ? strdup(description) : NULL;
    sym->property_key = NULL;
    scriptgo_gc_register(sym, 11, 0);
    return sym;
}

int scriptgo_symbol_create(const char *description, void **out_symbol) {
    if (out_symbol == NULL) {
        return scriptgo_runtime_set_error("invalid argument to symbol create");
    }
    scriptgo_symbol_t *sym = create_symbol_internal(description);
    if (sym == NULL) {
        return scriptgo_runtime_set_error("failed to allocate symbol");
    }
    *out_symbol = sym;
    return 0;
}

int scriptgo_symbol_for(const char *key, void **out_symbol) {
    if (key == NULL || out_symbol == NULL) {
        return scriptgo_runtime_set_error("invalid argument to Symbol.for");
    }
    scriptgo_symbol_entry_t *curr = g_symbol_registry;
    while (curr != NULL) {
        if (strcmp(curr->key, key) == 0) {
            *out_symbol = curr->symbol;
            return 0;
        }
        curr = curr->next;
    }
    scriptgo_symbol_t *sym = create_symbol_internal(key);
    if (sym == NULL) {
        return scriptgo_runtime_set_error("failed to allocate symbol for registry");
    }
    scriptgo_symbol_entry_t *entry = (scriptgo_symbol_entry_t *)malloc(sizeof(scriptgo_symbol_entry_t));
    if (entry == NULL) {
        return scriptgo_runtime_set_error("failed to allocate symbol registry entry");
    }
    entry->key = strdup(key);
    entry->symbol = sym;
    entry->next = g_symbol_registry;
    g_symbol_registry = entry;

    *out_symbol = sym;
    return 0;
}

int scriptgo_symbol_key_for(void *symbol, char **out_key) {
    if (symbol == NULL || out_key == NULL) {
        return scriptgo_runtime_set_error("invalid argument to Symbol.keyFor");
    }
    scriptgo_symbol_t *sym = (scriptgo_symbol_t *)symbol;
    scriptgo_symbol_entry_t *curr = g_symbol_registry;
    while (curr != NULL) {
        if (curr->symbol->id == sym->id) {
            *out_key = strdup(curr->key);
            return 0;
        }
        curr = curr->next;
    }
    *out_key = strdup("undefined");
    return 0;
}

int scriptgo_symbol_description(void *symbol, char **out_description) {
    if (symbol == NULL || out_description == NULL) {
        return scriptgo_runtime_set_error("invalid argument to symbol description");
    }
    scriptgo_symbol_t *sym = (scriptgo_symbol_t *)symbol;
    /* An absent description reads as undefined. */
    *out_description = sym->description != NULL ? strdup(sym->description) : (char *)&scriptgo_undefined_sentinel;
    return 0;
}

int scriptgo_symbol_to_string(void *symbol, char **out_string) {
    if (symbol == NULL || out_string == NULL) {
        return scriptgo_runtime_set_error("invalid argument to symbol toString");
    }
    scriptgo_symbol_t *sym = (scriptgo_symbol_t *)symbol;
    char buffer[256];
    if (sym->description != NULL && sym->description[0] != '\0') {
        snprintf(buffer, sizeof(buffer), "Symbol(%s)", sym->description);
    } else {
        snprintf(buffer, sizeof(buffer), "Symbol()");
    }
    *out_string = strdup(buffer);
    return 0;
}

int scriptgo_gc_add_root(void *ptr);

/* A symbol-keyed property is stored under "\x01" followed by the symbol's
 * id. No string property key starts with \x01 in practice, and every
 * key enumeration (Object.keys, JSON, for..in) skips such keys. The key is
 * borrowed: it lives as long as the symbol, which stays rooted. */
int scriptgo_symbol_property_key(void *symbol, char **out_key) {
    if (symbol == NULL || out_key == NULL) {
        return scriptgo_runtime_set_error("invalid argument to symbol property key");
    }
    scriptgo_symbol_t *sym = (scriptgo_symbol_t *)symbol;
    if (sym->property_key == NULL) {
        char buffer[32];
        snprintf(buffer, sizeof(buffer), "\x01%llu", (unsigned long long)sym->id);
        scriptgo_keyed_symbol_t *entry = malloc(sizeof(*entry));
        char *key = strdup(buffer);
        if (entry == NULL || key == NULL) {
            free(entry);
            free(key);
            return scriptgo_runtime_set_error("failed to allocate symbol property key");
        }
        sym->property_key = key;
        entry->symbol = sym;
        entry->next = g_keyed_symbols;
        g_keyed_symbols = entry;
        scriptgo_gc_add_root(sym);
    }
    *out_key = sym->property_key;
    return 0;
}

/* scriptgo_symbol_for_property_key is the symbol a stored property key
 * names, or NULL when the key is not a symbol key. */
void *scriptgo_symbol_for_property_key(const char *key, size_t length) {
    if (key == NULL || length < 2 || key[0] != '\x01') return NULL;
    for (scriptgo_keyed_symbol_t *entry = g_keyed_symbols; entry != NULL; entry = entry->next) {
        const char *candidate = entry->symbol->property_key;
        if (strlen(candidate) == length && memcmp(candidate, key, length) == 0) return entry->symbol;
    }
    return NULL;
}

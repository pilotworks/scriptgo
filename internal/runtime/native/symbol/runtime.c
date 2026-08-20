#include <stddef.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int scriptgo_runtime_set_error(const char *message);

typedef struct scriptgo_symbol {
    uint64_t id;
    char *description;
} scriptgo_symbol_t;

typedef struct scriptgo_symbol_entry {
    char *key;
    scriptgo_symbol_t *symbol;
    struct scriptgo_symbol_entry *next;
} scriptgo_symbol_entry_t;

static uint64_t g_symbol_id_counter = 1000;
static scriptgo_symbol_entry_t *g_symbol_registry = NULL;

static scriptgo_symbol_t *create_symbol_internal(const char *description) {
    scriptgo_symbol_t *sym = (scriptgo_symbol_t *)malloc(sizeof(scriptgo_symbol_t));
    if (sym == NULL) return NULL;
    sym->id = ++g_symbol_id_counter;
    sym->description = (description && description[0] != '\0') ? strdup(description) : NULL;
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
    if (sym->description != NULL) {
        *out_description = strdup(sym->description);
    } else {
        *out_description = strdup("undefined");
    }
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

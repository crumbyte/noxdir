//go:build darwin

#include "../include/arena.h"

#include <stdlib.h>
#include <string.h>

Arena *new_arena(const size_t capacity, const bool dynamic) {
    Arena *arena = malloc(sizeof(Arena));
    if (arena == NULL) {
        return NULL;
    }

    arena->capacity = capacity;
    arena->offset = 0;
    arena->dynamic = dynamic;
    arena->region = malloc(arena->capacity);

    if (arena->region == NULL) {
        free(arena);

        return NULL;
    }

    return arena;
}

static size_t align_up(const size_t offset, const size_t align) {
    return offset + (align - 1) & ~(align - 1);
}

void *arena_alloc(Arena **arena_ptr, const size_t size, const size_t align) {
    Arena *arena = *arena_ptr;

    if (arena == NULL || arena->region == NULL) {
        return NULL;
    }

    const size_t offset = align_up(arena->offset, align);

    if (arena->capacity < offset + size) {
        if (!arena->dynamic) {
            return NULL;
        }

        size_t grow_size = arena->capacity * 2;

        if (grow_size < offset + size)
            grow_size = offset + size;

        arena = arena_grow_safe(arena, grow_size);
        if (arena == NULL) {
            return NULL;
        }
    }

    char *region_ptr = arena->region + offset;
    arena->offset = offset + size;

    return region_ptr;
}

Arena *arena_grow(Arena *arena, const size_t size) {
    if (arena == NULL || arena->region == NULL) {
        return NULL;
    }

    if (size < arena->capacity) {
        return arena;
    }

    char *new_region = realloc(arena->region, size);
    if (new_region == NULL) {
        return NULL;
    }

    arena->region = new_region;

    return arena;
}

Arena *arena_grow_safe(Arena *arena, const size_t size) {
    if (arena == NULL || arena->region == NULL) {
        return NULL;
    }

    if (size < arena->capacity) {
        return arena;
    }

    char *new_region = malloc(size);
    if (new_region == NULL) {
        return NULL;
    }

    memcpy(new_region, arena->region, arena->offset);
    free(arena->region);

    arena->region = new_region;
    arena->capacity = size;

    return arena;
}

void clear_arena(Arena *arena) {
    if (arena != NULL) {
        arena->offset = 0;
    }
}

void free_arena(Arena *arena) {
    if (arena != NULL) {
        free(arena->region);
        free(arena);
    }
}

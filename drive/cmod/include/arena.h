//go:build darwin

#ifndef ARENA_H
#define ARENA_H

#include <stddef.h>
#include <stdbool.h>

typedef struct {
    char *region;
    size_t offset;
    size_t capacity;
    bool dynamic;
} Arena;

Arena *new_arena(size_t capacity, bool dynamic);

void *arena_alloc(Arena **arena_ptr, size_t size, size_t align);

Arena *arena_grow(Arena *arena, size_t size);

Arena *arena_grow_safe(Arena *arena, size_t size);

void clear_arena(Arena *arena);

void free_arena(Arena *arena);

#endif

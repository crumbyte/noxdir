//go:build darwin

#ifndef READDIR_H
#define READDIR_H

#include <stdint.h>

typedef struct {
    char     name[256];
    uint64_t ino;
    int64_t  dev;
    int      isDir;
    int64_t  size;
    int64_t  modSec;
    int64_t  modNSec;
} FileInfoC;

int read_dir(const char* path, FileInfoC** out, int* count);

void free_result(FileInfoC* arr);

#endif

//go:build darwin

#ifndef READDIR_H
#define READDIR_H

#include <sys/stat.h>
#include <stdint.h>

typedef struct {
    char     name[256];
    int      isDir;
    int64_t  size;
    int64_t  modSec;
    int64_t  modNSec;
} FileInfoC;

int ReadDirC(const char* path, FileInfoC** out, int* count);

void FreeFileInfoC(FileInfoC* arr);

#endif

//go:build darwin

#include "readdir.h"
#include <dirent.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <errno.h>
#include <unistd.h>
#include <fcntl.h>
#include <stdalign.h>

#ifdef _DARWIN_FEATURE_64_BIT_INODE

int darwin_legacy_getdirentries(int, char *, int, long *) __asm("_getdirentries");

#define getdirentries darwin_legacy_getdirentries

struct darwin_legacy_dirent {
    __uint32_t d_ino;
    __uint16_t d_reclen;
    __uint8_t d_type;
    __uint8_t d_namlen;
    char d_name[__DARWIN_MAXNAMLEN + 1];
};

#define dirent darwin_legacy_dirent

#endif

void get_fstat_at(const int fd, const struct dirent *de, FileInfoC *fi) {
    struct stat st;

    if (fstatat(fd, de->d_name, &st, AT_SYMLINK_NOFOLLOW) != 0) {
        return;
    }

    fi->isDir = S_ISDIR(st.st_mode);
    fi->size = st.st_size;
    fi->dev = st.st_dev;
    fi->ino = st.st_ino;
    fi->modSec = st.st_mtimespec.tv_sec;
    fi->modNSec = st.st_mtimespec.tv_nsec;
}

int read_dir(const char *path, FileInfoC **out, int *count) {
    const int fd = open(path, O_RDONLY | O_DIRECTORY);
    if (fd < 0) {
        return errno;
    }

    int capacity = 64;
    int n = 0;

    FileInfoC *result = malloc(capacity * sizeof(FileInfoC));
    if (result == NULL) {
        close(fd);

        return ENOMEM;
    }

    char buf[4096];
    long base = 0;
    ssize_t bytesRead;

    while ((bytesRead = getdirentries(fd, buf, sizeof(buf), &base)) > 0) {
        char *p = buf;
        while (p < buf + bytesRead) {
            const struct dirent *de = (struct dirent *) p;
            p += de->d_reclen;

            if (de->d_name[0] == '.' &&
                (de->d_name[1] == '\0' ||
                 (de->d_name[1] == '.' && de->d_name[2] == '\0'))) {
                continue;
                 }

            if (de->d_name[0] == '\0' || de->d_name[0] < 0) {
                continue;
            }

            if (n == capacity) {
                capacity *= 2;
                FileInfoC *tmp = realloc(result, capacity * sizeof(FileInfoC));
                if (!tmp) {
                    free(result);
                    close(fd);
                    return ENOMEM;
                }
                result = tmp;
            }

            FileInfoC *fi = &result[n];

            memset(fi, 0, sizeof(FileInfoC));
            strncpy(fi->name, de->d_name, sizeof(fi->name)-1);
            get_fstat_at(fd, de, fi);

            n++;
        }
    }

    if (bytesRead < 0) {
        free(result);
        close(fd);
        return errno;
    }

    close(fd);

    *out = result;
    *count = n;
    return 0;
}

void free_result(FileInfoC *arr) {
    free(arr);
}

#ifndef GRIMOIRE_VECTOR_H
#define GRIMOIRE_VECTOR_H

#include <stddef.h>
#include <stdint.h>

#ifdef _WIN32
#define GV_API __declspec(dllimport)
#else
#define GV_API
#endif

typedef struct GvSearchResult {
    uint64_t id_offset;
    uint32_t id_len;
    uint32_t reserved0;
    float score;
    uint32_t reserved1;
    uint64_t index;
} GvSearchResult;

GV_API uint32_t gv_abi_version(void);
GV_API size_t gv_last_error_message(uint8_t *buffer, size_t capacity);
GV_API int32_t gv_object_exists(const uint8_t*, size_t, const uint8_t*, size_t, const uint8_t*, size_t, uint8_t*);
GV_API int32_t gv_ingest_jsonl(const uint8_t*, size_t, const uint8_t*, size_t, const uint8_t*, size_t, uint64_t*);
GV_API int32_t gv_materialize_jsonl(const uint8_t*, size_t, const uint8_t*, size_t, const uint8_t*, size_t, const uint8_t*, size_t, uint8_t*, size_t, size_t*);
GV_API int32_t gv_open_snapshot(const uint8_t*, size_t, uint64_t*);
GV_API int32_t gv_close_snapshot(uint64_t);
GV_API int32_t gv_snapshot_info(uint64_t, uint32_t*, uint64_t*, uint8_t*, size_t, size_t*);
GV_API int32_t gv_search(uint64_t, const float*, size_t, size_t, GvSearchResult*, size_t, uint8_t*, size_t, size_t*, size_t*);

#endif

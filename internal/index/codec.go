package index

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

const (
	shardMagic      = "GRSH"
	shardVersion    = byte(1)
	shardHeaderSize = len(shardMagic) + 1 + 4
)

func shardName(path string) string {
	digest := sha256.Sum256([]byte(path))
	return hex.EncodeToString(digest[:1])
}

func encodeShard(records map[string][]byte) ([]byte, error) {
	paths := make([]string, 0, len(records))
	size := shardHeaderSize
	for path, value := range records {
		if err := validateRecordPath(path); err != nil {
			return nil, err
		}
		if uint64(len(path)) > uint64(^uint32(0)) || uint64(len(value)) > uint64(^uint32(0)) {
			return nil, fmt.Errorf("index record is too large: %q", path)
		}
		paths = append(paths, path)
		size += 8 + len(path) + len(value)
	}
	sort.Strings(paths)

	encoded := make([]byte, size)
	copy(encoded, shardMagic)
	encoded[len(shardMagic)] = shardVersion
	binary.BigEndian.PutUint32(encoded[len(shardMagic)+1:], uint32(len(paths)))
	offset := shardHeaderSize
	for _, path := range paths {
		value := records[path]
		binary.BigEndian.PutUint32(encoded[offset:], uint32(len(path)))
		offset += 4
		binary.BigEndian.PutUint32(encoded[offset:], uint32(len(value)))
		offset += 4
		copy(encoded[offset:], path)
		offset += len(path)
		copy(encoded[offset:], value)
		offset += len(value)
	}
	return encoded, nil
}

func decodeShard(data []byte) (map[string][]byte, error) {
	if len(data) < shardHeaderSize {
		return nil, fmt.Errorf("malformed index shard: truncated header")
	}
	if string(data[:len(shardMagic)]) != shardMagic {
		return nil, fmt.Errorf("malformed index shard: invalid magic")
	}
	if data[len(shardMagic)] != shardVersion {
		return nil, fmt.Errorf("unsupported index shard version %d", data[len(shardMagic)])
	}

	offset := len(shardMagic) + 1
	count := binary.BigEndian.Uint32(data[offset:])
	offset += 4
	if uint64(count) > uint64((len(data)-offset)/8) {
		return nil, fmt.Errorf("malformed index shard: impossible record count")
	}
	records := make(map[string][]byte, count)
	for index := uint32(0); index < count; index++ {
		pathLength, err := readLength(data, &offset, "path")
		if err != nil {
			return nil, err
		}
		valueLength, err := readLength(data, &offset, "value")
		if err != nil {
			return nil, err
		}
		pathBytes, err := readField(data, &offset, pathLength, "path")
		if err != nil {
			return nil, err
		}
		value, err := readField(data, &offset, valueLength, "value")
		if err != nil {
			return nil, err
		}
		path := string(pathBytes)
		if err := validateRecordPath(path); err != nil {
			return nil, err
		}
		if _, exists := records[path]; exists {
			return nil, fmt.Errorf("malformed index shard: duplicate path %q", path)
		}
		records[path] = append([]byte(nil), value...)
	}
	if offset != len(data) {
		return nil, fmt.Errorf("malformed index shard: trailing data")
	}
	return records, nil
}

func validateRecordPath(path string) error {
	if path == "" || !utf8.ValidString(path) {
		return fmt.Errorf("invalid index path %q", path)
	}
	if strings.HasPrefix(path, "/") || isDriveAbsolute(path) || strings.ContainsAny(path, "\\\x00") {
		return fmt.Errorf("invalid index path %q", path)
	}
	for _, segment := range strings.Split(path, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return fmt.Errorf("invalid index path %q", path)
		}
	}
	return nil
}

func isDriveAbsolute(path string) bool {
	return len(path) >= 3 && path[1] == ':' && path[2] == '/'
}

func isShardName(name string) bool {
	if len(name) != 2 {
		return false
	}
	return strings.ContainsRune("0123456789abcdef", rune(name[0])) &&
		strings.ContainsRune("0123456789abcdef", rune(name[1]))
}

func readLength(data []byte, offset *int, field string) (uint32, error) {
	if len(data)-*offset < 4 {
		return 0, fmt.Errorf("malformed index shard: truncated %s length", field)
	}
	length := binary.BigEndian.Uint32(data[*offset:])
	*offset += 4
	return length, nil
}

func readField(data []byte, offset *int, length uint32, field string) ([]byte, error) {
	if uint64(length) > uint64(len(data)-*offset) {
		return nil, fmt.Errorf("malformed index shard: truncated %s", field)
	}
	end := *offset + int(length)
	value := data[*offset:end]
	*offset = end
	return value, nil
}

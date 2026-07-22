package index

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

const (
	fileMagic   = "GRFL"
	fileVersion = byte(2)
	fileHeader  = len(fileMagic) + 1 + 32 + 8 + 4
)

func encodeFile(file FileRecord) ([]byte, error) {
	if err := validateRecordPath(file.Path); err != nil {
		return nil, err
	}
	hash, err := hex.DecodeString(file.Hash)
	if err != nil || len(hash) != 32 {
		return nil, fmt.Errorf("invalid content hash for %q", file.Path)
	}
	if file.Size < 0 || uint64(len(file.Chunks)) > uint64(^uint32(0)) {
		return nil, fmt.Errorf("invalid file metadata for %q", file.Path)
	}

	size := fileHeader
	for _, chunk := range file.Chunks {
		if chunk.Path != "" && chunk.Path != file.Path {
			return nil, fmt.Errorf("chunk %q belongs to %q, not %q", chunk.ID, chunk.Path, file.Path)
		}
		if !fitsUint32(len(chunk.ID)) || !fitsUint32(len(chunk.Text)) ||
			!fitsUint32(chunk.StartLine) || !fitsUint32(chunk.EndLine) || !fitsUint32(chunk.TokenCount) {
			return nil, fmt.Errorf("chunk %q metadata is too large", chunk.ID)
		}
		size += 20 + len(chunk.ID) + len(chunk.Text)
	}

	encoded := make([]byte, size)
	copy(encoded, fileMagic)
	encoded[len(fileMagic)] = fileVersion
	offset := len(fileMagic) + 1
	copy(encoded[offset:], hash)
	offset += 32
	binary.BigEndian.PutUint64(encoded[offset:], uint64(file.Size))
	offset += 8
	binary.BigEndian.PutUint32(encoded[offset:], uint32(len(file.Chunks)))
	offset += 4
	for _, chunk := range file.Chunks {
		binary.BigEndian.PutUint32(encoded[offset:], uint32(len(chunk.ID)))
		offset += 4
		binary.BigEndian.PutUint32(encoded[offset:], uint32(chunk.StartLine))
		offset += 4
		binary.BigEndian.PutUint32(encoded[offset:], uint32(chunk.EndLine))
		offset += 4
		binary.BigEndian.PutUint32(encoded[offset:], uint32(chunk.TokenCount))
		offset += 4
		binary.BigEndian.PutUint32(encoded[offset:], uint32(len(chunk.Text)))
		offset += 4
		copy(encoded[offset:], chunk.ID)
		offset += len(chunk.ID)
		copy(encoded[offset:], chunk.Text)
		offset += len(chunk.Text)
	}
	return encoded, nil
}

func decodeFile(path string, data []byte) (FileRecord, error) {
	if len(data) < fileHeader {
		return FileRecord{}, fmt.Errorf("malformed file record %q: truncated header", path)
	}
	if string(data[:len(fileMagic)]) != fileMagic {
		return FileRecord{}, fmt.Errorf("malformed file record %q: invalid magic", path)
	}
	if data[len(fileMagic)] != fileVersion {
		return FileRecord{}, fmt.Errorf(
			"%w: file record %q uses version %d",
			ErrIncompatibleIndex,
			path,
			data[len(fileMagic)],
		)
	}
	offset := len(fileMagic) + 1
	hash := hex.EncodeToString(data[offset : offset+32])
	offset += 32
	size := binary.BigEndian.Uint64(data[offset:])
	offset += 8
	if size > uint64(^uint64(0)>>1) {
		return FileRecord{}, fmt.Errorf("malformed file record %q: invalid size", path)
	}
	count := binary.BigEndian.Uint32(data[offset:])
	offset += 4
	if uint64(count) > uint64((len(data)-offset)/20) {
		return FileRecord{}, fmt.Errorf("malformed file record %q: impossible chunk count", path)
	}

	chunks := make([]Chunk, 0, count)
	for index := uint32(0); index < count; index++ {
		idLength, err := readFileUint32(data, &offset, path, "chunk id length")
		if err != nil {
			return FileRecord{}, err
		}
		start, err := readFileUint32(data, &offset, path, "start line")
		if err != nil {
			return FileRecord{}, err
		}
		end, err := readFileUint32(data, &offset, path, "end line")
		if err != nil {
			return FileRecord{}, err
		}
		tokens, err := readFileUint32(data, &offset, path, "token count")
		if err != nil {
			return FileRecord{}, err
		}
		textLength, err := readFileUint32(data, &offset, path, "chunk text length")
		if err != nil {
			return FileRecord{}, err
		}
		id, err := readFileField(data, &offset, idLength, path, "chunk id")
		if err != nil {
			return FileRecord{}, err
		}
		text, err := readFileField(data, &offset, textLength, path, "chunk text")
		if err != nil {
			return FileRecord{}, err
		}
		if end < start {
			return FileRecord{}, fmt.Errorf("malformed file record %q: invalid line range", path)
		}
		chunks = append(chunks, Chunk{
			ID: string(id), Path: path, StartLine: int(start), EndLine: int(end),
			TokenCount: int(tokens), Text: string(text),
		})
	}
	if offset != len(data) {
		return FileRecord{}, fmt.Errorf("malformed file record %q: trailing data", path)
	}
	return FileRecord{Path: path, Hash: hash, Size: int64(size), Chunks: chunks}, nil
}

func fitsUint32(value int) bool {
	return value >= 0 && uint64(value) <= uint64(^uint32(0))
}

func readFileUint32(data []byte, offset *int, path, field string) (uint32, error) {
	if len(data)-*offset < 4 {
		return 0, fmt.Errorf("malformed file record %q: truncated %s", path, field)
	}
	value := binary.BigEndian.Uint32(data[*offset:])
	*offset += 4
	return value, nil
}

func readFileField(data []byte, offset *int, length uint32, path, field string) ([]byte, error) {
	if uint64(length) > uint64(len(data)-*offset) {
		return nil, fmt.Errorf("malformed file record %q: truncated %s", path, field)
	}
	end := *offset + int(length)
	value := data[*offset:end]
	*offset = end
	return value, nil
}

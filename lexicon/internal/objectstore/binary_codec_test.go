package objectstore

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
)

func TestBinaryObjectRoundTripIsDeterministic(t *testing.T) {
	object := binaryFixture(12)
	first, err := encodeBinaryObject(object)
	if err != nil {
		t.Fatal(err)
	}
	second, err := encodeBinaryObject(object)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("repeated binary encoding changed bytes")
	}
	if !isBinaryObject(first) {
		t.Fatal("encoded object does not have the Lexicon binary magic")
	}
	decoded, err := decodeBinaryObject(first)
	if err != nil {
		t.Fatal(err)
	}
	assertObjectEquivalent(t, decoded, object)
}

func TestBinaryObjectRejectsTruncationAndTrailingBytes(t *testing.T) {
	encoded, err := encodeBinaryObject(binaryFixture(2))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := decodeBinaryObject(encoded[:len(encoded)-1]); err == nil {
		t.Fatal("expected truncated object error")
	}
	if _, err := decodeBinaryObject(append(append([]byte(nil), encoded...), 0)); err == nil {
		t.Fatal("expected trailing byte error")
	}
}

func TestBinaryObjectRejectsInvalidUTF8(t *testing.T) {
	encoded, err := encodeBinaryObject(binaryFixture(1))
	if err != nil {
		t.Fatal(err)
	}
	reader := binaryObjectReader{data: encoded}
	if !reader.magic() {
		t.Fatal("missing binary magic")
	}
	if _, err := reader.uvarint("object version"); err != nil {
		t.Fatal(err)
	}
	if _, err := reader.uvarint("schema version"); err != nil {
		t.Fatal(err)
	}
	count, err := reader.count("string table", maxBinaryStrings)
	if err != nil || count < 2 {
		t.Fatalf("string table: count=%d err=%v", count, err)
	}
	if _, err := reader.bytes("empty string", maxBinaryStringSize); err != nil {
		t.Fatal(err)
	}
	length, size := binary.Uvarint(encoded[reader.position:])
	if size <= 0 || length == 0 {
		t.Fatal("first non-empty string is unavailable")
	}
	corrupt := append([]byte(nil), encoded...)
	corrupt[reader.position+size] = 0xff
	if _, err := decodeBinaryObject(corrupt); err == nil {
		t.Fatal("expected invalid UTF-8 error")
	}
}

func TestStoreReadsLegacyJSONObject(t *testing.T) {
	store := Store{Root: t.TempDir()}
	object := binaryFixture(1)
	data, err := json.Marshal(object)
	if err != nil {
		t.Fatal(err)
	}
	id := digest("lexicon:fact-object:v1\x00", data)
	if err := writeImmutable(store.ObjectPath(id), append(data, '\n')); err != nil {
		t.Fatal(err)
	}
	loaded, err := store.LoadObject(id)
	if err != nil {
		t.Fatal(err)
	}
	assertObjectEquivalent(t, loaded, object)
}

func TestStoreWritesBinaryObjects(t *testing.T) {
	store := Store{Root: t.TempDir()}
	object := binaryFixture(3)
	id, err := store.WriteObject(object)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(store.ObjectPath(id))
	if err != nil {
		t.Fatal(err)
	}
	if !isBinaryObject(data) {
		t.Fatal("object store wrote JSON instead of binary")
	}
	loaded, err := store.LoadObject(id)
	if err != nil {
		t.Fatal(err)
	}
	object.Version = ObjectVersion
	assertObjectEquivalent(t, loaded, object)
}

func TestBinaryObjectIsSmallerForRepeatedFacts(t *testing.T) {
	object := binaryFixture(200)
	binaryData, err := encodeBinaryObject(object)
	if err != nil {
		t.Fatal(err)
	}
	jsonData, err := json.Marshal(object)
	if err != nil {
		t.Fatal(err)
	}
	reduction := 100 * (1 - float64(len(binaryData))/float64(len(jsonData)))
	t.Logf("binary bytes = %d, JSON bytes = %d, reduction = %.1f%%", len(binaryData), len(jsonData), reduction)
	if len(binaryData) >= len(jsonData) {
		t.Fatalf("binary bytes = %d, JSON bytes = %d", len(binaryData), len(jsonData))
	}
}

func FuzzDecodeBinaryObject(f *testing.F) {
	seed, err := encodeBinaryObject(binaryFixture(2))
	if err != nil {
		f.Fatal(err)
	}
	f.Add(seed)
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = decodeBinaryObject(data)
	})
}

func BenchmarkFactObjectCodecs(b *testing.B) {
	object := binaryFixture(500)
	jsonData, err := json.Marshal(object)
	if err != nil {
		b.Fatal(err)
	}
	typed, err := parseTypedRecords(object.Records)
	if err != nil {
		b.Fatal(err)
	}
	typedObject := object
	typedObject.typed = &typed
	binaryData, err := encodeBinaryObject(typedObject)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportMetric(float64(len(jsonData)), "json-bytes")
	b.ReportMetric(float64(len(binaryData)), "binary-bytes")
	b.Run("json-encode", func(b *testing.B) {
		for b.Loop() {
			if _, err := json.Marshal(object); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("binary-encode-raw", func(b *testing.B) {
		for b.Loop() {
			if _, err := encodeBinaryObject(object); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("binary-encode-typed", func(b *testing.B) {
		for b.Loop() {
			if _, err := encodeBinaryObject(typedObject); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("json-decode", func(b *testing.B) {
		for b.Loop() {
			var decoded FactObject
			if err := json.Unmarshal(jsonData, &decoded); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("binary-decode", func(b *testing.B) {
		for b.Loop() {
			if _, err := decodeBinaryObject(binaryData); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func binaryFixture(count int) FactObject {
	records := make([]json.RawMessage, 0, count*2+1)
	for index := range count {
		id := fmt.Sprintf("sha256:node-%04d", index)
		records = append(records, json.RawMessage(fmt.Sprintf(
			`{"attributes":{"visibility":"public"},"id":%q,"kind":"function","name":%q,"owner":"src/main.go","path":"src/main.go","qualified_name":%q,"record":"node","span":{"end_column":2,"end_line":%d,"path":"src/main.go","start_column":1,"start_line":%d}}`,
			id, fmt.Sprintf("function%d", index), fmt.Sprintf("demo.function%d", index), index+1, index+1,
		)))
	}
	for index := 1; index < count; index++ {
		records = append(records, json.RawMessage(fmt.Sprintf(
			`{"owner":"src/main.go","record":"edge","relation":"calls","source":"sha256:node-%04d","target":"sha256:node-%04d"}`,
			index-1, index,
		)))
	}
	if count > 0 {
		records = append(records, json.RawMessage(
			`{"candidate_name":"dynamic","expression":"dynamic()","owner":"src/main.go","reason":"dynamic-target","record":"unresolved","relation":"calls","source":"sha256:node-0000"}`,
		))
	}
	return FactObject{
		Version: ObjectVersion, Language: "go", Owner: "src/main.go",
		SourceContentID: "sha256:content", AdapterVersion: "1.2.3",
		SchemaVersion: 1, AnalysisConfigID: "sha256:config", Records: records,
	}
}

func assertObjectEquivalent(t *testing.T, got, want FactObject) {
	t.Helper()
	if got.Version != want.Version || got.Language != want.Language || got.Owner != want.Owner ||
		got.SourceContentID != want.SourceContentID || got.AdapterVersion != want.AdapterVersion ||
		got.SchemaVersion != want.SchemaVersion || got.AnalysisConfigID != want.AnalysisConfigID {
		t.Fatalf("metadata mismatch: got %#v want %#v", got, want)
	}
	if len(got.Records) != len(want.Records) {
		t.Fatalf("records = %d, want %d", len(got.Records), len(want.Records))
	}
	for index := range got.Records {
		var gotValue, wantValue any
		if err := json.Unmarshal(got.Records[index], &gotValue); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(want.Records[index], &wantValue); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(gotValue, wantValue) {
			t.Fatalf("record %d mismatch:\n got %s\nwant %s", index, got.Records[index], want.Records[index])
		}
	}
}

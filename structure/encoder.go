package structure

import (
	"encoding/binary"
	"fmt"
	"io"
	"sync"
	"unsafe"
)

type Encoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

var bufferPool = sync.Pool{
	New: func() any {
		// 48 bytes for 2 int64 and 4 uint64, 1 byte for dir flag, and 4 bytes
		// for the number of child entries.
		buf := make([]byte, 8*6+1+4)

		return &buf
	},
}

func (e *Encoder) Encode(v any) error {
	entry, ok := v.(*Entry)
	if !ok {
		return fmt.Errorf("structure: encoding %T: not *Entry", v)
	}

	buf, ok := bufferPool.Get().(*[]byte)
	if !ok {
		buf = new([]byte)
	}

	defer bufferPool.Put(buf)

	//nolint:gosec // too bad
	{
		binary.LittleEndian.PutUint64((*buf)[0:], uint64(entry.ModTime))
		binary.LittleEndian.PutUint64((*buf)[8:], uint64(entry.Size))
	}

	binary.LittleEndian.PutUint64((*buf)[16:], entry.LocalDirs)
	binary.LittleEndian.PutUint64((*buf)[24:], entry.LocalFiles)
	binary.LittleEndian.PutUint64((*buf)[32:], entry.TotalDirs)
	binary.LittleEndian.PutUint64((*buf)[40:], entry.TotalFiles)
	(*buf)[48] = 0

	if entry.IsDir {
		(*buf)[48] = 1
	}

	//nolint:gosec // ...
	binary.LittleEndian.PutUint32((*buf)[49:], uint32(len(entry.Child)))

	if _, err := e.w.Write(*buf); err != nil {
		return fmt.Errorf("structure: write buffer: %w", err)
	}

	if err := e.writeString(entry.Path); err != nil {
		return fmt.Errorf("structure: write path: %w", err)
	}

	for _, child := range entry.Child {
		if err := e.Encode(child); err != nil {
			return err
		}
	}

	return nil
}

func (e *Encoder) writeString(s string) error {
	//nolint:gosec // ...
	err := binary.Write(e.w, binary.LittleEndian, int32(len(s)))
	if err != nil {
		return err
	}

	_, err = e.w.Write(unsafeBytes(s))

	return err
}

type Decoder struct {
	r io.Reader
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

func (d *Decoder) Decode(v any) error {
	entry, ok := v.(*Entry)
	if !ok {
		return fmt.Errorf("decoding %T: not *Entry", v)
	}

	var err error

	buf, ok := bufferPool.Get().(*[]byte)
	if !ok {
		buf = new([]byte)
	}

	if _, err = io.ReadFull(d.r, *buf); err != nil {
		return fmt.Errorf("structure: read buffer: %w", err)
	}

	//nolint:gosec // how could I
	{
		entry.ModTime = int64(binary.LittleEndian.Uint64((*buf)[0:]))
		entry.Size = int64(binary.LittleEndian.Uint64((*buf)[8:]))
	}

	entry.LocalDirs = uint64(binary.LittleEndian.Uint32((*buf)[16:]))
	entry.LocalFiles = uint64(binary.LittleEndian.Uint32((*buf)[24:]))
	entry.TotalDirs = uint64(binary.LittleEndian.Uint32((*buf)[32:]))
	entry.TotalFiles = uint64(binary.LittleEndian.Uint32((*buf)[40:]))
	entry.IsDir = (*buf)[48] == 1

	childCount := binary.LittleEndian.Uint32((*buf)[49:])

	bufferPool.Put(buf)

	entry.Path, err = d.readString()
	if err != nil {
		return fmt.Errorf("decoding path: %w", err)
	}

	entry.Child = make([]*Entry, 0, childCount)

	for range childCount {
		child := &Entry{}

		if err = d.Decode(child); err != nil {
			return err
		}

		entry.Child = append(entry.Child, child)
	}

	return nil
}

func (d *Decoder) readString() (string, error) {
	var l int32

	if err := binary.Read(d.r, binary.LittleEndian, &l); err != nil {
		return "", err
	}

	b := make([]byte, l)
	_, err := io.ReadFull(d.r, b)

	return unsafeString(b), err
}

func unsafeString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func unsafeBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

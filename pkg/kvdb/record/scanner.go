package record

import (
	"bufio"
	"errors"
	"io"
)

// Scanner implements the bufio.Scanner interface with a custom split function
// for tokenizing Records
type Scanner struct {
	*bufio.Scanner
}

// NewScanner returns a new Record-Scanner for the reader. maxScanTokenSize is
// the largest possible size that the scanner will buffer and should be set to
// at least the byte size of the key and value combined.
func NewScanner(r io.Reader, maxScanTokenSize int) (*Scanner, error) {
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 4096)
	scanner.Buffer(buf, maxScanTokenSize+metaLength)
	scanner.Split(split)
	return &Scanner{scanner}, nil
}

func split(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	r, err := FromBytes(data)
	if errors.Is(err, ErrInsufficientData) {
		return 0, nil, nil
	}

	if err != nil {
		return 0, nil, err
	}

	adv := r.Size()

	return adv, data[:adv], nil
}

// Record returns the most recent record generated by a call to scan, since
// Scanner keeps track of any error encountered we can ignore them here.
func (r *Scanner) Record() *Record {
	data := r.Bytes()
	record, _ := FromBytes(data)
	return record
}

type scannerCursor struct {
	record  *Record
	scanner *Scanner
}

func newScannerCursor(scanner *Scanner) *scannerCursor {
	return &scannerCursor{
		record:  nil,
		scanner: scanner,
	}
}

func (s *scannerCursor) key() *string {
	if s.record == nil {
		return nil
	}

	key := s.record.Key()
	return &key
}

func (s *scannerCursor) next() {
	if s != nil {
		s.scanner.Scan()
		s.record = s.scanner.Record()
	}
}

func (s *scannerCursor) write(w io.Writer) {
	if s.record != nil {
		s.record.Write(w)
	}
}

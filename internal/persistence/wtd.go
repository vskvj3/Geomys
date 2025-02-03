package persistence

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Singleton persistence instance and mutex for thread safety
var (
	instance *Persistence
	mu       sync.Mutex
)

// Persistence manages binary log storage
type Persistence struct {
	file *os.File
	mu   sync.Mutex
}

// NewPersistence initializes persistence storage
func NewPersistence() (*Persistence, error) {
	homeDir, _ := os.UserHomeDir()
	persistenceFile := filepath.Join(homeDir, ".geomys", "binlog.dat")

	// Open the file in append mode, create if needed
	file, err := os.OpenFile(persistenceFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	return &Persistence{file: file}, nil
}

// CreateOrReplacePersistence returns an existing persistence instance or creates a new one
func CreateOrReplacePersistence() (*Persistence, error) {
	mu.Lock()
	defer mu.Unlock()

	// If an instance already exists, return it
	if instance != nil {
		return instance, nil
	}

	// Otherwise, create a new instance
	p, err := NewPersistence()
	if err != nil {
		return nil, err
	}

	instance = p
	return instance, nil
}

// LogRequest writes a request into the disk
func (p *Persistence) LogRequest(req map[string]interface{}) error {
	p.mu.Lock() // Protect file writes
	defer p.mu.Unlock()

	buf := new(bytes.Buffer)

	// Write command length and command
	cmd := req["command"].(string)
	if err := binary.Write(buf, binary.LittleEndian, int32(len(cmd))); err != nil {
		return err
	}
	buf.WriteString(cmd)

	// Write key length and key
	key := req["key"].(string)
	if err := binary.Write(buf, binary.LittleEndian, int32(len(key))); err != nil {
		return err
	}
	buf.WriteString(key)

	// Write value length and value (if present)
	val, valExists := req["value"].(string)
	if valExists {
		binary.Write(buf, binary.LittleEndian, int32(len(val)))
		buf.WriteString(val)
	} else {
		binary.Write(buf, binary.LittleEndian, int32(0)) // No value
	}

	// Write offset length and offset (if present)
	offset, offsetExists := req["offset"].(string)
	if offsetExists {
		binary.Write(buf, binary.LittleEndian, int32(len(offset)))
		buf.WriteString(offset)
	} else {
		binary.Write(buf, binary.LittleEndian, int32(0)) // No value
	}

	// Write End Marker (4 bytes "EOF\0")
	buf.Write([]byte{0x45, 0x4F, 0x46, 0x00})

	// Write to file
	_, err := p.file.Write(buf.Bytes())
	return err
}

// LoadRequests reads the binary log and returns parsed requests
func (p *Persistence) LoadRequests() ([]map[string]interface{}, error) {
	p.mu.Lock() // Protect file reads
	defer p.mu.Unlock()

	// Move file pointer to start
	if _, err := p.file.Seek(0, 0); err != nil {
		return nil, err
	}

	var requests []map[string]interface{}
	buf := make([]byte, 1024) // Buffer for reading

	for {
		// Read command length
		if _, err := p.file.Read(buf[:4]); err != nil {
			break
		}
		cmdLen := binary.LittleEndian.Uint32(buf[:4])

		// Read command
		cmdBuf := make([]byte, cmdLen)
		if _, err := p.file.Read(cmdBuf); err != nil {
			break
		}
		command := string(cmdBuf)

		// Read key length
		if _, err := p.file.Read(buf[:4]); err != nil {
			break
		}
		keyLen := binary.LittleEndian.Uint32(buf[:4])

		// Read key
		keyBuf := make([]byte, keyLen)
		if _, err := p.file.Read(keyBuf); err != nil {
			break
		}
		key := string(keyBuf)

		// Read value length
		if _, err := p.file.Read(buf[:4]); err != nil {
			break
		}
		valLen := binary.LittleEndian.Uint32(buf[:4])

		// Read value (if present)
		var value string
		if valLen > 0 {
			valBuf := make([]byte, valLen)
			if _, err := p.file.Read(valBuf); err != nil {
				break
			}
			value = string(valBuf)
		}

		// Read offset length
		if _, err := p.file.Read(buf[:4]); err != nil {
			break
		}
		offsetLen := binary.LittleEndian.Uint32(buf[:4])

		// Read offset (if present)
		var offset string
		if offsetLen > 0 {
			offsetBuf := make([]byte, offsetLen)
			if _, err := p.file.Read(offsetBuf); err != nil {
				break
			}
			offset = string(offsetBuf)
		}

		// Read end marker (4 bytes)
		endMarker := make([]byte, 4)
		if _, err := p.file.Read(endMarker); err != nil || string(endMarker) != "EOF\x00" {
			break
		}

		// Construct request map
		req := map[string]interface{}{
			"command": command,
			"key":     key,
		}
		if value != "" {
			req["value"] = value
		}
		if offset != "" {
			req["offset"] = offset
		}

		requests = append(requests, req)
	}

	if len(requests) == 0 {
		return nil, errors.New("no data found in binary log")
	}

	return requests, nil
}

// Clear removes all logged requests by truncating the binary log file.
func (p *Persistence) Clear() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.file != nil {
		/**
		Why do we need to close the file: windows acts weird if the file is not closed and we try to truncatw
		*/
		p.file.Close()

		// Truncate the file to zero length
		if err := os.Truncate(p.file.Name(), 0); err != nil {
			return fmt.Errorf("failed to truncate file: %w", err)
		}

		// Reopen the file in the same mode as before
		file, err := os.OpenFile(p.file.Name(), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			return fmt.Errorf("failed to reopen file: %w", err)
		}

		p.file = file
	}

	return nil
}

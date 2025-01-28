package persistence

import (
	"encoding/gob"
	"os"
	"path/filepath"
)

// Operation represents an event that occurs in the database.
type Operation struct {
	Command string
	Key     string      // Key involved in the operation.
	Value   string      // Value associated with the key.
	Lvalue  interface{} // List value item.
	TTL     int64       // Time-to-live in milliseconds.
	OffSet  int         // Offset for the INCR command.
}

// Persistence handles the append-only log for the database.
type Persistence struct {
	file *os.File
	enc  *gob.Encoder
	dec  *gob.Decoder
}

// NewPersistence creates a new persistence instance.
func NewPersistence(persistenceType string) (*Persistence, error) {
	// Register the Operation type to avoid "Duplicate Types Received" errors.
	gob.Register(Operation{})

	homeDir, _ := os.UserHomeDir()
	persistenceDir := filepath.Join(homeDir, ".geomys", "persistence.db")

	// Open the file in append mode, create it if it doesn't exist.
	file, err := os.OpenFile(persistenceDir, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	return &Persistence{
		file: file,
		enc:  gob.NewEncoder(file),
		dec:  gob.NewDecoder(file),
	}, nil
}

// LogOperation writes an operation to the append-only file.
func (p *Persistence) LogOperation(op Operation) error {
	return p.enc.Encode(op) // Encode operation as binary.
}

// Close closes the persistence file.
func (p *Persistence) Close() error {
	return p.file.Close()
}

// LoadOperations reads the append-only log and returns the operations.
func (p *Persistence) LoadOperations() ([]Operation, error) {
	// Register the Operation type before decoding.
	gob.Register(Operation{})

	file, err := os.Open(p.file.Name())
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var operations []Operation
	decoder := gob.NewDecoder(file)
	for {
		var op Operation
		err := decoder.Decode(&op)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, err
		}
		operations = append(operations, op)
	}
	return operations, nil
}

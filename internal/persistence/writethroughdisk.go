package persistence

import (
	"bufio"
	"os"
	"path/filepath"
)

// Persistence handles the append-only log for the database.
type Persistence struct {
	file *os.File
}

// NewPersistence creates a new persistence instance.
func NewPersistence(persistencetype string) (*Persistence, error) {
	homeDir, _ := os.UserHomeDir()

	persistenceDir := filepath.Join(homeDir, ".geomys", "persistence.db")
	// Open the file in append mode, create it if it doesn't exist.
	file, err := os.OpenFile(persistenceDir, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	return &Persistence{file: file}, nil
}

// LogOperation writes an operation to the append-only file.
func (p *Persistence) LogOperation(operation string) error {
	if _, err := p.file.WriteString(operation + "\n"); err != nil {
		return err
	}
	return nil
}

// Close closes the persistence file.
func (p *Persistence) Close() error {
	return p.file.Close()
}

// LoadOperations reads the append-only log and returns the operations.
func (p *Persistence) LoadOperations() ([]string, error) {
	file, err := os.Open(p.file.Name())
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var operations []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		operations = append(operations, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return operations, nil
}

package spec

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
)

// OperationFingerprint represents a fingerprint of an operation's signature
type OperationFingerprint struct {
	Path        string
	Method      string
	OperationID string
	Hash        string
}

// SpecFingerprint contains fingerprints for all operations in a spec
type SpecFingerprint struct {
	SpecPath     string
	SpecHash     string // Overall spec hash for quick comparison
	Operations   map[string]OperationFingerprint // Key: "METHOD /path"
	OperationIDs map[string]string               // Map operationID to operation key
}

// CreateSpecFingerprint creates a fingerprint for an entire OpenAPI spec
func CreateSpecFingerprint(specPath string, spec *OpenAPISpec) (*SpecFingerprint, error) {
	fingerprint := &SpecFingerprint{
		SpecPath:     specPath,
		Operations:   make(map[string]OperationFingerprint),
		OperationIDs: make(map[string]string),
	}

	// Create fingerprints for all operations
	operations := spec.GetOperations()
	for _, op := range operations {
		opKey := fmt.Sprintf("%s %s", op.Method, op.Path)
		opHash, err := hashOperation(op)
		if err != nil {
			return nil, fmt.Errorf("failed to hash operation %s: %w", opKey, err)
		}

		opFingerprint := OperationFingerprint{
			Path:        op.Path,
			Method:      op.Method,
			OperationID: op.OperationID,
			Hash:        opHash,
		}

		fingerprint.Operations[opKey] = opFingerprint

		// Map operationID to operation key for lookup
		if op.OperationID != "" {
			fingerprint.OperationIDs[op.OperationID] = opKey
		}
	}

	// Create overall spec hash from all operation hashes
	specHash, err := hashSpec(fingerprint.Operations)
	if err != nil {
		return nil, fmt.Errorf("failed to create spec hash: %w", err)
	}
	fingerprint.SpecHash = specHash

	return fingerprint, nil
}

// hashOperation creates a SHA256 hash of an operation's significant fields
func hashOperation(op OperationInfo) (string, error) {
	// Create a canonical representation of the operation
	canonical := map[string]interface{}{
		"path":   op.Path,
		"method": op.Method,
	}

	if op.Operation != nil {
		// Include operation details that affect generated code
		if op.Operation.OperationID != "" {
			canonical["operationId"] = op.Operation.OperationID
		}

		// Include parameters (affects function signature)
		if len(op.Operation.Parameters) > 0 {
			canonical["parameters"] = op.Operation.Parameters
		}

		// Include request body (affects function signature)
		if op.Operation.RequestBody != nil {
			canonical["requestBody"] = op.Operation.RequestBody
		}

		// Include responses (affects return types)
		if len(op.Operation.Responses) > 0 {
			canonical["responses"] = op.Operation.Responses
		}

		// Include tags (may affect generated code organization)
		if len(op.Operation.Tags) > 0 {
			canonical["tags"] = op.Operation.Tags
		}
	}

	// Convert to JSON for hashing
	data, err := json.Marshal(canonical)
	if err != nil {
		return "", fmt.Errorf("failed to marshal operation: %w", err)
	}

	// Create SHA256 hash
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash), nil
}

// hashSpec creates a hash from all operation hashes
func hashSpec(operations map[string]OperationFingerprint) (string, error) {
	// Get all operation keys and sort them for deterministic hashing
	keys := make([]string, 0, len(operations))
	for key := range operations {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Concatenate all operation hashes in sorted order
	var allHashes string
	for _, key := range keys {
		allHashes += operations[key].Hash
	}

	// Hash the concatenated hashes
	hash := sha256.Sum256([]byte(allHashes))
	return fmt.Sprintf("%x", hash), nil
}

// CompareFingerprints compares two spec fingerprints and returns changes
func CompareFingerprints(old, new *SpecFingerprint) *FingerprintComparison {
	comparison := &FingerprintComparison{
		Added:     make([]string, 0),
		Modified:  make([]string, 0),
		Deleted:   make([]string, 0),
		Unchanged: make([]string, 0),
	}

	// Quick check: if spec hashes match, nothing changed
	if old.SpecHash == new.SpecHash {
		for key := range new.Operations {
			comparison.Unchanged = append(comparison.Unchanged, key)
		}
		return comparison
	}

	// Find added and modified operations
	for key, newOp := range new.Operations {
		if oldOp, exists := old.Operations[key]; exists {
			if oldOp.Hash != newOp.Hash {
				comparison.Modified = append(comparison.Modified, key)
			} else {
				comparison.Unchanged = append(comparison.Unchanged, key)
			}
		} else {
			comparison.Added = append(comparison.Added, key)
		}
	}

	// Find deleted operations
	for key := range old.Operations {
		if _, exists := new.Operations[key]; !exists {
			comparison.Deleted = append(comparison.Deleted, key)
		}
	}

	// Sort for consistent output
	sort.Strings(comparison.Added)
	sort.Strings(comparison.Modified)
	sort.Strings(comparison.Deleted)
	sort.Strings(comparison.Unchanged)

	return comparison
}

// FingerprintComparison contains the results of comparing two fingerprints
type FingerprintComparison struct {
	Added     []string // Operations added in new spec
	Modified  []string // Operations modified in new spec
	Deleted   []string // Operations deleted in new spec
	Unchanged []string // Operations that didn't change
}

// HasChanges returns true if there are any changes
func (c *FingerprintComparison) HasChanges() bool {
	return len(c.Added) > 0 || len(c.Modified) > 0 || len(c.Deleted) > 0
}

// Summary returns a human-readable summary of changes
func (c *FingerprintComparison) Summary() string {
	if !c.HasChanges() {
		return "No changes detected"
	}

	return fmt.Sprintf("Changes: +%d added, ~%d modified, -%d deleted (%d unchanged)",
		len(c.Added), len(c.Modified), len(c.Deleted), len(c.Unchanged))
}

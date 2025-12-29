package remark

import (
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// Type represents the type of remark
type Type string

const (
	TypeThought  Type = "thought"
	TypeDoubt    Type = "doubt"
	TypeTodo     Type = "todo"
	TypeDecision Type = "decision"
)

// State represents the state of a remark
type State string

const (
	StateActive   State = "active"
	StateResolved State = "resolved"
)

// Remark represents a single note attached to a commit
type Remark struct {
	ID        string    `yaml:"id"`
	Type      Type      `yaml:"type"`
	Branch    string    `yaml:"branch"`
	State     State     `yaml:"state"`
	CreatedAt time.Time `yaml:"created_at"`
	Body      string    `yaml:"body"`
}

// Remarks is a container for multiple remarks on a single commit
type Remarks struct {
	Remarks []Remark `yaml:"remarks"`
}

// NewRemark creates a new remark with a generated ID
func NewRemark(remarkType Type, branch, body string) Remark {
	return Remark{
		ID:        generateShortUUID(),
		Type:      remarkType,
		Branch:    branch,
		State:     StateActive,
		CreatedAt: time.Now().UTC(),
		Body:      body,
	}
}

// generateShortUUID generates an 8-character UUID
func generateShortUUID() string {
	id := uuid.New()
	return id.String()[:8]
}

// ParseRemarks parses YAML content into Remarks
func ParseRemarks(data []byte) (*Remarks, error) {
	if len(data) == 0 {
		return &Remarks{}, nil
	}

	var remarks Remarks
	if err := yaml.Unmarshal(data, &remarks); err != nil {
		return nil, err
	}
	return &remarks, nil
}

// Marshal converts Remarks to YAML
func (r *Remarks) Marshal() ([]byte, error) {
	return yaml.Marshal(r)
}

// Add adds a new remark to the collection
func (r *Remarks) Add(remark Remark) {
	r.Remarks = append(r.Remarks, remark)
}

// Merge combines remarks from another Remarks collection
func (r *Remarks) Merge(other *Remarks) {
	if other == nil {
		return
	}
	r.Remarks = append(r.Remarks, other.Remarks...)
}

// FindByID finds a remark by its ID
func (r *Remarks) FindByID(id string) *Remark {
	for i := range r.Remarks {
		if r.Remarks[i].ID == id {
			return &r.Remarks[i]
		}
	}
	return nil
}

// RemoveByID removes a remark by its ID and returns true if found
func (r *Remarks) RemoveByID(id string) bool {
	for i := range r.Remarks {
		if r.Remarks[i].ID == id {
			r.Remarks = append(r.Remarks[:i], r.Remarks[i+1:]...)
			return true
		}
	}
	return false
}

// ActiveForBranch returns all active remarks for a given branch
func (r *Remarks) ActiveForBranch(branch string) []Remark {
	var result []Remark
	for _, remark := range r.Remarks {
		if remark.State == StateActive && remark.Branch == branch {
			result = append(result, remark)
		}
	}
	return result
}

// IsEmpty returns true if there are no remarks
func (r *Remarks) IsEmpty() bool {
	return len(r.Remarks) == 0
}

// ValidateType checks if a string is a valid remark type
func ValidateType(t string) bool {
	switch Type(t) {
	case TypeThought, TypeDoubt, TypeTodo, TypeDecision:
		return true
	default:
		return false
	}
}


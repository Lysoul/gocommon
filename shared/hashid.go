package shared

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/kelseyhightower/envconfig"
	"github.com/speps/go-hashids"
)

type HashIDConfig struct {
	Salt      string `envconfig:"HASHID_SALT" required:"true"`
	MinLength uint   `envconfig:"HASHID_MIN_LENGTH" required:"true"`
}

//nolint:gochecknoglobals // later
var hasher = sync.OnceValue(func() *HashID {
	c := HashIDConfig{}
	envconfig.MustProcess("", &c)
	data := hashids.NewData()
	data.Salt = c.Salt
	data.MinLength = int(c.MinLength)
	hasher, err := hashids.NewWithData(data)
	if err != nil {
		panic(err)
	}
	return &HashID{
		hasher: hasher,
	}
})

// DecodeHash decodes hash into an ID.
func DecodeHash(hash []byte) (ID, error) {
	return hasher().Decode(hash)
}

// EncodeID encodes ID into a random string (hash).
func EncodeID(id ID) ([]byte, error) {
	return hasher().Encode(id)
}

// ID represents a unique identifier.
// It means to replace the old int64 as unique ID.
// Using this type allows the int64 to be obfuscated
// into a random string using the Hashids algorithm.
// Read more about hashids in https://hashids.org/.
type ID int64

// MarshalJSON marshals the ID to JSON.
func (id ID) MarshalJSON() ([]byte, error) {
	if id == 0 {
		return json.Marshal(nil)
	}

	res, err := hasher().Encode(id)
	if err != nil {
		return nil, err
	}
	return json.Marshal(string(res))
}

// UnmarshalJSON unmarshals the JSON back to ID.
func (id *ID) UnmarshalJSON(hash []byte) error {
	if strings.TrimSpace(string(hash)) == "null" {
		*id = 0
		return nil
	}

	if len(hash) >= 2 {
		hash = hash[1 : len(hash)-1]
	}

	res, err := hasher().Decode(hash)
	if err != nil {
		return err
	}
	*id = res
	return nil
}

// EncodeString encodes ID to hashsids format and returns as a string.
// It ignores the error coming from encoding process.
// Thus, if there is any error during the process, it returns empty string.
func (id ID) EncodeString() string {
	res, err := hasher().Encode(id)
	if err != nil {
		return ""
	}
	return string(res)
}

// Hash defines the contract to encode and decode the ID.
type Hash interface {
	// Encode encodes the ID into a slice of byte.
	// The slice of byte generated is the result of Hashids algorithm.
	Encode(ID) ([]byte, error)
	// Decode decodes the slice of byte into an ID.
	Decode([]byte) (ID, error)
}

// HashID can be used to encode and decode hashids.
// It implements the Hash interface.
type HashID struct {
	hasher *hashids.HashID
}

// Encode encodes the ID into a slice of byte.
func (h *HashID) Encode(id ID) ([]byte, error) {
	res, err := h.hasher.EncodeInt64([]int64{int64(id)})
	if err != nil {
		return nil, err
	}
	return []byte(res), nil
}

// Decode decodes the slice of byte into an ID.
func (h *HashID) Decode(hash []byte) (ID, error) {
	if len(hash) == 0 {
		return 0, nil
	}

	res, err := h.hasher.DecodeInt64WithError(string(hash))
	if err != nil {
		return 0, err
	}
	if len(res) != 1 {
		return 0, fmt.Errorf("expected decoded value must be only 1 ID, turns out be %d ID(s)", len(res))
	}
	return ID(res[0]), nil
}

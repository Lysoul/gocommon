package shared_test

import (
	"testing"

	"github.com/Lysoul/gocommon/shared"
	"github.com/stretchr/testify/require"
)

func Test_HashID(t *testing.T) {
	t.Setenv("HASHID_SALT", "test")
	t.Setenv("HASHID_MIN_LENGTH", "10")

	t.Run("MarshallJSON", func(t *testing.T) {
		id := shared.ID(1)
		b, err := id.MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != `"3wedgpzLRq"` {
			t.Fatalf("expected bX, got %s", string(b))
		}
	})

	t.Run("UnmarshallJSON", func(t *testing.T) {
		b := []byte(`"3wedgpzLRq"`)
		var id shared.ID
		err := id.UnmarshalJSON(b)
		if err != nil {
			t.Fatal(err)
		}
		require.Equal(t, shared.ID(1), id)
	})
}

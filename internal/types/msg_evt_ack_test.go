package types

import (
	"testing"
)

func TestNewAckMessage(t *testing.T) {
	t.Run("should reject illegal serial numbers", func(t *testing.T) {
		illegal := []string{
			"",
			"315g3078",
			"3158307",
			"315830781",
		}

		for _, sn := range illegal {
			if _, err := NewAckMessage(sn); err == nil {
				t.Fatalf("expected error but was nil for sn: %s", sn)
			}
		}
	})

	t.Run("should format correctly", func(t *testing.T) {
		sn := "31583078"
		expectedHex := "68001068105031583078000000007816"

		result, err := NewAckMessage(sn)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result) != len(rawPollMessage) {
			t.Fatalf("unexpected message length: %d", len(result))
		}

		resultHex := string(result)
		if resultHex != expectedHex {
			t.Fatalf("unexpected message: %s", resultHex)
		}
	})
}

package luhn

import "testing"

func TestCheckNumber(t *testing.T) {
	cases := []struct {
		name   string
		number string
		want   bool
	}{
		{"valid from spec", "12345678903", true},
		{"valid classic", "79927398713", true},
		{"invalid checksum", "79927398710", false},
		{"invalid from spec neighbour", "12345678901", false},
		{"empty", "", false},
		{"non-digit", "1234abc", false},
		{"single zero", "0", true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := CheckNumber(c.number); got != c.want {
				t.Errorf("CheckNumber(%q) = %v, want %v", c.number, got, c.want)
			}
		})
	}
}

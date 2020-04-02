package libatnd

import "testing"

func TestIsValidMACAddress(t *testing.T) {
	tests := []struct {
		addr string
		ok   bool
	}{
		{"01:23:45:67:89:ab", true},
		{"45:67:89:ab:cd:ef", true},
		{"45:67:89:AB:CD:EF", true},
		{"01:23:67:89:ab", false},
		{"45:67:89:ab:cd0ef", false},
		{"", false},
	}

	for idx, test := range tests {
		ok := IsValidMACAddress(test.addr)
		if test.ok != ok {
			t.Errorf("[%d] expected %v, got %v", idx, test.ok, ok)
		}
	}
}

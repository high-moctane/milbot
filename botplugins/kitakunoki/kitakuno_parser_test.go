package kitakunoki

import (
	"testing"
)

func TestKitakunoList(t *testing.T) {
	_, err := kitakunoList()
	if err != nil {
		t.Fatal(err)
	}
}

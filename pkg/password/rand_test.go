package password

import (
	"fmt"
	"testing"
)

func Test_Rand(t *testing.T) {
	out, err := Random(12)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(out)
}

package oftenfun

import (
	"fmt"
	"testing"
)

func TestDeepCopy(t *testing.T) {
	a := map[string]interface{}{"a": 1}
	b := CopyJson(a)
	fmt.Println(b)
}

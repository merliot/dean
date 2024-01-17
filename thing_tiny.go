//go:build tinygo

package dean

import "fmt"

func ThingStore(t Thinger) {
	if t.TestFlag(ThingFlagMetal) {
		fmt.Printf("THINGSTORE - not implemented\r\n")
	}
}

func ThingRestore(t Thinger) {
	if t.TestFlag(ThingFlagMetal) {
		fmt.Printf("THINGRESTORE - not implemented\r\n")
	}
}

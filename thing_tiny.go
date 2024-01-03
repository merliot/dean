//go:build tinygo

package dean

import "fmt"

func ThingStore(t Thinger) {
	if t.TestFlag(ThingFlagMetal) {
		fmt.Println("THINGSTORE - not implemented")
	}
}

func ThingRestore(t Thinger) {
	if t.TestFlag(ThingFlagMetal) {
		fmt.Println("THINGRESTORE - not implemented")
	}
}

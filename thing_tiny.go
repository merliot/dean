//go:build tinygo

package dean

func ThingStore(t Thinger) {
	if t.TestFlag(ThingFlagMetal) {
		println("THINGSTORE - not implemented")
	}
}

func ThingRestore(t Thinger) {
	if t.TestFlag(ThingFlagMetal) {
		println("THINGRESTORE - not implemented")
	}
}

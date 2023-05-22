//go:build !tinygo

package dean

import (
	"encoding/json"
	"os"
)

func ThingStore(t Thinger) {
	if t.TestFlag(ThingFlagMetal) {
		println("THINGSTORE")
		storeName := t.Model() + "-" + t.Id()
		bytes, _ := json.Marshal(t)
		os.WriteFile(storeName, bytes, 0600)
	}
}

func ThingRestore(t Thinger) {
	if t.TestFlag(ThingFlagMetal) {
		println("THINGRESTORE")
		storeName := t.Model() + "-" + t.Id()
		bytes, err := os.ReadFile(storeName)
		if err == nil {
			json.Unmarshal(bytes, t)
		} else {
			ThingStore(t)
		}
	}
}

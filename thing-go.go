//go:build !tinygo

package dean

import (
	"encoding/json"
	"fmt"
	"os"
)

func ThingStore(t Thinger) {
	if t.TestFlag(ThingFlagMetal) {
		fmt.Printf("THINGSTORE\r\n")
		id, model, _ := t.Identity()
		storeName := model + "-" + id
		bytes, _ := json.Marshal(t)
		os.WriteFile(storeName, bytes, 0600)
	}
}

func ThingRestore(t Thinger) {
	if t.TestFlag(ThingFlagMetal) {
		fmt.Printf("THINGRESTORE\r\n")
		id, model, _ := t.Identity()
		storeName := model + "-" + id
		bytes, err := os.ReadFile(storeName)
		if err == nil {
			json.Unmarshal(bytes, t)
		} else {
			ThingStore(t)
		}
	}
}

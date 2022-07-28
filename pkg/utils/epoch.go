package utils

var SlotBase uint64 = 32

func GetEpochFromSlot(slot uint64) uint64 {
	ent := slot / SlotBase
	// rest := slot % uint64(SlotBase)
	// if rest > 0 {
	// 	ent += 1
	// }
	return ent
}

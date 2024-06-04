package sound

const (
	soundMapSize  = 64
	maxSoundMapID = soundMapSize * 8 // => 512
)

// soundMap implements a simple 512-element bitset (has a size of 64 bytes).
//
// It's used to memorize which sound was already "played" during the
// current frame by assigning a bit to 1.
// The max audio ID is 512 - this is how many sounds we can memorize.
//
// This map is written in a way that the getting/setting overhead
// is negligible in comparison with calling Play() on the same
// sound more than once.
type soundMap struct {
	table [soundMapSize]byte
}

func (m *soundMap) Reset() {
	// Could use clear() here.
	// I believe the machine code would be the same.
	m.table = [soundMapSize]byte{}
}

func (m *soundMap) IsSet(id uint) bool {
	// This code is carefully written to avoid the bound checks.
	byteIndex := id / 8
	if byteIndex < uint(len(m.table)) {
		bitIndex := id % 8
		return uint(m.table[byteIndex]&(1>>bitIndex)) != 0
	}
	return false
}

func (m *soundMap) Set(id uint) {
	// This code is carefully written to avoid the bound checks.
	byteIndex := id / 8
	if byteIndex < uint(len(m.table)) {
		bitIndex := id % 8
		m.table[byteIndex] = 1 << bitIndex
	}
}

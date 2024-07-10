package sound

import (
	"runtime"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
	resource "github.com/quasilyte/ebitengine-resource"
)

type System struct {
	loader *resource.Loader

	audioContext *audio.Context

	// This small bitset is used to track sounds with id<maxSoundMapID.
	// These sounds will be "played" only once during a frame.
	// Therefore, doing multiple PlaySound(id) during a single frame
	// is more efficient.
	soundMap soundMap

	groupVolume [8]float64

	globalVolume float64
}

type PlayOptions struct {
	Volume float64

	Position time.Duration
}

func (sys *System) Init(a *audio.Context, l *resource.Loader) {
	sys.loader = l
	sys.audioContext = a
	sys.globalVolume = 1.0

	for i := range sys.groupVolume {
		sys.groupVolume[i] = 1.0
	}

	if runtime.GOOS != "android" {
		// Audio player factory has lazy initialization that may lead
		// to a ~0.2s delay before the first sound can be played.
		// To avoid that delay, we force that factory to initialize
		// right now, before the game is started.
		dummy := sys.audioContext.NewPlayerFromBytes(nil)
		dummy.Rewind()
	}
}

func (sys *System) GetContext() *audio.Context {
	return sys.audioContext
}

// Update adjusts the audio system state for the next tick of the game.
//
// It needs to be called somewhere near the beginning of [ebiten.Game.Update] method.
func (sys *System) Update() {
	sys.soundMap.Reset()
}

// GetGlobalVolume reports the current global volume multiplier.
func (sys *System) GetGlobalVolume() float64 {
	return sys.globalVolume
}

// SetGlobalVolume assigns an extra volume multiplier.
// It's used when computing the effective volume level.
// Setting it to 0.5 would make all sounds two times quiter.
//
// Imagine an average game that has separate volume controls (sfx, voice, music)
// plus "master volume" that would be applied on top of that.
// The global volume level is that multiplier.
func (sys *System) SetGlobalVolume(volume float64) {
	sys.globalVolume = volume
}

// GetGroupVolume reports the current volume multiplier for the given group.
// The max groupID is 7 (therefore, there could be 8 groups in total).
// Use [SetGroupVolume] to adjust the group's volume multiplier.
func (sys *System) GetGroupVolume(groupID uint) float64 {
	if groupID >= uint(len(sys.groupVolume)) {
		panic("invalid group ID")
	}
	return sys.groupVolume[groupID]
}

// SetGroupVolume assigns the volume multiplier for the given group.
// The max groupID is 7 (therefore, there could be 8 groups in total).
// Use [GetGroupVolume] to get the group's current volume multiplier.
//
// A sound multiplier of 0 effectively mutes the group.
func (sys *System) SetGroupVolume(groupID uint, multiplier float64) {
	if groupID >= uint(len(sys.groupVolume)) {
		panic("invalid group ID")
	}
	sys.groupVolume[groupID] = multiplier
}

func (sys *System) PlaySoundWithOptions(id resource.AudioID, opts PlayOptions) resource.Audio {
	return sys.playSound(id, opts)
}

// PlaySound is a shorthand for [PlaySoundWithOptions](id, {Volume: 1.0}).
//
// The returned Audio object can be ignored unless you want to check it's
// player field for a readonly operation like IsPlaying.
func (sys *System) PlaySound(id resource.AudioID) resource.Audio {
	return sys.PlaySoundWithOptions(id, PlayOptions{
		Volume: 1,
	})
}

func (sys *System) playSound(id resource.AudioID, opts PlayOptions) resource.Audio {
	res := sys.loader.LoadWAV(id)

	if sys.soundMap.IsSet(uint(id)) {
		return res
	}
	sys.soundMap.Set(uint(id))

	finalVolume := sys.calculateVolume(res, opts.Volume)
	if finalVolume != 0 {
		res.Player.SetVolume(finalVolume)
		// Rewind() calls SetPosition(0) internally anyway,
		// so there is no gain in branching here between Rewind and SetPosition.
		res.Player.SetPosition(opts.Position)
		res.Player.Play()
	}
	return res
}

func (sys *System) calculateVolume(a resource.Audio, vol float64) float64 {
	return sys.globalVolume * sys.groupVolume[a.Group] * a.Volume * vol
}

package sound

import (
	resource "github.com/quasilyte/ebitengine-resource"
)

type Playlist struct {
	sys          *System
	current      resource.Audio
	currentIndex int
	list         []audioWithOptions
	silence      float64
	nextDelay    float64
	paused       bool

	// SelectFunc implements the next track selection.
	//
	// By default, this field is set to a function
	// that selects the next track and wraps to track 0.
	// Another example of a selection strategy would be
	// a random selection or ping-pong style wrapping.
	//
	// When the first track needs to be selected,
	// the currentIndex argument would be -1 as
	// no tracks are being played at that moment.
	//
	// Use [Playlist.Len] to determine the appropriate
	// track index bound.
	SelectFunc func(currentIndex int) int
}

func NewPlaylist(sys *System) *Playlist {
	pl := &Playlist{
		sys:          sys,
		list:         make([]audioWithOptions, 0, 4),
		currentIndex: -1,
		paused:       true,
	}
	pl.SelectFunc = func(currentIndex int) int {
		nextIndex := currentIndex + 1
		if nextIndex < pl.Len() {
			return nextIndex
		}
		return 0 // Wrap to the beginning of the playlist
	}
	return pl
}

func (pl *Playlist) AddWithOptions(id resource.AudioID, opts PlayOptions) {
	pl.list = append(pl.list, audioWithOptions{
		id:   id,
		opts: opts,
	})
}

func (pl *Playlist) Add(id resource.AudioID) {
	pl.AddWithOptions(id, PlayOptions{Volume: 1.0})
}

func (pl *Playlist) Len() int {
	return len(pl.list)
}

func (pl *Playlist) IsPaused() bool {
	return pl.paused
}

func (pl *Playlist) SetPaused(paused bool) {
	pl.paused = paused

	if pl.current.Player == nil {
		return
	}
	if paused {
		pl.current.Player.Pause()
	} else {
		vol := pl.sys.calculateVolume(pl.current, pl.list[pl.currentIndex].opts.Volume)
		if vol != 0 {
			pl.current.Player.SetVolume(vol)
			pl.current.Player.Play()
		}
	}
}

func (pl *Playlist) GetSilenceDuration() float64 {
	return pl.silence
}

func (pl *Playlist) SetSilenceDuration(seconds float64) {
	pl.silence = seconds
}

func (pl *Playlist) Update(delta float64) {
	if len(pl.list) == 0 {
		return
	}
	if pl.paused {
		return
	}

	if pl.currentIndex == -1 {
		pl.currentIndex = pl.SelectFunc(pl.currentIndex)
		// el := pl.list[pl.currentIndex]
		// pl.current = pl.sys.PlaySoundWithOptions(el.id, el.opts)
	}

	if pl.current.Player != nil && pl.current.Player.IsPlaying() {
		return
	}

	pl.nextDelay -= delta
	if pl.nextDelay > 0 {
		return
	}
	pl.nextDelay = pl.silence
	el := pl.list[pl.currentIndex]
	pl.current = pl.sys.PlaySoundWithOptions(el.id, el.opts)
}

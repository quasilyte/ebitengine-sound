package sound

import (
	resource "github.com/quasilyte/ebitengine-resource"
)

type Queue struct {
	sys     *System
	current resource.Audio
	queued  []audioWithOptions
}

func NewQueue(sys *System) *Queue {
	return &Queue{
		sys:    sys,
		queued: make([]audioWithOptions, 0, 4),
	}
}

// Reset stops the currently playing sounds and clears all queued sounds.
func (q *Queue) Reset() {
	if q.current.Player != nil {
		q.current.Player.Pause()
		q.current = resource.Audio{}
	}
	q.queued = q.queued[:0]
}

// PlaySoundWithOptions adds the sound to the queue.
// If the queue is currently empty, it will start playing right away.
// Otherwise, it will be played after all other queued sound have played.
func (q *Queue) PlaySoundWithOptions(id resource.AudioID, opts PlayOptions) {
	q.queued = append(q.queued, audioWithOptions{
		id:   id,
		opts: opts,
	})
}

// PlaySound is a shorthand for [PlaySoundWithOptions](id, {Volume: 1.0}).
func (q *Queue) PlaySound(id resource.AudioID) {
	q.PlaySoundWithOptions(id, PlayOptions{Volume: 1.0})
}

func (q *Queue) Update() {
	if q.current.Player == nil {
		if len(q.queued) == 0 {
			// Nothing to play in the queue.
			return
		}

		// Do a dequeue.
		el := q.queued[0]
		copy(q.queued, q.queued[1:])
		q.queued = q.queued[:len(q.queued)-1]

		q.current = q.sys.PlaySoundWithOptions(el.id, el.opts)
		return
	}

	if !q.current.Player.IsPlaying() {
		// Finished playing the current enqueued sound.
		q.current = resource.Audio{}
	}
}

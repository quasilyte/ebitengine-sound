package sound

import (
	resource "github.com/quasilyte/ebitengine-resource"
)

type audioWithOptions struct {
	id   resource.AudioID
	opts PlayOptions
}

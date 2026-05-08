package ai

import "time"

type AudioRequest interface {
	Model(model string) AudioRequest
	Provider(provider string) AudioRequest
	Voice(voice string) AudioRequest
	Male() AudioRequest
	Female() AudioRequest
	Instructions(instructions string) AudioRequest
	Timeout(timeout time.Duration) AudioRequest
	Store(disk ...string) (string, error)
	StoreAs(path string, disk ...string) (string, error)
	Generate() (AudioResponse, error)
}

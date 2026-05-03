package ai

import "time"

type ImageQuality string

const (
	ImageQualityLow    ImageQuality = "low"
	ImageQualityMedium ImageQuality = "medium"
	ImageQualityHigh   ImageQuality = "high"
)

type ImageSize string

const (
	ImageSizeSquare    ImageSize = "1024x1024"
	ImageSizePortrait  ImageSize = "1024x1536"
	ImageSizeLandscape ImageSize = "1536x1024"
)

type ImageRequest interface {
	Model(model string) ImageRequest
	Provider(provider string) ImageRequest
	Square() ImageRequest
	Portrait() ImageRequest
	Landscape() ImageRequest
	Quality(quality ImageQuality) ImageRequest
	Attachments(attachments ...Attachment) ImageRequest
	Timeout(timeout time.Duration) ImageRequest
	Generate() (ImageResponse, error)
}

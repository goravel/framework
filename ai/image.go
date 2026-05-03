package ai

import contractsai "github.com/goravel/framework/contracts/ai"

type ImageQuality = contractsai.ImageQuality
type ImageSize = contractsai.ImageSize

const (
	ImageQualityLow    = contractsai.ImageQualityLow
	ImageQualityMedium = contractsai.ImageQualityMedium
	ImageQualityHigh   = contractsai.ImageQualityHigh

	ImageSizeSquare    = contractsai.ImageSizeSquare
	ImageSizePortrait  = contractsai.ImageSizePortrait
	ImageSizeLandscape = contractsai.ImageSizeLandscape
)

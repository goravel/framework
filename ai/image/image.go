package image

import contractsai "github.com/goravel/framework/contracts/ai"

type Quality = contractsai.ImageQuality
type Size = contractsai.ImageSize

const (
	QualityLow    = contractsai.ImageQualityLow
	QualityMedium = contractsai.ImageQualityMedium
	QualityHigh   = contractsai.ImageQualityHigh

	SizeSquare    = contractsai.ImageSizeSquare
	SizePortrait  = contractsai.ImageSizePortrait
	SizeLandscape = contractsai.ImageSizeLandscape
)

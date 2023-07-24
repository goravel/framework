package gorm

import "github.com/mitchellh/mapstructure"

type CursorImpl struct {
	row map[string]any
}

func (c *CursorImpl) Scan(value any) error {
	msConfig := &mapstructure.DecoderConfig{
		Squash: true,
		Result: value,
	}

	decoder, err := mapstructure.NewDecoder(msConfig)
	if err != nil {
		return err
	}

	return decoder.Decode(c.row)
}

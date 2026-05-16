package core

type Options struct {
	Quantizer      Quantizer
	Strength       float32
	Threshold      uint8
	Seed           uint64
	RandomStrength uint8
}

const DefaultOrderedStrength = 64.0 / 255.0

func (o Options) WithDefaults() Options {
	if o.Strength == 0 {
		o.Strength = DefaultOrderedStrength
	}
	if o.Seed == 0 {
		o.Seed = 1
	}
	return o
}

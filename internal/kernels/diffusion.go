package kernels

type Tap struct {
	DX        int8
	DY        int8
	WeightNum int16
}

type ErrorKernel struct {
	Taps      []Tap
	WeightDen int16
}

var (
	FloydSteinberg      = ErrorKernel{Taps: []Tap{{1, 0, 7}, {-1, 1, 3}, {0, 1, 5}, {1, 1, 1}}, WeightDen: 16}
	FalseFloydSteinberg = ErrorKernel{Taps: []Tap{{1, 0, 3}, {0, 1, 3}, {1, 1, 2}}, WeightDen: 8}
	JarvisJudiceNinke   = ErrorKernel{Taps: []Tap{{1, 0, 7}, {2, 0, 5}, {-2, 1, 3}, {-1, 1, 5}, {0, 1, 7}, {1, 1, 5}, {2, 1, 3}, {-2, 2, 1}, {-1, 2, 3}, {0, 2, 5}, {1, 2, 3}, {2, 2, 1}}, WeightDen: 48}
	Stucki              = ErrorKernel{Taps: []Tap{{1, 0, 8}, {2, 0, 4}, {-2, 1, 2}, {-1, 1, 4}, {0, 1, 8}, {1, 1, 4}, {2, 1, 2}, {-2, 2, 1}, {-1, 2, 2}, {0, 2, 4}, {1, 2, 2}, {2, 2, 1}}, WeightDen: 42}
	Burkes              = ErrorKernel{Taps: []Tap{{1, 0, 8}, {2, 0, 4}, {-2, 1, 2}, {-1, 1, 4}, {0, 1, 8}, {1, 1, 4}, {2, 1, 2}}, WeightDen: 32}
	Sierra              = ErrorKernel{Taps: []Tap{{1, 0, 5}, {2, 0, 3}, {-2, 1, 2}, {-1, 1, 4}, {0, 1, 5}, {1, 1, 4}, {2, 1, 2}, {-1, 2, 2}, {0, 2, 3}, {1, 2, 2}}, WeightDen: 32}
	TwoRowSierra        = ErrorKernel{Taps: []Tap{{1, 0, 4}, {2, 0, 3}, {-2, 1, 1}, {-1, 1, 2}, {0, 1, 3}, {1, 1, 2}, {2, 1, 1}}, WeightDen: 16}
	SierraLite          = ErrorKernel{Taps: []Tap{{1, 0, 2}, {-1, 1, 1}, {0, 1, 1}}, WeightDen: 4}
	StevensonArce       = ErrorKernel{Taps: []Tap{{2, 0, 32}, {-3, 1, 12}, {-1, 1, 26}, {1, 1, 30}, {3, 1, 16}, {-2, 2, 12}, {0, 2, 26}, {2, 2, 12}, {-3, 3, 5}, {-1, 3, 12}, {1, 3, 12}, {3, 3, 5}}, WeightDen: 200}
	Atkinson            = ErrorKernel{Taps: []Tap{{1, 0, 1}, {2, 0, 1}, {-1, 1, 1}, {0, 1, 1}, {1, 1, 1}, {0, 2, 1}}, WeightDen: 8}
)

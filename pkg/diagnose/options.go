package diagnose

type Op struct {
	nvidiaSMIQueryCommand string
	ibstatCommand         string

	debug         bool
	createArchive bool

	pollGPMEvents bool

	netcheck  bool
	diskcheck bool

	dmesgCheck bool

	checkInfiniband bool
}

type OpOption func(*Op)

func (op *Op) applyOpts(opts []OpOption) error {
	for _, opt := range opts {
		opt(op)
	}

	if op.nvidiaSMIQueryCommand == "" {
		op.nvidiaSMIQueryCommand = "nvidia-smi --query"
	}
	if op.ibstatCommand == "" {
		op.ibstatCommand = "ibstat"
	}

	return nil
}

func WithNvidiaSMIQueryCommand(p string) OpOption {
	return func(op *Op) {
		op.nvidiaSMIQueryCommand = p
	}
}

// Specifies the ibstat binary path to overwrite the default path.
func WithIbstatCommand(p string) OpOption {
	return func(op *Op) {
		op.ibstatCommand = p
	}
}

func WithDebug(b bool) OpOption {
	return func(op *Op) {
		op.debug = b
	}
}

func WithCreateArchive(b bool) OpOption {
	return func(op *Op) {
		op.createArchive = b
	}
}

func WithPollGPMEvents(b bool) OpOption {
	return func(op *Op) {
		op.pollGPMEvents = b
	}
}

// WithNetcheck enables network connectivity checks to global edge/derp servers.
func WithNetcheck(b bool) OpOption {
	return func(op *Op) {
		op.netcheck = b
	}
}

func WithDiskcheck(b bool) OpOption {
	return func(op *Op) {
		op.diskcheck = b
	}
}

func WithDmesgCheck(b bool) OpOption {
	return func(op *Op) {
		op.dmesgCheck = b
	}
}

func WithCheckInfiniband(b bool) OpOption {
	return func(op *Op) {
		op.checkInfiniband = b
	}
}

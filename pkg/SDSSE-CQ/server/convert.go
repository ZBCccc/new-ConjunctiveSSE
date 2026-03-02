package server

import (
	pb "ConjunctiveSSE/pkg/SDSSE-CQ/proto"
)

// No conversion needed for SDSSE-CQ as it uses basic types
var _ = pb.SetupRequest{}
var _ = pb.SetupResponse{}

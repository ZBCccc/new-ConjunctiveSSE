package client

import (
	aura "ConjunctiveSSE/pkg/Aura/Core/SSEClient"
)
type SDSSECQClient struct {
	TSet          *aura.SSEClient
	XSet          *aura.SSEClient
	CT            map[string]int
	k, kx, ki, kz []byte
	iv            []byte
	// client pb.SDSSECQServiceClient
}
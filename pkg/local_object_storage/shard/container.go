package shard

import (
	"context"
	"fmt"

	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
)

type ContainerSizePrm struct {
	cnr cid.ID
}

type ContainerSizeRes struct {
	size uint64
}

func (p *ContainerSizePrm) SetContainerID(cnr cid.ID) {
	p.cnr = cnr
}

func (r ContainerSizeRes) Size() uint64 {
	return r.size
}

func (s *Shard) ContainerSize(prm ContainerSizePrm) (ContainerSizeRes, error) {
	s.m.RLock()
	defer s.m.RUnlock()

	if s.info.Mode.NoMetabase() {
		return ContainerSizeRes{}, ErrDegradedMode
	}

	size, err := s.metaBase.ContainerSize(prm.cnr)
	if err != nil {
		return ContainerSizeRes{}, fmt.Errorf("could not get container size: %w", err)
	}

	return ContainerSizeRes{
		size: size,
	}, nil
}

// DeleteContainer deletes any information related to the container
// including:
// - Metabase;
// - Blobstor;
// - Pilorama (if configured);
// - Write-cache (if configured).
func (s *Shard) DeleteContainer(_ context.Context, cID cid.ID) error {
	s.m.RLock()
	defer s.m.RUnlock()

	m := s.info.Mode
	if m.ReadOnly() {
		return ErrReadOnlyMode
	}

	inhumedAvailable, err := s.metaBase.InhumeContainer(cID)
	if err != nil {
		return fmt.Errorf("inhuming container in metabase: %w", err)
	}

	s.decObjectCounterBy(logical, inhumedAvailable)

	return nil
}

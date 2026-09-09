//go:build !linux

package processes

import (
	"errors"
	"time"

	"github.com/melvynx/code-os/internal/model"
)

type Scanner struct {
	ProcRoot string
}

func (Scanner) Scan() ([]model.AgentProcess, error) {
	return nil, nil
}

func (Scanner) Terminate(string) error {
	return errors.New("agent process controls currently require Linux")
}

func (Scanner) EnrichApplications(_ []model.Application, previous map[int]CPUMark, _ time.Time) map[int]CPUMark {
	return previous
}

func (Scanner) EnrichApplications(applications []model.Application, previous map[int]CPUMark, _ time.Time) map[int]CPUMark {
	return previous
}

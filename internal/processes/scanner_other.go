//go:build !linux

package processes

import (
	"errors"

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

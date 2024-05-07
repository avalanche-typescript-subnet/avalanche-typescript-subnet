// Copyright (C) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"context"
	"time"

	"github.com/akamensky/argparse"
	"github.com/ava-labs/avalanchego/utils/logging"
)

var _ Cmd = (*InterpreterCmd)(nil)

type InterpreterCmd struct {
	cmd *argparse.Command
}

func (s *InterpreterCmd) New(parser *argparse.Parser) {
	s.cmd = parser.NewCommand("interpreter", "Read input from a buffered stdin")
}

func (s *InterpreterCmd) Run(ctx context.Context, log logging.Logger, args []string) (*Response, error) {
	resp := newResponse(0)
	resp.setTimestamp(time.Now().Unix())
	return resp, nil
}

func (s *InterpreterCmd) Happened() bool {
	return s.cmd.Happened()
}

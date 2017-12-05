// Copyright (C) 2017 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-nebulas library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see <http://www.gnu.org/licenses/>.
//

package core

import (
	"encoding/json"

	"github.com/nebulasio/go-nebulas/nf/nvm"
	"github.com/nebulasio/go-nebulas/util"
)

// DeployPayload carry contract deploy information
type DeployPayload struct {
	Source string
	Args   string
}

// LoadDeployPayload from bytes
func LoadDeployPayload(bytes []byte) (*DeployPayload, error) {
	payload := &DeployPayload{}
	if err := json.Unmarshal(bytes, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

// NewDeployPayload with source & args
func NewDeployPayload(source, args string) *DeployPayload {
	return &DeployPayload{
		Source: source,
		Args:   args,
	}
}

// ToBytes serialize payload
func (payload *DeployPayload) ToBytes() ([]byte, error) {
	return json.Marshal(payload)
}

// Execute deploy payload in tx, deploy a new contract
func (payload *DeployPayload) Execute(tx *Transaction, block *Block) error {
	ctx, err := generateDeployContext(tx, block)
	if err != nil {
		return err
	}
	engine := nvm.NewV8Engine(ctx)
	//add gas limit and memory use limit
	engine.SetExecutionLimits(tx.GasLimit().Uint64(), nvm.DefaultLimitsOfTotalMemorySize)
	defer engine.Dispose()

	return engine.DeployAndInit(payload.Source, payload.Args)
}

// EstimateGas the payload in tx
func (payload *DeployPayload) EstimateGas(tx *Transaction, block *Block) (*util.Uint128, error) {
	ctx, err := generateDeployContext(tx, block)
	if err != nil {
		return nil, err
	}
	engine := nvm.NewV8Engine(ctx)
	//add gas limit and memory use limit
	engine.SetExecutionLimits(TransactionMaxGas.Uint64(), nvm.DefaultLimitsOfTotalMemorySize)
	defer engine.Dispose()
	err = engine.SimulationRun(payload.Source, "init", payload.Args)
	if err != nil {
		return nil, err
	}
	return util.NewUint128FromInt(int64(engine.ExecutionInstructions())), nil
}

func generateDeployContext(tx *Transaction, block *Block) (*nvm.Context, error) {
	addr, err := tx.GenerateContractAddress()
	if err != nil {
		return nil, err
	}
	context := block.accState
	owner := context.GetOrCreateUserAccount(tx.from.Bytes())
	contract, err := context.CreateContractAccount(addr.Bytes(), tx.Hash())
	if err != nil {
		return nil, err
	}
	ctx := nvm.NewContext(block, convertNvmTx(tx), owner, contract, context)
	return ctx, nil
}

func convertNvmTx(tx *Transaction) *nvm.ContextTransaction {
	ctxTx := &nvm.ContextTransaction{
		From:      tx.from.String(),
		To:        tx.to.String(),
		Value:     tx.value.String(),
		Timestamp: tx.timestamp,
		Nonce:     tx.Nonce(),
		Hash:      tx.Hash().String(),
		GasPrice:  tx.GasPrice().String(),
		GasLimit:  tx.GasLimit().String(),
	}
	return ctxTx
}

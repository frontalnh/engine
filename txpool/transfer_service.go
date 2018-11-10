/*
 * Copyright 2018 It-chain
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package txpool

import (
	"sync"

	"github.com/it-chain/engine/common"
	"github.com/it-chain/iLogger"
)

type Publisher func(topic string, data interface{}) (err error) //해당 publish함수는 midgard 에서 의존성 주입을 받기 위해 interface로 작성한다.
//모든 의존성 주입은 컴포넌트.go 에서 이루어짐

type TransferService struct {
	txpoolRepository TransactionRepository
	leaderRepository LeaderRepository
	eventService     EventService
	sync.RWMutex
}

func NewTransferService(txpoolRepository TransactionRepository, leaderRepository LeaderRepository, eventService EventService) *TransferService {
	return &TransferService{
		txpoolRepository: txpoolRepository,
		leaderRepository: leaderRepository,
		eventService:     eventService,
		RWMutex:          sync.RWMutex{},
	}
}

// This method is used to send transactions to the leader
func (ts TransferService) SendTransactionsToLeader() error {

	ts.Lock()
	defer ts.Unlock()

	transactions, err := ts.txpoolRepository.FindAll()
	if err != nil {
		return err
	}

	if len(transactions) == 0 {
		return nil
	}

	leader := ts.leaderRepository.Get()

	deliverCommand, err := common.CreateGrpcDeliverCommand(SendTransactionsToLeader, transactions)
	if err != nil {
		return err
	}

	deliverCommand.RecipientList = append(deliverCommand.RecipientList, leader.Id)

	err = ts.eventService.Publish("message.deliver", deliverCommand)
	if err != nil {
		return err
	}

	ts.clearTransactions(transactions)

	iLogger.Info(nil, "[Txpool] Transaction has been sent to the leader")

	return nil
}

func (ts TransferService) clearTransactions(transactions []Transaction) {
	for _, tx := range transactions {
		ts.txpoolRepository.Remove(tx.ID)
	}
}

package domain

import (
	"time"
	pb "it-chain/network/protos"
	"sync"
	"sync/atomic"
)

type MsgType int

const (
	PreprepareMsg  MsgType = iota
	PrepareMsg
	CommitMsg
)

//consesnsus message can has 3 types
type ConsensusMessage struct {
	ConsensusID string
	ViewID      string
	SequenceID  int64
	Block       *Block
	PeerID      string
	MsgType     MsgType
	TimeStamp   time.Time
}

type Stage int

const (
	PrePrepared   Stage = iota           // The ReqMsgs is processed successfully. The node is ready to head to the Prepare stage.
	Prepared                			 // Same with `prepared` stage explained in the original paper.
	Committed                			 // Same with `committed-local` stage explained in the original paper.
)

type EndConsensusHandle func(ConsensusState)

//동시에 여러개의 consensus가 진행될 수 있다.
//한개의 consensus는 1개의 state를 갖는다.
type ConsensusState struct {
	ID                  string
	ViewID              string
	CurrentStage        Stage
	Block               *Block
	PrepareMsgs         map[string]ConsensusMessage
	CommitMsgs          map[string]ConsensusMessage
	endChannel          chan struct{}
	endConsensusHandler EndConsensusHandle
	IsEnd               int32
	period 				int32
	sync.RWMutex
}

type View struct{
	ID string
}

//tested
func NewConsensusState(viewID string, consensusID string, block *Block, currentStage Stage, endConsensusHandler EndConsensusHandle, periodSeconds int32) *ConsensusState{

	cs := &ConsensusState{
		ID:consensusID,
		ViewID:viewID,
		CurrentStage:currentStage,
		Block: block,
		PrepareMsgs: make(map[string]ConsensusMessage),
		CommitMsgs: make(map[string]ConsensusMessage),
		endConsensusHandler: endConsensusHandler,
		IsEnd: int32(0),
		period: periodSeconds,
	}

	go cs.start()

	return cs
}

//tested
func NewConsesnsusMessage(consensusID string, viewID string,sequenceID int64, block *Block,peerID string, msgType MsgType) ConsensusMessage{

	return ConsensusMessage{
		ConsensusID: consensusID,
		ViewID: viewID,
		SequenceID: sequenceID,
		MsgType:msgType,
		TimeStamp: time.Now(),
		PeerID:peerID,
		Block: block,
	}
}

//todo block을 넣어야함
func FromConsensusProtoMessage(consensusMessage pb.ConsensusMessage) ConsensusMessage{

	return ConsensusMessage{
		ViewID: consensusMessage.ViewID,
		SequenceID: consensusMessage.SequenceID,
		PeerID: consensusMessage.PeerID,
		ConsensusID: consensusMessage.ConsensusID,
		MsgType: MsgType(consensusMessage.MsgType),
	}
}

//timer의 time이 끝나면 consensus를 종료한다.
//Consensus timer는 new 했을 때 시작된다.
func (cs *ConsensusState) start(){
	time.Sleep(time.Duration(cs.period)*time.Second)
	cs.Lock()
	defer cs.Unlock()

	//time out
	//consensus did not end
	//need to delete
	if cs.IsEnd == 0{
		cs.endConsensusHandler(*cs)
	}
}

func (cs *ConsensusState) End(){
	atomic.StoreInt32(&(cs.IsEnd), int32(1))
}

//message 종류에 따라서 다르게 넣어야함
func (cs *ConsensusState) AddMessage(consensusMessage ConsensusMessage){
	//PreprepareMsg는 block이 존재
	//commit, prepareMsg는 block 존재 안함
	//prepare가 2/3이상일 경우
	//commit이 2/3이상일 경우

	msgType := consensusMessage.MsgType

	switch msgType {
	case PreprepareMsg:
		cs.Block = consensusMessage.Block
		//prepareMsg broadcast 해야함
		break

	case PrepareMsg:
		//cs.PrepareMsgs = append(cs.PrepareMsgs, consensusMessage)
		//commitMsg broadcast 해야함
		break

	case CommitMsg:
		//cs.CommitMsgs = append(cs.CommitMsgs, consensusMessage)
		//block 저장해야함
		break
	default:
		break
	}
}

type Command interface{
	Execute()
}

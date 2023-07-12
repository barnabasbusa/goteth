package analyzer

import (
	"time"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/cortze/eth-cl-state-analyzer/pkg/spec"
)

func (s *ChainAnalyzer) DownloadNewState(
	queue *StateQueue,
	slot phase0.Slot,
	finalized bool) {

	log := log.WithField("routine", "download")
	ticker := time.NewTicker(minStateReqTime)

	log.Infof("requesting Beacon State from endpoint: epoch %d", slot/spec.SlotsPerEpoch)
	nextBstate, err := s.cli.RequestBeaconState(slot)
	if err != nil {
		// close the channel (to tell other routines to stop processing and end)
		log.Panicf("unable to retrieve beacon state from the beacon node, closing requester routine. %s", err.Error())
	}

	queue.AddNewState(*nextBstate)

	if queue.Complete() {
		// only execute tasks if prevBState is something (we have state and nextState in this case)

		epochTask := &EpochTask{
			NextState: queue.nextState,
			State:     queue.currentState,
			PrevState: queue.prevState,
			Finalized: finalized,
		}

		log.Debugf("sending task for slot: %d", epochTask.State.Slot)
		s.epochTaskChan <- epochTask
	}
	// check if the min Request time has been completed (to avoid spaming the API)
	<-ticker.C
}

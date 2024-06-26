package dtmsvr

import (
	"fmt"

	"github.com/dtm-labs/dtm/client/dtmcli"
	"github.com/dtm-labs/dtm/client/dtmcli/dtmimp"
	"github.com/dtm-labs/logger"
)

type transTccProcessor struct {
	*TransGlobal
}

func init() {
	registerProcessorCreator("tcc", func(trans *TransGlobal) transProcessor {
		return &transTccProcessor{TransGlobal: trans}
	})
}

func (t *transTccProcessor) GenBranches() []TransBranch {
	return []TransBranch{}
}

func (t *transTccProcessor) ProcessOnce(branches []TransBranch) error {
	if !t.needProcess() {
		return nil
	}

	// prepared 状态 并且 超时
	if t.Status == dtmcli.StatusPrepared && t.isTimeout() {
		// 修改状态为失败
		t.changeStatus(dtmcli.StatusAborting, withRollbackReason(fmt.Sprintf("Timeout after %d seconds", t.TimeoutToFail)))
	}

	op := dtmimp.If(t.Status == dtmcli.StatusSubmitted, dtmimp.OpConfirm, dtmimp.OpCancel).(string)

	for current := len(branches) - 1; current >= 0; current-- {
		if branches[current].Op == op && branches[current].Status == dtmcli.StatusPrepared {
			logger.Debugf("branch info: current: %d ID: %d", current, branches[current].ID)
			err := t.execBranch(&branches[current], current)
			if err != nil {
				return err
			}
		}
	}

	t.changeStatus(dtmimp.If(t.Status == dtmcli.StatusSubmitted, dtmcli.StatusSucceed, dtmcli.StatusFailed).(string))
	return nil
}

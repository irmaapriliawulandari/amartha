package mq

import (
	"amartha-test/entity"
	"encoding/json"

	"github.com/nsqio/go-nsq"
	log "github.com/sirupsen/logrus"
)

type billingUsecase interface {
	InitBilling(loan entity.DisburseLoanMsg) error
}

type disburseLoanHandler struct {
	uc billingUsecase
}

func NewDisburseLoanHandler(uc billingUsecase) *disburseLoanHandler {
	return &disburseLoanHandler{
		uc: uc,
	}
}

func (h *disburseLoanHandler) HandleMessage(m *nsq.Message) error {
	var msg entity.DisburseLoanMsg
	err := json.Unmarshal(m.Body, &msg)
	if err != nil {
		log.Errorf("[disburseLoanHandler][HandleMessage] unmarshal disburse loan message: %v", err)
		return nil
	}

	if !msg.Validate() {
		log.Errorf("[disburseLoanHandler][HandleMessage] invalid loan message")
		return nil
	}

	err = h.uc.InitBilling(msg)
	if err != nil {
		if m.Attempts == 5 {
			log.Errorf("[disburseLoanHandler][HandleMessage] failed process and reach limit: %v", err)
			return nil
		}

		log.Errorf("[disburseLoanHandler][HandleMessage] failed process: %v", err)
		return err
	}

	log.Infof("[disburseLoanHandler][HandleMessage] consumed: %s", string(m.Body))

	return nil
}

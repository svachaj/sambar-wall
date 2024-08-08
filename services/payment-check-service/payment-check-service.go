package paymentcheckservice

import (
	"errors"

	"github.com/emersion/go-imap/client"
	"github.com/rs/zerolog/log"
	"github.com/svachaj/sambar-wall/utils"

	"github.com/robfig/cron/v3"
)

type IPaymentCheckService interface {
	StartCheckingPayments() error
}

type PaymentService struct {
	emailService  utils.IEmailService
	imapClient    *client.Client
	cronScheduler *cron.Cron
}

func NewPaymentService(emailService utils.IEmailService, imapAddress, imapUsername, imapPassword string) IPaymentCheckService {

	cs := cron.New()
	if cs == nil {
		log.Err(errors.New("fail to init cron scheduler")).Msg("Fail to init cron scheduler")
		return nil
	}

	ic, err := client.DialTLS(imapAddress, nil)
	if err != nil {
		log.Err(err).Msg("Fail to dial IMAP server")
		return nil
	}

	err = ic.Login(imapUsername, imapPassword)
	if err != nil {
		log.Err(err).Msg("Fail to login to the IMAP server")
	}

	return &PaymentService{emailService: emailService, cronScheduler: cs, imapClient: ic}
}

func (svc *PaymentService) StartCheckingPayments() error {

	svc.cronScheduler.AddFunc("@every 2m", func() {
		log.Info().Msg("Check payments...")
	})

	svc.cronScheduler.Start()

	return nil
}

package paymentcheckservice

import (
	"errors"
	"fmt"
	"io"
	"mime/quotedprintable"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"github.com/svachaj/sambar-wall/db/types"
	"github.com/svachaj/sambar-wall/utils"

	"github.com/robfig/cron/v3"
)

type Payment struct {
	Amount          float64
	ReferenceNumber string
	Timestamp       int64
}

// IPaymentCheckService interface for payment checking service.
type IPaymentCheckService interface {
	StartCheckingPayments() error
}

// PaymentService handles automated payment checking from IMAP emails.
type PaymentService struct {
	db            *sqlx.DB
	emailService  utils.IEmailService
	imapAddress   string
	imapUsername  string
	imapPassword  string
	cronScheduler *cron.Cron
}

// NewPaymentService creates a new payment service instance.
func NewPaymentService(db *sqlx.DB, emailService utils.IEmailService, imapAddress, imapUsername, imapPassword string) IPaymentCheckService {

	cs := cron.New()
	if cs == nil {
		log.Err(errors.New("fail to init cron scheduler")).Msg("Fail to init cron scheduler")
		return nil
	}

	return &PaymentService{
		emailService:  emailService,
		cronScheduler: cs,
		db:            db,
		imapAddress:   imapAddress,
		imapUsername:  imapUsername,
		imapPassword:  imapPassword,
	}
}

// createIMAPClient creates and authenticates a new IMAP client connection.
func (svc *PaymentService) createIMAPClient() (*client.Client, error) {
	ic, err := client.DialTLS(svc.imapAddress, nil)
	if err != nil {
		return nil, fmt.Errorf("fail to dial IMAP server: %w", err)
	}

	err = ic.Login(svc.imapUsername, svc.imapPassword)
	if err != nil {
		ic.Close()
		return nil, fmt.Errorf("fail to login to IMAP server: %w", err)
	}

	return ic, nil
}

// StartCheckingPayments starts the cron scheduler for checking payments.
func (svc *PaymentService) StartCheckingPayments() error {

	svc.cronScheduler.AddFunc("@every 2m", func() {
		log.Info().Msg("Check payments started")

		// Create new IMAP client for each run
		imapClient, err := svc.createIMAPClient()
		if err != nil {
			log.Err(err).Msg("Failed to create IMAP client")
			return
		}
		defer imapClient.Close()

		// Check for new emails
		mbox, err := imapClient.Select("INBOX", false)

		if err != nil {
			log.Err(err).Msg("Fail to select INBOX")
			return
		}

		if mbox.Messages == 0 {
			log.Info().Msg("No new messages")
			return
		}

		seqset := new(imap.SeqSet)
		seqset.AddRange(1, mbox.Messages)

		msgsToDelete := new(imap.SeqSet)
		paymentsToProcess := make([]Payment, 0)

		// Get the whole message body
		section := &imap.BodySectionName{}
		items := []imap.FetchItem{section.FetchItem()}

		messages := make(chan *imap.Message, mbox.Messages)
		done := make(chan error, 1)
		go func() {
			done <- imapClient.Fetch(seqset, items, messages)
		}()

		for msg := range messages {

			r := msg.GetBody(section)
			if r == nil {
				log.Info().Msg("Server didn't return message body")
				msgsToDelete.AddNum(msg.SeqNum)
				continue
			}

			m, err := mail.ReadMessage(r)
			if err != nil {
				log.Err(err).Msgf("Failed to read message:  %v", err)
				continue
			}

			header := m.Header

			from := header.Get("From")
			date := header.Get("Date")

			if strings.Contains(from, "notification@csob.cz") {

				body, err := io.ReadAll(m.Body)
				if err != nil {
					log.Err(err).Msgf("Failed to read body:  %v", err)
					continue
				}
				encodedString := string(body)
				bodyString, err := decodeQuotedPrintable(encodedString)
				if err != nil {
					log.Err(err).Msgf("Failed to decode body:  %v", err)
					continue
				}

				//go line by line and find the payment
				lines := strings.Split(bodyString, "\n")
				payment := Payment{}
				for _, line := range lines {
					if strings.Contains(line, "Částka:") {
						amount := stringBetween(line, "+", " CZK")
						amount = strings.ReplaceAll(amount, ",", ".")
						amount = strings.ReplaceAll(amount, "\u00a0", "")
						amount = strings.TrimSpace(amount)

						//parse string to float
						amnt, err := strconv.ParseFloat(amount, 64)
						if err != nil {
							log.Err(err).Msgf("Failed to parse amount:  %v", err)
							continue
						}
						payment.Amount = amnt
					}
					if strings.Contains(line, "Variabilní symbol:") {
						referenceNumber := strings.ReplaceAll(line, "Variabilní symbol:", "")
						referenceNumber = strings.TrimSpace(referenceNumber)
						payment.ReferenceNumber = referenceNumber
					}

					if len(payment.ReferenceNumber) > 0 && payment.Amount > 0 {
						ts, err := time.Parse("Mon, 2 Jan 2006 15:04:05 -0700 (MST)", date)
						if err != nil {
							log.Err(err).Msgf("Failed to parse date:  %v", err)
							continue
						}
						payment.Timestamp = ts.UnixMilli()
						paymentsToProcess = append(paymentsToProcess, payment)
						payment = Payment{}
						break
					}
				}

				msgsToDelete.AddNum(msg.SeqNum)
			} else {
				msgsToDelete.AddNum(msg.SeqNum)
			}
		}

		if err := <-done; err != nil {
			log.Err(err).Msgf("Error: %v", err)
			return
		}

		// Process payments
		log.Info().Msg("Payments to process count: " + strconv.Itoa(len(paymentsToProcess)))
		for _, payment := range paymentsToProcess {

			// convert reference number to int
			var refNumber int
			refNumber, err = strconv.Atoi(payment.ReferenceNumber)
			if err != nil {
				log.Err(err).Msgf("Failed to convert reference number to int:  %v", err)
				continue
			}

			// check if application exists
			var countApp int = 0
			err = svc.db.Get(&countApp, "SELECT COUNT(id) FROM t_course_application_form WHERE ID = @p1", refNumber)
			if err != nil {
				log.Err(err).Msgf("Failed to get countApp:  %v", err)
				continue
			}

			if countApp > 0 {

				//check if payment exists by reference number + timestamp
				var count int
				err = svc.db.Get(&count, "SELECT COUNT(id) FROM t_payment WHERE reference_number = @p1 AND timestamp = @p2", payment.ReferenceNumber, payment.Timestamp)
				if err != nil {
					log.Err(err).Msgf("Failed to get count:  %v", err)
					continue
				}

				if count == 0 {
					//insert payment
					_, err := svc.db.Exec("INSERT INTO t_payment (reference_number, amount, timestamp) VALUES (@p1, @p2, @p3)", payment.ReferenceNumber, payment.Amount, payment.Timestamp)
					if err != nil {
						log.Err(err).Msgf("Failed to insert payment:  %v", err)
						continue
					}
				}

				// get sum of amount of payments with the same reference number and compare with the price of the course

				var sum float64
				err = svc.db.Get(&sum, "SELECT SUM(amount) FROM t_payment WHERE reference_number = @p1", payment.ReferenceNumber)
				if err != nil {
					log.Err(err).Msgf("Failed to get sum:  %v", err)
					continue
				}

				// get price of the course
				var price float64

				// select price of the course by reference number (application id)
				err = svc.db.Get(&price, `
			SELECT tc.price
			FROM t_course_application_form tcaf
			LEFT JOIN t_course tc on tc.ID = tcaf.ID_course
			WHERE tcaf.ID = @p1
			`, refNumber)
				if err != nil {
					log.Err(err).Msgf("Failed to get price:  %v", err)
					continue
				}

				if sum >= price {
					// update course status to paid
					_, err := svc.db.Exec("UPDATE t_course_application_form SET Paid = 1 WHERE ID = @p1", refNumber)
					if err != nil {
						log.Err(err).Msgf("Failed to update course status:  %v", err)
						continue
					}

					// send email to user only if the paymnet_email_sent is false
					var paymnetEmailSent bool
					err = svc.db.Get(&paymnetEmailSent, "SELECT payment_email_sent FROM t_course_application_form WHERE ID = @p1", refNumber)
					if err != nil {
						log.Err(err).Msgf("Failed to get payment_email_sent:  %v", err)
						continue
					}

					if !paymnetEmailSent {
						// get application form detail
						applicationForm := types.ApplicationForm{}

						err = svc.db.Get(&applicationForm, `
	SELECT 
tcaf.ID as id, 
tcaf.Paid as paid,
tcaf.PersonalId as personalId,
tc.ID as courseId,
tct.Name1 as courseName,
tct.Code as courseCode,
tcd.Name1 as courseDays,
tc.TimeFrom as courseTimeFrom,
tc.TimeTo as courseTimeTo,
tcag.Name1 as courseAgeGroup,
tc.Price as coursePrice,
tsup.FirstName as firstName,
tsup.LastName as lastName,
tcaf.Email as email
FROM t_course_application_form tcaf
LEFT JOIN t_course tc on tc.ID = tcaf.ID_course
LEFT join t_course_type tct on tct.ID = tc.ID_typeOfCourse
LEFT JOIN t_system_user_participant tsup on tcaf.ID_participant = tsup.ID
LEFT JOIN t_system_user tsu on tsu.ID = tsup.ID_ParentUser
LEFT JOIN t_course_day tcd on tc.ID_dayOfCourse = tcd.ID 
LEFT JOIN t_course_age_group tcag on tc.ID_ageGroup = tcag.ID 
WHERE tcaf.ID = @p1;
	`, refNumber)

						if err != nil {
							log.Err(err).Msg("Fail to send payment confirmation email due to get application form detail. Application form ID:" + payment.ReferenceNumber)
							continue
						}

						subject := "Potvrzení o zaplacení kurzu"
						body := "<div style=\"width: 100%; max-width: 600px;line-heigth:1.5rem; margin: 0 auto; padding: 20px; border: 1px solid #ccc; border-radius: 10px;\">\n"
						body += "<p style=\"font-size: 20px; margin-bottom: 20px;\">Dobrý den,</p>\n\n"
						body += "<p style=\"margin-bottom: 20px;\">Potvrzujeme úspěšné přijetí platby za kurz.</p>\n\n"

						body += "<strong>Přihláška číslo:</strong> " + payment.ReferenceNumber + "<br>\n"
						body += "<strong>Jméno úšastníka:</strong> " + applicationForm.FirstName + " " + applicationForm.LastName + "<br>\n"
						body += "<strong>Název kurzu:</strong> " + applicationForm.CourseName + "<br>\n"
						body += "<strong>Termín kurzu:</strong> " + applicationForm.CourseDays + " (" + applicationForm.CourseTimeFrom.Format("15:04") + " - " + applicationForm.CourseTimeTo.Format("15:04") + ")" + "<br>\n"

						body += "<br><br>\n\n"

						body += "<p style=\"font-size: 14px; color: #555;\">S pozdravem,<br>\n"
						body += "Lezecká Stěna Kladno</p>\n"
						body += "</div>\n"

						err = svc.emailService.SendEmail(subject, body, *applicationForm.Email)
						if err != nil {
							log.Err(err).Msg(fmt.Sprintf("Fail to send confirmation email of the payment. Application form: %v, email: %b", payment.ReferenceNumber, applicationForm.Email))
						}
					}
				}
				// update paymnet_email_sent to true
				_, err := svc.db.Exec("UPDATE t_course_application_form SET payment_email_sent = 1 WHERE ID = @p1", refNumber)
				if err != nil {
					log.Err(err).Msgf("Failed to update paymnet_email_sent:  %v", err)
					continue
				}
			}

		}

		// if !msgsToDelete.Empty() {
		// 	item := imap.FormatFlagsOp(imap.AddFlags, true)
		// 	flags := []interface{}{imap.DeletedFlag}
		// 	if err := svc.imapClient.Store(msgsToDelete, item, flags, nil); err != nil {
		// 		log.Err(err).Msgf("Error: %v", err)
		// 		return
		// 	}
		// 	// Then delete it
		// 	if err := svc.imapClient.Expunge(nil); err != nil {
		// 		log.Err(err).Msgf("Error: %v", err)
		// 		return
		// 	}
		// }

	})

	svc.cronScheduler.Start()

	return nil
}

func stringBetween(str, start, end string) string {
	s := strings.Index(str, start)
	if s == -1 {
		return ""
	}
	s += len(start)
	e := strings.Index(str[s:], end)
	if e == -1 {
		return ""
	}
	return str[s : s+e]
}

func decodeQuotedPrintable(input string) (string, error) {
	reader := quotedprintable.NewReader(strings.NewReader(input))
	decoded, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

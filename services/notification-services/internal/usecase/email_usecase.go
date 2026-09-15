package usecase

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"math"
	"net/smtp"
	"notification-services/internal/config"
	"notification-services/internal/dto"
	"notification-services/internal/helper"
	"notification-services/internal/models"
	"notification-services/internal/repository"
	"strings"
	"time"
)

type EmailUsecaseInterface interface {
	SaveEmailLog(ctx context.Context, payload *dto.UserCreatedConsumer) (*dto.EmailLogDetailResponse, error)
	GetAllEmailLogs(ctx context.Context, page int, limit int) ([]dto.EmailLogAllResponse, helper.PaginationMeta, error)
	GetEmailLogByID(ctx context.Context, ID string) (*dto.EmailLogDetailResponse, error)
}

type EmailUsecase struct {
	repository repository.EmailRepositoryInterface
	cfg        *config.SMTPConfig
}

func NewEmailUsecase(emailRepository repository.EmailRepositoryInterface, cfg *config.SMTPConfig) *EmailUsecase {
	return &EmailUsecase{
		repository: emailRepository,
		cfg:        cfg,
	}
}

func (u *EmailUsecase) SaveEmailLog(ctx context.Context, payload *dto.UserCreatedConsumer) (*dto.EmailLogDetailResponse, error) {

	subject := "Verify Your Account Please!"

	bodyHTML := fmt.Sprintf(`<!DOCTYPE html>
			<html lang="en">
			<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<meta name="color-scheme" content="light only">
			<meta name="supported-color-schemes" content="light only">
			<title>Verify Your Account</title>
			</head>
			<body style="margin:0; padding:0; background-color:#eef1f6; font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;">

			<div style="display:none; max-height:0; overflow:hidden; opacity:0;">Confirm your email address to activate your account.</div>

			<table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0" style="background-color:#eef1f6;">
			<tr>
			<td align="center" style="padding:40px 16px;">

			<table role="presentation" width="600" cellpadding="0" cellspacing="0" border="0" style="width:100%%; max-width:600px; background-color:#ffffff; border-radius:16px; border:1px solid #e3e8ef;">

			<tr>
			<td align="center" bgcolor="#111c35" style="background-color:#111c35; padding:38px 40px; border-radius:16px 16px 0 0;">
			<div style="font-size:12px; letter-spacing:3px; text-transform:uppercase; color:#8ea3c9;">Library App</div>
			<div style="margin-top:12px; font-size:24px; font-weight:600; color:#ffffff;">Verify your account</div>
			</td>
			</tr>

			<tr>
			<td style="padding:44px 40px 32px 40px;">
			<p style="margin:0 0 18px 0; font-size:19px; font-weight:600; color:#111c35;">Halo %[1]s %[2]s,</p>
			<p style="margin:0 0 34px 0; font-size:15px; line-height:1.7; color:#4a5568;">Thank you for registering. Please click the button below to verify your account:</p>

			<table role="presentation" cellpadding="0" cellspacing="0" border="0" align="center" style="margin:0 auto;">
			<tr>
			<td align="center" bgcolor="#2f6bff" style="background-color:#2f6bff; border-radius:10px;">
			<a href="%[3]s" target="_blank" style="display:inline-block; padding:16px 42px; font-size:16px; font-weight:600; color:#ffffff; text-decoration:none; border-radius:10px;">Verify My Email</a>
			</td>
			</tr>
			</table>

			<p style="margin:34px 0 8px 0; font-size:13px; color:#8a94a6;">Or copy this link into your browser:</p>
			<p style="margin:0; font-size:13px; line-height:1.6; word-break:break-all;"><a href="%[3]s" style="color:#2f6bff;">%[3]s</a></p>
			</td>
			</tr>

			<tr>
			<td style="padding:0 40px;">
			<div style="height:1px; background-color:#e8ecf3; font-size:0; line-height:0;">&nbsp;</div>
			</td>
			</tr>

			<tr>
			<td style="padding:24px 40px 36px 40px;">
			<p style="margin:0; font-size:13px; line-height:1.7; color:#8a94a6;">This link is valid for a limited time. If you did not create this account, you can safely ignore this email.</p>
					<p style="margin:0; font-size:13px; line-height:1.7; color:#8a94a6;">Please Do Not Reply to This Email. Because it is automatically generated.</p>
			</td>
			</tr>

			</table>

			<p style="margin:24px 0 0 0; font-size:12px; color:#9aa4b6;">&copy; Library App</p>

			</td>
			</tr>
			</table>

			</body>
			</html>`, payload.FirstName, payload.LastName, payload.VerificationLink)

	errSend := u.sendSMTP(payload.Email, subject, bodyHTML)

	if errSend != nil {
		errMessage := errSend.Error()
		emailLog := &models.EmailLogs{
			UserID:       payload.KeycloakID,
			Email:        payload.Email,
			FirstName:    payload.FirstName,
			LastName:     payload.LastName,
			Subject:      subject,
			Status:       models.EmailStatusFailed,
			ErrorMessage: &errMessage,
		}

		errSave := u.repository.SaveEmailLog(ctx, emailLog)

		if errSave != nil {
			return nil, helper.NewInternalServerError("An Error During Save Email Log", helper.ErrorDetail{Detail: errSave.Error()})
		}

		result := &dto.EmailLogDetailResponse{
			ID:           emailLog.ID,
			UserID:       emailLog.UserID,
			Email:        emailLog.Email,
			FirstName:    emailLog.FirstName,
			LastName:     emailLog.LastName,
			Subject:      emailLog.Subject,
			Status:       string(emailLog.Status),
			ErrorMessage: emailLog.ErrorMessage,
			CreatedAt:    helper.FormatTimeRFC3339Jakarta(emailLog.CreatedAt),
		}

		return result, nil
	}

	emailLog := &models.EmailLogs{
		UserID:       payload.KeycloakID,
		Email:        payload.Email,
		FirstName:    payload.FirstName,
		LastName:     payload.LastName,
		Subject:      subject,
		Status:       models.EmailStatusSuccess,
		ErrorMessage: nil,
	}

	errSave := u.repository.SaveEmailLog(ctx, emailLog)

	if errSave != nil {
		return nil, helper.NewInternalServerError("An Error During Save Email Log", helper.ErrorDetail{Detail: errSave.Error()})
	}

	result := &dto.EmailLogDetailResponse{
		ID:           emailLog.ID,
		UserID:       emailLog.UserID,
		Email:        emailLog.Email,
		FirstName:    emailLog.FirstName,
		LastName:     emailLog.LastName,
		Subject:      emailLog.Subject,
		Status:       string(emailLog.Status),
		ErrorMessage: emailLog.ErrorMessage,
		CreatedAt:    helper.FormatTimeRFC3339Jakarta(emailLog.CreatedAt),
	}

	return result, nil

}

func (u *EmailUsecase) sendSMTP(toEmail string, subject string, bodyHTML string) error {

	auth := smtp.PlainAuth("", u.cfg.SMTPUser, u.cfg.SMTPPassword, u.cfg.SMTPHost)

	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return err
	}

	domain := "library.com"
	if parts := strings.SplitN(u.cfg.SMTPUser, "@", 2); len(parts) == 2 {
		domain = parts[1]
	}

	headers := make(map[string]string)
	headers["From"] = u.cfg.SMTPUser
	headers["To"] = toEmail
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"UTF-8\""
	headers["From"] = fmt.Sprintf("Library App <%s>", u.cfg.SMTPUser)
	headers["Date"] = time.Now().Format(time.RFC1123Z)
	headers["Message-ID"] = fmt.Sprintf("<%x@%s>", randomBytes, domain)

	message := ""

	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + bodyHTML

	addr := fmt.Sprintf("%s:%s", u.cfg.SMTPHost, u.cfg.SMTPPort)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         u.cfg.SMTPHost,
	}

	client, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer client.Close()

	if err = client.StartTLS(tlsConfig); err != nil {
		return err
	}

	if err = client.Auth(auth); err != nil {
		return err
	}

	if err = client.Mail(u.cfg.SMTPUser); err != nil {
		return err
	}

	if err = client.Rcpt(toEmail); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		return err
	}

	return w.Close()

}

func (u *EmailUsecase) GetAllEmailLogs(ctx context.Context, page int, limit int) ([]dto.EmailLogAllResponse, helper.PaginationMeta, error) {

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	offset := (page - 1) * limit

	emailLogs, totalData, errGet := u.repository.GetAllEmailLogs(ctx, limit, offset)

	if errGet != nil {
		return nil, helper.PaginationMeta{}, helper.NewInternalServerError("An Error During Get All Email Logs", helper.ErrorDetail{Detail: errGet.Error()})
	}

	result := make([]dto.EmailLogAllResponse, 0, len(emailLogs))

	for _, value := range emailLogs {
		result = append(result, dto.EmailLogAllResponse{
			ID:        value.ID,
			Email:     value.Email,
			Subject:   value.Subject,
			Status:    string(value.Status),
			CreatedAt: helper.FormatTimeRFC3339Jakarta(value.CreatedAt),
		})
	}

	totalPage := int(math.Ceil(float64(totalData) / float64(limit)))

	pagination := helper.PaginationMeta{
		TotalData: totalData,
		TotalPage: totalPage,
		Page:      page,
		Limit:     limit,
		Keywords:  nil,
	}

	return result, pagination, nil

}

func (u *EmailUsecase) GetEmailLogByID(ctx context.Context, ID string) (*dto.EmailLogDetailResponse, error) {

	emailLog, errGet := u.repository.GetEmailLogByID(ctx, ID)

	if errGet != nil {
		return nil, helper.NewInternalServerError("An Error During Get Email Log By ID", helper.ErrorDetail{Detail: errGet.Error()})
	}

	result := &dto.EmailLogDetailResponse{
		ID:           emailLog.ID,
		UserID:       emailLog.UserID,
		Email:        emailLog.Email,
		FirstName:    emailLog.FirstName,
		LastName:     emailLog.LastName,
		Subject:      emailLog.Subject,
		Status:       string(emailLog.Status),
		ErrorMessage: emailLog.ErrorMessage,
		CreatedAt:    helper.FormatTimeRFC3339Jakarta(emailLog.CreatedAt),
	}

	return result, nil

}

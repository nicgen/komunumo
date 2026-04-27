package fakes

import (
	"context"
	"errors"

	"komunumo/backend/internal/ports"
)

var _ ports.EmailSender = (*EmailSender)(nil)

type EmailCall struct {
	Method      string
	To          string
	DisplayName string
	RawToken    string
}

type EmailSender struct {
	Calls   []EmailCall
	failOn  string
	failErr error
}

func NewEmailSender() *EmailSender { return &EmailSender{} }

func (s *EmailSender) FailOn(method string, err error) {
	s.failOn = method
	s.failErr = err
}

func (s *EmailSender) check(method string) error {
	if s.failOn == method {
		return s.failErr
	}
	return nil
}

func (s *EmailSender) SendVerification(_ context.Context, to, displayName, rawToken string) error {
	if err := s.check("SendVerification"); err != nil {
		return err
	}
	s.Calls = append(s.Calls, EmailCall{"SendVerification", to, displayName, rawToken})
	return nil
}

func (s *EmailSender) SendAccountAlreadyExists(_ context.Context, to, displayName string) error {
	if err := s.check("SendAccountAlreadyExists"); err != nil {
		return err
	}
	s.Calls = append(s.Calls, EmailCall{"SendAccountAlreadyExists", to, displayName, ""})
	return nil
}

func (s *EmailSender) SendPasswordReset(_ context.Context, to, displayName, rawToken string) error {
	if err := s.check("SendPasswordReset"); err != nil {
		return err
	}
	s.Calls = append(s.Calls, EmailCall{"SendPasswordReset", to, displayName, rawToken})
	return nil
}

func (s *EmailSender) SendPasswordChanged(_ context.Context, to, displayName string) error {
	if err := s.check("SendPasswordChanged"); err != nil {
		return err
	}
	s.Calls = append(s.Calls, EmailCall{"SendPasswordChanged", to, displayName, ""})
	return nil
}

func (s *EmailSender) Called(method string) bool {
	for _, c := range s.Calls {
		if c.Method == method {
			return true
		}
	}
	return false
}

var ErrEmailFailed = errors.New("fake: email send failed")

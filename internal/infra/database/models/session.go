package models

import (
	"time"
	sessionDomain "wazmeow/internal/domain/session"
)

// SessionModel represents the database model for sessions
type SessionModel struct {
	ID        string    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Status    string    `db:"status" json:"status"`
	WAJID     string    `db:"wa_jid" json:"wa_jid"`
	QRCode    string    `db:"qr_code" json:"qr_code"`
	ProxyURL  string    `db:"proxy_url" json:"proxy_url"`
	IsActive  bool      `db:"is_active" json:"is_active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// TableName returns the table name for the session model
func (SessionModel) TableName() string {
	return "sessions"
}

// ToEntity converts the database model to domain entity
func (m *SessionModel) ToEntity() *sessionDomain.Session {
	return sessionDomain.RestoreSession(
		sessionDomain.SessionID(m.ID),
		m.Name,
		sessionDomain.Status(m.Status),
		m.WAJID,
		m.QRCode,
		m.ProxyURL,
		m.IsActive,
		m.CreatedAt,
		m.UpdatedAt,
	)
}

// FromEntity converts domain entity to database model
func (m *SessionModel) FromEntity(s *sessionDomain.Session) {
	m.ID = s.ID().String()
	m.Name = s.Name()
	m.Status = string(s.Status())
	m.WAJID = s.WAJID()
	m.QRCode = s.QRCode()
	m.ProxyURL = s.ProxyURL()
	m.IsActive = s.IsActive()
	m.CreatedAt = s.CreatedAt()
	m.UpdatedAt = s.UpdatedAt()
}

// NewSessionModelFromEntity creates a new session model from domain entity
func NewSessionModelFromEntity(s *sessionDomain.Session) *SessionModel {
	model := &SessionModel{}
	model.FromEntity(s)
	return model
}

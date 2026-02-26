package models

import "time"

type Course struct {
	Id               uint       `json:"id" gorm:"primaryKey"`
	Title            string     `json:"title" gorm:"not null"`
	Issuer           string     `json:"issuer"`
	IssuedAt         *time.Time `json:"issued_at"`
	ExpiredAt        *time.Time `json:"expired_at"`
	CredentialUrl    string     `json:"credential_url"`
	CertificateImage string     `json:"certificate_image"`
	Description      string     `json:"description" gorm:"type:text"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

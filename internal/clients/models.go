package clients

import "strings"

type Client struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	LegalName   string `json:"legalName"`
	Document    string `json:"document"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	ContactName string `json:"contactName"`
	Address     string `json:"address"`
	City        string `json:"city"`
	State       string `json:"state"`
	PostalCode  string `json:"postalCode"`
	Description string `json:"description"`
	Notes       string `json:"notes"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type Project struct {
	ID           string `json:"id"`
	ClientID     string `json:"clientId"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Location     string `json:"location"`
	Address      string `json:"address"`
	LocalContact string `json:"localContact"`
	Notes        string `json:"notes"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

func (c *Client) Validate() error {
	c.Name = strings.TrimSpace(c.Name)
	if c.Name == "" {
		return ValidationError{Field: "name", Message: "Nome é obrigatório"}
	}
	if len(c.Name) > 160 {
		return ValidationError{Field: "name", Message: "Nome deve ter no máximo 160 caracteres"}
	}
	return validateFields(map[string]string{"legal_name": c.LegalName, "document": c.Document, "phone": c.Phone, "email": c.Email, "contact_name": c.ContactName, "address": c.Address, "city": c.City, "state": c.State, "postal_code": c.PostalCode, "description": c.Description, "notes": c.Notes})
}
func (p *Project) Validate() error {
	p.Name = strings.TrimSpace(p.Name)
	if p.Name == "" {
		return ValidationError{Field: "name", Message: "Nome é obrigatório"}
	}
	if len(p.Name) > 160 {
		return ValidationError{Field: "name", Message: "Nome deve ter no máximo 160 caracteres"}
	}
	return validateFields(map[string]string{"description": p.Description, "location": p.Location, "address": p.Address, "local_contact": p.LocalContact, "notes": p.Notes})
}
func validateFields(values map[string]string) error {
	for field, value := range values {
		if len(value) > 4000 {
			return ValidationError{Field: field, Message: "Campo deve ter no máximo 4000 caracteres"}
		}
	}
	return nil
}

type ValidationError struct{ Field, Message string }

func (e ValidationError) Error() string { return e.Message }

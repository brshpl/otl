package entity

// OneTimeLink -.
type OneTimeLink struct {
	Data    string `json:"data"`
	Link    string `json:"link"`
	Expired bool   `json:"expired"`
}

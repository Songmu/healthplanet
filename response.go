package healthplanet

type Response struct {
	BirthDate string  `json:"birth_date"`
	Data      []*Data `json:"data"`
	Height    string  `json:"height"`
	Sex       string  `json:"sex"`
}

type Data struct {
	Date    string `json:"date"`
	KeyData string `json:"keydata"`
	Model   string `json:"model"`
	Tag     string `json:"tag"`
}

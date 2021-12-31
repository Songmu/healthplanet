package healthplanet

type response struct {
	BirthDate string  `json:"birth_date"`
	Data      []*data `json:"data"`
	Height    string  `json:"height"`
	Sex       string  `json:"sex"`
}

type data struct {
	Date    string `json:"date"`
	KeyData string `json:"keydata"`
	Model   string `json:"model"`
	Tag     string `json:"tag"`
}

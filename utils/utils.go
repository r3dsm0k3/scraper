package utils

// DataPoints composition
type PotentialApartment struct {
	URL      string `json:"url"`
	Rent     string `json:"rent"`
	Location string `json:"location"`
	ZipCode  string `json:"zip"`
}

type Queue struct {
	Channel chan PotentialApartment
}

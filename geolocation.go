package dodod

//type GeoLocation struct {
//	Accuracy string `json:"-"`
//	Lon float64 `json:"-"`
//	Lat float64 `json:"-"`
//	Data string `json:"data"`
//}
//
//func (u *GeoLocation) MarshalJSON() ([]byte, error) {
//
//	if data, err := json.Marshal(&struct {
//		Lon     float64 `json:"lon"`
//		Lat     float64  `json:"lat"`
//	}{
//		Lon: u.Lon,
//		Lat: u.Lat,
//	}); err == nil {
//		u.Data = string(data)
//	}
//}
//
//func (u *GeoLocation) UnmarshalJSON(data []byte) error {
//	aux := &struct{
//		Lon     float64 `json:"lon"`
//		Lat     float64  `json:"lat"`
//	}{
//	}
//	if err := json.Unmarshal(data, &aux); err != nil {
//		return err
//	}
//	u.Lat = aux.Lat
//	u.Lon = aux.Lon
//	return nil
//}

//type GeoHash string

//func (g *GeoLocation) Type() string {
//	return "GeoLocation"
//}

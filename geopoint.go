package dgogm

type GeoPoint struct {
	Type       string            `json:"type"`
	Geometry   GeoGeometry       `json:"geometry"`
	Properties map[string]string `json:"properties"`
}

type GeoGeometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

func (gp *GeoPoint) Json() *string {
	return StrPtr(ToJsonUnsafe(gp))
}

package main

type code struct {
	Name    string   `json:"name"`
	Domains []string `json:"domains"`
}

type layer struct {
	Name  string `json:"name"`
	Codes []code `json:"data"`
}

type scheme struct {
	Layers []layer `json:"layers"`
}

func (layer *layer) getCode(codeName string) (*code, bool) {
	// Check if it is there... return if it is
	for i, code := range layer.Codes {
		if codeName == code.Name {
			return &layer.Codes[i], true
		}
	}
	// If not there, crate it, add it, extend, and return
	var newCode code
	newCode.Name = codeName
	newCode.Domains = make([]string, 0)
	layer.Codes = append(layer.Codes, newCode)
	return &layer.Codes[len(layer.Codes)-1], false
}

func (scheme *scheme) getLayer(layerName string) (*layer, bool) {
	// Check if it is there... return if it is
	for i, layer := range scheme.Layers {
		if layerName == layer.Name {
			return &scheme.Layers[i], true
		}
	}
	// If not there, crate it, add it, extend, and return
	var newLayer layer
	newLayer.Name = layerName
	newLayer.Codes = make([]code, 0)
	scheme.Layers = append(scheme.Layers, newLayer)
	return &scheme.Layers[len(scheme.Layers)-1], false
}

func newScheme() scheme {
	var ret scheme

	// This is how we set the right order for the
	// layers... from Outside to Inside.
	ret.getLayer("internal elements")
	ret.getLayer("environmental cues")
	ret.getLayer("objective indoor climatic factors")
	ret.getLayer("perceptions")
	ret.getLayer("trade-offs")
	ret.getLayer("expected outcomes")
	ret.getLayer("comfort")

	return ret
}

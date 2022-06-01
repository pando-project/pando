package controller

import "testing"

func TestController_SchemaRegister(t *testing.T) {
	c := Controller{}
	err := c.SchemaRegister([]byte(`type MinerLocationsModel struct {
    Epoch Int
    Date String
    MinerLocations [MinerLocationModel]
}

type MinerLocationModel struct {
    Miner String
    Region String
    Long Float
    Lat Float
    NumLocations Int
    Country String
    City String
    SubDiv1 String
}
`))
	if err != nil {
		t.Error(err)
	}
}

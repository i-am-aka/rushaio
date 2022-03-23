package akgen

import (
	// "encoding/json"
	"io/ioutil"
	"github.com/stretchr/testify/assert"
	// "os"
	"path/filepath"
	"testing"
	"strings"
)

// func TestSerialize(t *testing.T) {
// 	assert := assert.New(t)
// 	assert.Nil(nil)
// }

// func TestParse(t *testing.T) {
// 	assert := assert.New(t)

// 	sdjPath := filepath.Join(CWD, "test", "sd.json")
// 	sdj, err := ioutil.ReadFile(sdjPath)
// 	assert.Nil(err)

// 	var sdfixture AkWebSensorData
// 	err = json.Unmarshal([]byte(sdj), &sdfixture)
// 	assert.Nil(err)

// 	sdstrPath := filepath.Join(CWD, "test", "sensor.txt")
// 	sdstr, err := ioutil.ReadFile(sdstrPath)
// 	assert.Nil(err)

// 	sdgen, err := ParseAkWebSensor(string(sdstr))
// 	assert.Nil(err)
// 	assert.Equal(sdgen, sdfixture)
// }

func TestRoundtrip(t *testing.T) {
	assert := assert.New(t)

	sdstrPath := filepath.Join(CWD, "test", "sensor.txt")
	sdbytes, err := ioutil.ReadFile(sdstrPath)
	assert.Nil(err)
	sdstr := strings.TrimSpace(string(sdbytes))

	sdgen, err := ParseAkWebSensor(sdstr)
	assert.Nil(err)
	sdgenstr := sdgen.Serialize()

	// sdgenstrparsed, err := ParseAkWebSensor(sdgenstr)
	// assert.Nil(err)
	// assert.Equal(sdgenstrparsed, sdgen)
	assert.Equal(sdgenstr, sdstr)
}

package hub

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"

	"github.com/qredo/signing-agent/defs"
)

func TestDefaultDialer(t *testing.T) {
	//Arrange
	sut := NewDefaultDialer()

	//Act
	res := sut.(*defaultDialer)

	//Assert
	assert.NotNil(t, res)
	assert.IsType(t, &websocket.Dialer{}, res.dialer)
}

func TestDefaultDialer_dial_invalid_url(t *testing.T) {
	//Arrange
	sut := NewDefaultDialer()
	headers := http.Header{}
	headers.Set(defs.AuthHeader, "some zkp oone pass")

	//Act
	res, _, err := sut.Dial("some url", headers)

	//Assert
	assert.NotNil(t, err)
	assert.Contains(t, "malformed ws or wss URL", err.Error())
	assert.Nil(t, res)
}

func TestDefaultUpgrader(t *testing.T) {
	//Arrange
	sut := NewDefaultUpgrader(500, 1024)

	//Act
	res := sut.(*defaultUpgrader)

	//Assert
	assert.NotNil(t, res)
	assert.IsType(t, &websocket.Upgrader{}, res.upgrader)
}

func TestDefaultUpgrader_upgrade_wrong_protocol(t *testing.T) {
	//Arrange
	sut := NewDefaultUpgrader(500, 1024)
	request, _ := http.NewRequest("GET", "/path", nil)

	//Act
	res, err := sut.Upgrade(httptest.NewRecorder(), request, nil)

	//Assert
	assert.NotNil(t, err)
	assert.Contains(t, "websocket: the client is not using the websocket protocol: 'upgrade' token not found in 'Connection' header", err.Error())
	assert.Nil(t, res)
}

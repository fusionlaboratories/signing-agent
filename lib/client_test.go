package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/jinzhu/copier"
	"github.com/stretchr/testify/assert"

	"github.com/qredo/signing-agent/api"
	"github.com/qredo/signing-agent/config"
	"github.com/qredo/signing-agent/util"
)

const (
	FixturePathRegisterClientInitResponse = "../testdata/lib/registerClientInitResponse.json"
)

func TestClient(t *testing.T) {
	var (
		cfg *config.Config
		err error
	)
	cfg = &config.Config{
		Base: config.Base{
			PIN:      1234,
			QredoAPI: "https://play-api.qredo.network/api/v1/p",
		},
		AutoApprove: config.AutoApprove{
			Enabled: true,
		},
	}

	agentName := "Test name agent"
	kv := util.NewFileStore(TestDataDBStoreFilePath)
	err = kv.Init()
	assert.NoError(t, err)
	defer func() {
		err = os.Remove(TestDataDBStoreFilePath)
		assert.NoError(t, err)
	}()

	core, err := NewMock(cfg, kv)
	assert.NoError(t, err)

	var (
		pendingAgent           *Agent
		registerResponse       *api.ClientRegisterResponse
		initRequest            *api.QredoRegisterInitRequest
		initResponse           *api.QredoRegisterInitResponse
		registerFinishRequest  *api.ClientRegisterFinishRequest
		registerFinishResponse *api.ClientRegisterFinishResponse
	)

	t.Run(
		"Register agent - first step",
		func(t *testing.T) {

			registerResponse, err = core.ClientRegister(agentName)
			assert.NoError(t, err)
			assert.NotEmpty(t, registerResponse.BLSPublicKey)
			assert.NotEmpty(t, registerResponse.ECPublicKey)
			assert.NotEmpty(t, registerResponse.RefID)
			pendingAgent = core.store.GetPending(registerResponse.RefID)
			assert.Equal(t, true, pendingAgent.Pending)
			assert.Equal(t, agentName, pendingAgent.Name)
			assert.Empty(t, pendingAgent.ID)
		})

	t.Run(
		"Register agent - call init ep",
		func(t *testing.T) {
			initRequest = &api.QredoRegisterInitRequest{
				BLSPublicKey: registerResponse.BLSPublicKey,
				ECPublicKey:  registerResponse.ECPublicKey,
				Name:         agentName,
			}

			util.GetDoMockHTTPClientFunc = func(*http.Request) (*http.Response, error) {
				dataFromFixture, err := os.Open(FixturePathRegisterClientInitResponse)
				assert.NoError(t, err)
				body := io.NopCloser(dataFromFixture)

				return &http.Response{
					Status:     "200 OK",
					StatusCode: 200,
					Body:       body,
				}, nil

			}
			apikey := "SoMeTestAPIKey=="
			b64PrvKey := generatePrivateKeyBase64()
			initResponse, err = core.ClientInit(initRequest, registerResponse.RefID, apikey, b64PrvKey)
			assert.NoError(t, err)
			assert.NotEmpty(t, initResponse.AccountCode)

		})

	t.Run(
		"Register agent - finish registration",
		func(t *testing.T) {
			util.GetDoMockHTTPClientFunc = func(*http.Request) (*http.Response, error) {

				response := &api.CoreClientServiceRegisterFinishResponse{
					Feed: fmt.Sprintf(
						"%s/coreclient/%s/feed",
						core.cfg.Websocket.QredoWebsocket,
						initResponse.AccountCode,
					),
				}

				dataJSON, _ := json.Marshal(response)
				body := io.NopCloser(bytes.NewReader(dataJSON))

				return &http.Response{
					Status:     "200 OK",
					StatusCode: 200,
					Body:       body,
				}, nil

			}
			registerFinishRequest = &api.ClientRegisterFinishRequest{}
			_ = copier.Copy(&registerFinishRequest, &initResponse)
			registerFinishResponse, err = core.ClientRegisterFinish(registerFinishRequest, registerResponse.RefID)
			assert.NoError(t, err)
			assert.NotEmpty(t, registerFinishResponse.FeedURL)

			// logic verification after registration process
			assert.NotEmpty(t, core.GetSystemAgentID(), "At this stage, we should be able to get AgentID")
			assert.Nil(t, core.store.GetPending(registerResponse.RefID), "At this stage, we shouldn't get pending agent")
			registeredAgent := core.store.GetAgent(initResponse.AccountCode)
			assert.NotNil(t, registeredAgent, "At this stage, we should get agent")
			assert.False(t, registeredAgent.Pending, "Agent is not any more at Pending process.")
			assert.NotEmpty(t, registeredAgent.ID, "At this stage, agent if created properly")
			assert.NotEmpty(t, registeredAgent.Name, "At this stage, agent if created properly")
			assert.NotEmpty(t, registeredAgent.AccountCode, "At this stage, agent if created properly")
			assert.NotEmpty(t, registeredAgent.BLSSeed, "At this stage, agent if created properly")
			assert.NotEmpty(t, registeredAgent.ZKPID, "At this stage, agent if created properly")
			assert.NotEmpty(t, registeredAgent.ZKPToken, "At this stage, agent if created properly")
		})

	t.Run(
		"Register agent - finish registration fake Agent ID",
		func(t *testing.T) {
			registerFinishRequestFake := &api.ClientRegisterFinishRequest{}
			_, err = core.ClientRegisterFinish(registerFinishRequestFake, "fake RefID")
			assert.Error(t, err)
			_, err = core.ClientRegisterFinish(registerFinishRequestFake, registerResponse.RefID)
			assert.Error(t, err)
			registerFinishRequestFake.ClientID = initResponse.AccountCode
			_, err = core.ClientRegisterFinish(registerFinishRequestFake, registerResponse.RefID)
			assert.Error(t, err)
			registerFinishRequestFake.ClientSecret = initResponse.ClientSecret
			_, err = core.ClientRegisterFinish(registerFinishRequestFake, registerResponse.RefID)
			assert.Error(t, err)
		})

	t.Run(
		"GetAgentID",
		func(t *testing.T) {
			agentID := core.GetAgentID()
			assert.Equal(t, initResponse.AccountCode, agentID)
		})

	t.Run(
		"Agent - setting and getting",
		func(t *testing.T) {
			agentID := "BbCoiGKwPfc4DYWE6mE2zAEeuEowXLE8sk1Tc9TN8tos"
			_ = core.SetSystemAgentID(agentID)
			assert.Equal(t, core.GetSystemAgentID(), agentID)

			res := core.GetAgentID()
			assert.Equal(t, agentID, res)
		})
}

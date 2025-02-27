package switchbot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type WebhookService struct {
	c *Client
}

func newWebhookService(c *Client) *WebhookService {
	return &WebhookService{c: c}
}

func (c *Client) Webhook() *WebhookService {
	return c.webhookService
}

type webhookSetupRequest struct {
	Action     string `json:"action"`
	URL        string `json:"url,omitempty"`
	DeviceList string `json:"deviceList,omitempty"` // currently only ALL is supported
}

type webhookSetupResponse struct {
	StatusCode int         `json:"statusCode"`
	Body       interface{} `json:"body"`
	Message    string      `json:"message"`
}

// Setup configures the url that all the webhook events will be sent to.
// Currently the deviceList is only supporting "ALL".
func (svc *WebhookService) Setup(ctx context.Context, url, deviceList string) (string, error) {
	const path = "/v1.1/webhook/setupWebhook"

	if deviceList != "ALL" {
		return "", errors.New(`deviceList value is only supporting "ALL" for now`)
	}

	req := webhookSetupRequest{
		Action:     "setupWebhook",
		URL:        url,
		DeviceList: deviceList,
	}

	resp, err := svc.c.post(ctx, path, req)
	if err != nil {
		return "", err
	}
	byteArray, _ := ioutil.ReadAll(resp.Body)
	r := string(byteArray)
	defer resp.Close()

	return r, nil
}

type webhookQueryRequest struct {
	Action WebhookQueryActionType `json:"action"`
	URLs   []string               `json:"urls"`
}

type WebhookQueryActionType string

const (
	QueryURL     WebhookQueryActionType = "queryUrl"
	QueryDetails WebhookQueryActionType = "queryDetails"
)

type webhookQueryResponse struct {
}

// Query retrieves the current configuration info of the webhook.
// The second argument `url` is required for QueryDetails action type.
func (svc *WebhookService) Query(ctx context.Context, action WebhookQueryActionType, url string) (string, error) {
	const path = "/v1.1/webhook/queryWebhook"

	req := webhookQueryRequest{
		Action: action,
	}

	switch action {
	case QueryDetails:
		if url == "" {
			return "", errors.New("URL need to be specified when the action is queryDetails")
		}

		req.URLs = []string{url}
	}

	resp, err := svc.c.post(ctx, path, req)
	if err != nil {
		return "", err
	}
	byteArray, _ := ioutil.ReadAll(resp.Body)
	r := string(byteArray)

	defer resp.Close()

	return r, nil
}

type webhookUpdateRequest struct {
	Action string        `json:"action"` // the type of actions but for now this must be updateWebhook
	Config webhookConfig `json:"config"`
}

type webhookConfig struct {
	URL    string `json:"url"`
	Enable bool   `json:"enable"`
}

// Update do update the configuration of the webhook.
func (svc *WebhookService) Update(ctx context.Context, url string, enable bool) (string, error) {
	const path = "/v1.1/webhook/queryWebhook"

	req := webhookUpdateRequest{
		Action: "updateWebhook",
		Config: webhookConfig{
			URL:    url,
			Enable: enable,
		},
	}

	resp, err := svc.c.post(ctx, path, req)
	if err != nil {
		return "", err
	}
	byteArray, _ := ioutil.ReadAll(resp.Body)
	r := string(byteArray)
	defer resp.Close()

	return r, nil
}

type webhookDeleteRequest struct {
	Action string `json:"action"` // the type of actions but for now this must be deleteWebhook
	URL    string `json:"url"`
}

// Delete do delete the configuration of the webhook.
func (svc *WebhookService) Delete(ctx context.Context, url string) (string, error) {
	const path = "/v1.1/webhook/deleteWebhook"

	req := webhookDeleteRequest{
		Action: "deleteWebhook",
		URL:    url,
	}

	resp, err := svc.c.del(ctx, path, req)
	if err != nil {
		return "", err
	}
	byteArray, _ := ioutil.ReadAll(resp.Body)
	r := string(byteArray)
	defer resp.Close()

	return r, nil
}

func deviceTypeFromWebhookRequest(r *http.Request) (string, error) {
	var rawBody bytes.Buffer
	var deviceTypeBody struct {
		Context struct {
			DeviceType string `json:"deviceType"`
		} `json:"context"`
	}

	if err := json.NewDecoder(io.TeeReader(r.Body, &rawBody)).Decode(&deviceTypeBody); err != nil {
		return "", err
	}

	r.Body = io.NopCloser(&rawBody)

	return deviceTypeBody.Context.DeviceType, nil
}

type MotionSensorEvent struct {
	EventType    string                   `json:"eventType"`
	EventVersion string                   `json:"eventVersion"`
	Context      MotionSensorEventContext `json:"context"`
}

type MotionSensorEventContext struct {
	DeviceType   string `json:"deviceType"`
	DeviceMac    string `json:"deviceMac"`
	TimeOfSample int64  `json:"timeOfSample"`

	// the motion state of the device, "DETECTED" stands for motion is detected;
	// "NOT_DETECTED" stands for motion has not been detected for some time
	DetectionState string `json:"detectionState"`
}

type ContactSensorEvent struct {
	EventType    string                    `json:"eventType"`
	EventVersion string                    `json:"eventVersion"`
	Context      ContactSensorEventContext `json:"context"`
}

type ContactSensorEventContext struct {
	DeviceType   string `json:"deviceType"`
	DeviceMac    string `json:"deviceMac"`
	TimeOfSample int64  `json:"timeOfSample"`

	// the motion state of the device, "DETECTED" stands for motion is detected;
	// "NOT_DETECTED" stands for motion has not been detected for some time
	DetectionState string `json:"detectionState"`
	// when the enter or exit mode gets triggered, "IN_DOOR" or "OUT_DOOR" is returned
	DoorMode string `json:"doorMode"`
	// the level of brightness, can be "bright" or "dim"
	Brightness AmbientBrightness `json:"brightness"`
	// the state of the contact sensor, can be "open" or "close" or "timeOutNotClose"
	OpenState string `json:"openState"`
}

type MeterEvent struct {
	EventType    string            `json:"eventType"`
	EventVersion string            `json:"eventVersion"`
	Context      MeterEventContext `json:"context"`
}

type MeterEventContext struct {
	DeviceType   string `json:"deviceType"`
	DeviceMac    string `json:"deviceMac"`
	TimeOfSample int64  `json:"timeOfSample"`

	Temperature float64 `json:"temperature"`
	Scale       string  `json:"scale"`
	Humidity    int     `json:"humidity"`
}

type MeterPlusEvent struct {
	EventType    string                `json:"eventType"`
	EventVersion string                `json:"eventVersion"`
	Context      MeterPlusEventContext `json:"context"`
}

type MeterPlusEventContext struct {
	DeviceType   string `json:"deviceType"`
	DeviceMac    string `json:"deviceMac"`
	TimeOfSample int64  `json:"timeOfSample"`

	Temperature float64 `json:"temperature"`
	Scale       string  `json:"scale"`
	Humidity    int     `json:"humidity"`
}

type LockEvent struct {
	EventType    string           `json:"eventType"`
	EventVersion string           `json:"eventVersion"`
	Context      LockEventContext `json:"context"`
}

type LockEventContext struct {
	DeviceType   string `json:"deviceType"`
	DeviceMac    string `json:"deviceMac"`
	TimeOfSample int64  `json:"timeOfSample"`

	// the state of the device, "LOCKED" stands for the motor is rotated to locking position;
	// "UNLOCKED" stands for the motor is rotated to unlocking position; "JAMMED" stands for
	// the motor is jammed while rotating
	LockState string `json:"lockState"`
}

type IndoorCamEvent struct {
	EventType    string                `json:"eventType"`
	EventVersion string                `json:"eventVersion"`
	Context      IndoorCamEventContext `json:"context"`
}

type IndoorCamEventContext struct {
	DeviceType   string `json:"deviceType"`
	DeviceMac    string `json:"deviceMac"`
	TimeOfSample int64  `json:"timeOfSample"`

	// the detection state of the device, "DETECTED" stands for motion is detected
	DetectionState string `json:"detectionState"`
}

type PanTiltCamEvent struct {
	EventType    string                 `json:"eventType"`
	EventVersion string                 `json:"eventVersion"`
	Context      PanTiltCamEventContext `json:"context"`
}

type PanTiltCamEventContext struct {
	DeviceType   string `json:"deviceType"`
	DeviceMac    string `json:"deviceMac"`
	TimeOfSample int64  `json:"timeOfSample"`

	// the detection state of the device, "DETECTED" stands for motion is detected
	DetectionState string `json:"detectionState"`
}

type ColorBulbEvent struct {
	EventType    string                `json:"eventType"`
	EventVersion string                `json:"eventVersion"`
	Context      ColorBulbEventContext `json:"context"`
}

type ColorBulbEventContext struct {
	DeviceType   string `json:"deviceType"`
	DeviceMac    string `json:"deviceMac"`
	TimeOfSample int64  `json:"timeOfSample"`

	// the current power state of the device, "ON" or "OFF"
	PowerState PowerState `json:"powerState"`
	// the brightness value, range from 1 to 100
	Brightness int `json:"brightness"`
	// the color value, in the format of RGB value, "255:255:255"
	Color string `json:"color"`
	// the color temperature value, range from 2700 to 6500
	ColorTemperature int `json:"colorTemperature"`
}

type StripLightEvent struct {
	EventType    string                 `json:"eventType"`
	EventVersion string                 `json:"eventVersion"`
	Context      StripLightEventContext `json:"context"`
}

type StripLightEventContext struct {
	DeviceType   string `json:"deviceType"`
	DeviceMac    string `json:"deviceMac"`
	TimeOfSample int64  `json:"timeOfSample"`

	// the current power state of the device, "ON" or "OFF"
	PowerState PowerState `json:"powerState"`
	// the brightness value, range from 1 to 100
	Brightness int `json:"brightness"`
	// the color value, in the format of RGB value, "255:255:255"
	Color string `json:"color"`
}

type PlugMiniJPEvent struct {
	EventType    string                 `json:"eventType"`
	EventVersion string                 `json:"eventVersion"`
	Context      PlugMiniJPEventContext `json:"context"`
}

type PlugMiniJPEventContext struct {
	DeviceType   string `json:"deviceType"`
	DeviceMac    string `json:"deviceMac"`
	TimeOfSample int64  `json:"timeOfSample"`

	// the current power state of the device, "ON" or "OFF"
	PowerState PowerState `json:"powerState"`
}

type PlugMiniUSEvent struct {
	EventType    string                 `json:"eventType"`
	EventVersion string                 `json:"eventVersion"`
	Context      PlugMiniUSEventContext `json:"context"`
}

type PlugMiniUSEventContext struct {
	DeviceType   string `json:"deviceType"`
	DeviceMac    string `json:"deviceMac"`
	TimeOfSample int64  `json:"timeOfSample"`

	// the current power state of the device, "ON" or "OFF"
	PowerState PowerState `json:"powerState"`
}

type SweeperEvent struct {
	EventType    string              `json:"eventType"`
	EventVersion string              `json:"eventVersion"`
	Context      SweeperEventContext `json:"context"`
}

type SweeperEventContext struct {
	DeviceType   string `json:"deviceType"`
	DeviceMac    string `json:"deviceMac"`
	TimeOfSample int64  `json:"timeOfSample"`

	// the working status of the device, "StandBy", "Clearing",
	// "Paused", "GotoChargeBase", "Charging", "ChargeDone",
	// "Dormant", "InTrouble", "InRemoteControl", or "InDustCollecting"
	WorkingStatus CleanerWorkingStatus `json:"workingStatus"`
	// the connection status of the device, "online" or "offline"
	OnlineStatus CleanerOnlineStatus `json:"onlineStatus"`
	// the battery level.
	Battery int `json:"battery"`
}

type CeilingEvent struct {
	EventType    string              `json:"eventType"`
	EventVersion string              `json:"eventVersion"`
	Context      CeilingEventContext `json:"context"`
}

type CeilingEventContext struct {
	DeviceType   string `json:"deviceType"`
	DeviceMac    string `json:"deviceMac"`
	TimeOfSample int64  `json:"timeOfSample"`

	// ON/OFF state
	PowerState PowerState `json:"powerState"`
	// the brightness value, range from 1 to 100
	Brightness int `json:"brightness"`
	// the color temperature value, range from 2700 to 6500
	ColorTemperature int `json:"colorTemperature"`
}

type KeypadEvent struct {
	EventType    string             `json:"eventType"`
	EventVersion string             `json:"eventVersion"`
	Context      KeypadEventContext `json:"context"`
}

type KeypadEventContext struct {
	DeviceType   string `json:"deviceType"`
	DeviceMac    string `json:"deviceMac"`
	TimeOfSample int64  `json:"timeOfSample"`

	// the name fo the command being sent
	EventName string `json:"eventName"`
	// the command ID
	CommandID string `json:"commandId"`
	// the result of the command, success, failed, or timeout
	Result string `json:"result"`
}

func ParseWebhookRequest(r *http.Request) (interface{}, error) {
	deviceType, err := deviceTypeFromWebhookRequest(r)
	if err != nil {
		return nil, err
	}

	switch deviceType {
	case "WoPresence":
		// Motion Sensor
		var event MotionSensorEvent
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			return nil, err
		}
		return &event, nil
	case "WoContact":
		// Contact Sensor
		var event ContactSensorEvent
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			return nil, err
		}
		return &event, nil
	case "WoLock":
		// Lock
		var event LockEvent
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			return nil, err
		}
		return &event, nil
	case "WoCamera":
		// Indoor Cam
		var event IndoorCamEvent
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			return nil, err
		}
		return &event, nil
	case "WoPanTiltCam":
		// Pan/Tilt Cam
		var event PanTiltCamEvent
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			return nil, err
		}
		return &event, nil
	case "WoBulb":
		// Color Bulb
		var event ColorBulbEvent
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			return nil, err
		}
		return &event, nil
	case "WoStrip":
		// LED Strip Light
		var event StripLightEvent
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			return nil, err
		}
		return &event, nil
	case "WoPlugUS":
		// Plug Mini (US)
		var event PlugMiniUSEvent
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			return nil, err
		}
		return &event, nil
	case "WoPlugJP":
		// Plug Mini (JP)
		var event PlugMiniJPEvent
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			return nil, err
		}
		return &event, nil
	case "WoMeter":
		// Meter
		var event MeterEvent
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			return nil, err
		}
		return &event, nil
	case "WoMeterPlus":
		// Meter Plus
		var event MeterPlusEvent
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			return nil, err
		}
		return &event, nil
	case "WoSweeper", "WoSweeperPlus":
		// Cleaner
		var event SweeperEvent
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			return nil, err
		}
		return &event, nil
	case "WoCeiling", "WoCeilingPro":
		// Ceiling lights
		var event CeilingEvent
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			return nil, err
		}
		return &event, nil
	case "WoKeypad", "WoKeypadTouch":
		// keypad
		var event KeypadEvent
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			return nil, err
		}
		return &event, nil
	default:
		return nil, fmt.Errorf("unknown device type: %s", deviceType)
	}
}

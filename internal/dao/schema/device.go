package schema

import "fmt"

const (
	DevicePattern = "device:*"

	IdxDeviceUser = "idx_device_user"
)

var DeviceKey = func(userId, deviceId string) string { return fmt.Sprintf("device:%s:%s", userId, deviceId) }

type DeviceState int

const (
	DeviceStateOffline DeviceState = iota + 1 // offline
	DeviceStateOnline                         // online
)

type Device struct {
	Device string      `json:"device"` // device id
	User   string      `json:"user"`   // Owner user ID
	Name   string      `json:"name"`   // device host name
	State  DeviceState `json:"state"`  // 0: offline, 1: online
	Temp   bool        `json:"temp"`   // is temporary
}

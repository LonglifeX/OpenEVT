package types

import (
	"encoding/binary"
	"errors"
	"fmt"
)

var (
	ErrStatusFrameDecodeFailure = errors.New("decode: failed to decode evt status frame")
)

type InverterStatus struct {
	InverterId string
	Module1    InverterModuleStatus
	Module2    InverterModuleStatus
}

type InverterModuleStatus struct {
	ModuleId          string
	FirmwareVersion   string
	InputVoltageDC    float64
	OutputPowerAC     float64
	TotalEnergy       float64
	Temperature       float64
	OutputVoltageAC   float64
	OutputFrequencyAC float64
}

type rawInverterStatus struct {
	// reserved [0-5]
	_ uint32
	_ uint16

	// inverter ID (EVT ID) [6-9]
	InverterId uint32

	// reserved [10-19]
	_ uint64
	_ uint16

	// module 1 status frame [20-51]
	Module1 rawInverterModuleStatus

	// module 1 status frame [52-83]
	Module2 rawInverterModuleStatus

	// reserved [84-85]
	_        byte
	FrameEnd byte
}

type rawInverterModuleStatus struct {
	Id                   uint32
	FirmwareVersionMajor uint8
	FirmwareVersionMinor uint8
	InputVoltageDC       uint16
	OutputPowerAC        uint16
	TotalEnergy          uint32
	Temperature          uint16
	OutputVoltageAC      uint16
	OutputFrequencyAC    uint16

	// padding
	_ uint64
	_ uint32
}

func (s *InverterStatus) UnmarshalBinary(data []byte) error {
	var payload rawInverterStatus

	_, err := binary.Decode(data, binary.BigEndian, &payload)
	if err != nil {
		return errors.Join(ErrStatusFrameDecodeFailure, err)
	}

	if payload.FrameEnd != 0x16 {
		return errors.Join(ErrStatusFrameDecodeFailure, fmt.Errorf("(bug) unexpected frame end token [0x%x]", payload.FrameEnd))
	}

	s.InverterId = fmt.Sprintf("%x", payload.InverterId)
	s.Module1.ModuleId = fmt.Sprintf("%x", payload.Module1.Id)
	s.Module1.FirmwareVersion = fmt.Sprintf("%d/%d", payload.Module1.FirmwareVersionMajor, payload.Module1.FirmwareVersionMinor)
	s.Module1.InputVoltageDC = float64(payload.Module1.InputVoltageDC) / 512.0
	s.Module1.OutputPowerAC = float64(payload.Module1.OutputPowerAC) / 64.0
	s.Module1.TotalEnergy = float64(payload.Module1.TotalEnergy) / 8192.0
	s.Module1.Temperature = float64(payload.Module1.Temperature)/128.0 - 40.0
	s.Module1.OutputVoltageAC = float64(payload.Module1.OutputVoltageAC) / 64.0
	s.Module1.OutputFrequencyAC = float64(payload.Module1.OutputFrequencyAC) / 256.0

	s.Module2.ModuleId = fmt.Sprintf("%x", payload.Module2.Id)
	s.Module2.FirmwareVersion = fmt.Sprintf("%d/%d", payload.Module2.FirmwareVersionMajor, payload.Module2.FirmwareVersionMinor)
	s.Module2.InputVoltageDC = float64(payload.Module2.InputVoltageDC) / 512.0
	s.Module2.OutputPowerAC = float64(payload.Module2.OutputPowerAC) / 64.0
	s.Module2.TotalEnergy = float64(payload.Module2.TotalEnergy) / 8192.0
	s.Module2.Temperature = float64(payload.Module2.Temperature)/128.0 - 40.0
	s.Module2.OutputVoltageAC = float64(payload.Module2.OutputVoltageAC) / 64.0
	s.Module2.OutputFrequencyAC = float64(payload.Module2.OutputFrequencyAC) / 256.0

	return nil
}

package types

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
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
	s.Module1.InputVoltageDC = float64(payload.Module1.InputVoltageDC) * 64.0 / 32768.0
	s.Module1.OutputPowerAC = float64(payload.Module1.OutputPowerAC) * 512.0 / 32768.0
	s.Module1.TotalEnergy = float64(payload.Module1.TotalEnergy) * 4.0 / 32768.0
	s.Module1.Temperature = float64(payload.Module1.Temperature)*(256.0/32768.0) - 40.0
	s.Module1.OutputVoltageAC = float64(payload.Module1.OutputVoltageAC) * 512.0 / 32768.0
	s.Module1.OutputFrequencyAC = float64(payload.Module1.OutputFrequencyAC) * 128.0 / 32768.0

	s.Module2.ModuleId = fmt.Sprintf("%x", payload.Module2.Id)
	s.Module2.FirmwareVersion = fmt.Sprintf("%d/%d", payload.Module2.FirmwareVersionMajor, payload.Module2.FirmwareVersionMinor)
	s.Module2.InputVoltageDC = float64(payload.Module2.InputVoltageDC) * 64.0 / 32768.0
	s.Module2.OutputPowerAC = float64(payload.Module2.OutputPowerAC) * 512.0 / 32768.0
	s.Module2.TotalEnergy = float64(payload.Module2.TotalEnergy) * 4.0 / 32768.0
	s.Module2.Temperature = float64(payload.Module2.Temperature)*(256.0/32768.0) - 40.0
	s.Module2.OutputVoltageAC = float64(payload.Module2.OutputVoltageAC) * 512.0 / 32768.0
	s.Module2.OutputFrequencyAC = float64(payload.Module2.OutputFrequencyAC) * 128.0 / 32768.0

	return nil
}

func (i InverterStatus) String() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Inverter [%s] Status\n", i.InverterId))
	b.WriteString(fmt.Sprintf(">>>>        Total Energy: %fkWh\n", i.Module1.TotalEnergy+i.Module2.TotalEnergy))
	b.WriteString(fmt.Sprintf(">>>>       Current Power: %fW\n", i.Module1.OutputPowerAC+i.Module2.OutputPowerAC))
	b.WriteString(i.Module1.String())
	b.WriteString(i.Module2.String())

	return b.String()
}

func (m InverterModuleStatus) String() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("> Module [%s]\n", m.ModuleId))
	b.WriteString(fmt.Sprintf(">>>>    Firmware Version: %s\n", m.FirmwareVersion))
	b.WriteString(fmt.Sprintf(">>>>        Total Energy: %fkWh\n", m.TotalEnergy))
	b.WriteString(fmt.Sprintf(">>>>    Input Voltage DC: %fV\n", m.InputVoltageDC))
	b.WriteString(fmt.Sprintf(">>>>   Output Voltage AC: %fV\n", m.OutputVoltageAC))
	b.WriteString(fmt.Sprintf(">>>>     Output Power AC: %fW\n", m.OutputPowerAC))
	b.WriteString(fmt.Sprintf(">>>> Output Frequency AC: %fHz\n", m.OutputFrequencyAC))
	b.WriteString(fmt.Sprintf(">>>>         Temperature: %f°C\n", m.Temperature))

	return b.String()
}

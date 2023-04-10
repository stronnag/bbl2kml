package sitlgen

import "time"

type MspTX struct {
	m *MSPSerial
}

func NewMspTX(ms *MSPSerial) *MspTX {
	return &MspTX{m: ms}
}

func (t *MspTX) Send_TX(chans MSPChans, nchan int) time.Time {
	return t.m.send_tx(chans)
}

func (t *MspTX) Telem_reader() {
}

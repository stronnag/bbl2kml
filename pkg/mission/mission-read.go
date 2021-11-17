package mission

import (
	"fmt"
	"io/ioutil"
	"os"
)

func (mm *MultiMission) to_mission(mi int) *Mission {
	m := &Mission{}
	if mi > len(mm.Segment) {
		mi = len(mm.Segment)
	}
	mi--
	m.Version = mm.Version
	m.Comment = mm.Comment
	m.Metadata = mm.Segment[mi].Metadata
	m.MissionItems = mm.Segment[mi].MissionItems
	//	fmt.Fprintf(os.Stderr, "%#v\n", m)
	return m
}

func Read_Mission_File_Index(path string, idx int) (string, *Mission, error) {
	var dat []byte
	r, err := os.Open(path)
	if err == nil {
		defer r.Close()
		dat, err = ioutil.ReadAll(r)
	}
	if err != nil {
		return "?", nil, err
	} else {
		mtype, mm := handle_mission_data(dat, path)
		if mm == nil {
			fmt.Fprintf(os.Stderr, "Note: Mission fails verification %s\n", mtype)
			return mtype, nil, nil
		}
		m := mm.to_mission(idx)
		return mtype, m, nil
	}
}

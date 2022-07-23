package mission

import (
	"fmt"
	"io/ioutil"
	"os"
)

func (mm *MultiMission) To_mission(mi int) *Mission {
	m := &Mission{}
	if mi > len(mm.Segment) {
		mi = len(mm.Segment)
	}
	mi--
	if mi < 0 {
		mi = 0
	}
	m.Version = mm.Version
	m.Comment = mm.Comment
	m.Metadata = mm.Segment[mi].Metadata
	m.MissionItems = mm.Segment[mi].MissionItems
	//	fmt.Fprintf(os.Stderr, "%#v\n", m)
	return m
}

func Read_Mission_File(path string) (string, *MultiMission, error) {
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
		}
		return mtype, mm, nil
	}
}

func Read_Mission_File_Index(path string, idx int) (string, *Mission, error) {
	mtype, mm, err := Read_Mission_File(path)
	if err == nil {
		m := mm.To_mission(idx)
		return mtype, m, nil
	} else {
		return mtype, nil, err
	}
}

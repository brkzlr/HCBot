package main

import (
	"sync"
	"time"
)

var GlobalLock sync.Mutex

const (
	mccRoleID       = "985327507274350612"
	infiniteRoleID  = "985327648605614140"
	modernRoleID    = "985327740112760873"
	legacyRoleID    = "985327809566232626"
	lasochistRoleID = "985644631423320064"
	mccMasterRoleID = "985644874051231825"
	hcRoleID        = "985327939161849857"
	fcRoleID        = "985328007088590918"
)

type TimedRoles struct {
	ID   int
	Game string
}

var timedAchievRoles = map[int]TimedRoles{
	15: TimedRoles{990601630237982750, "Halo: Combat Evolved"},
	9:  TimedRoles{990601760961875998, "Halo 2 Classic"},
	25: TimedRoles{990601846391463967, "Halo 3"},
	6:  TimedRoles{990602246888779837, "Halo 4"},
	22: TimedRoles{990601924703297586, "Halo 3: ODST"},
	14: TimedRoles{990602198184501268, "Halo: Reach"},
}

type RoleDate struct {
	Day   int
	Month time.Month
}

var (
	destinationVacationDates = []RoleDate{
		RoleDate{1, time.January},
		RoleDate{7, time.July},
		RoleDate{31, time.October},
		RoleDate{25, time.December},
	}
	elderSignsDates = []RoleDate{
		RoleDate{1, time.January},
		RoleDate{22, time.April},
		RoleDate{5, time.May},
		RoleDate{4, time.July},
		RoleDate{7, time.July},
		RoleDate{31, time.October},
		RoleDate{11, time.November},
		RoleDate{25, time.December},
	}
)

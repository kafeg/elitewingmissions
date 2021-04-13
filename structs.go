package main

// -- dirs to search journals
func eliteDirs() [2]string {

	var eliteDirs [2]string
	eliteDirs[0] = "C:\\Users\\v3133\\Saved Games\\Frontier Developments\\Elite Dangerous"
	eliteDirs[1] = "Y:\\Elite Dangerous"

	//add here more dirs

	return eliteDirs
}

// --- consts
type UnstructuredJson map[string]interface{}
type HandlerFunction func (json UnstructuredJson)
const readIntervalInSecs = 3600 * 24 * 2 //last 3 days

// --- data structs
type TradeMission struct {
	MissionID float64
	Reward float64
	Commodity string
	Count float64
	CommanderName string
}

type PirateMission struct {
	MissionID float64
	Reward float64
	Faction string
	KillCount float64
	CommanderName string
	TargetFaction string
	Timestamp int64
}

type PFields struct {
	KillCount float64
	Reward float64
	CommanderName string
	Faction string
	MissionCount int
	AllRewards string
}

type MissionPackTimestamp struct {
	Start int64
	End int64
}
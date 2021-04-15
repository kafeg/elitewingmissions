package main

// -- dirs to search journals
func eliteDirs() [2]string {

	var eliteDirs [2]string
	eliteDirs[0] = "C:\\Users\\v3133\\Saved Games\\Frontier Developments\\Elite Dangerous"
	eliteDirs[1] = "\\\\192.168.1.183\\vesta_elite_saves\\Elite Dangerous"

	//add here more dirs

	return eliteDirs
}

// --- consts
type UnstructuredJson map[string]interface{}
type HandlerFunction func (json UnstructuredJson)
const readMissionsIntervalInSecs = 3600 * 24 * 3 //last 2 days
const readTransientDataIntervalInSecs = 3600 * 1 //last 1 hour
const onePirateTimeInSecs = 90 //seconds per one pirate, this is just for predict the future

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
	KillCount int64
	CommanderName string
	TargetFaction string
	Timestamp int64
}

type PFields struct {
	KillCount int64
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
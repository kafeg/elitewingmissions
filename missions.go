package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// --- data storages, active mission struct indexed by the mission ids
var activeTradeMissions = make(map[float64]*TradeMission)
var activePirateMissions = make(map[float64]PirateMission)

var currentCommanderName string
var bountiesCounts = make(map[string]int64)

var cmdrsCreditsCount = make(map[string]int64)
var cmdrsCarriersCreditsCount = make(map[string]int64)

// -- caches
var bountiesTimestamps = make(map[string]MissionPackTimestamp)
var victimFactions []string // append works on nil slices.

// --- event handlers

// When Written: when starting a mission
func hMissionAccepted(json UnstructuredJson) {
	missionId := json["MissionID"].(float64)

	// check Trade Wing missions
	if _, ok := activeTradeMissions[missionId]; !ok {
		if json["Commodity"] != nil && json["Reward"] != nil {

			commodityName := json["Commodity"].(string)
			commodityName = strings.ReplaceAll(commodityName, "$", "")
			commodityName = strings.ReplaceAll(commodityName, "_Name;", "")

			activeTradeMissions[missionId] = &TradeMission{
				missionId,
				json["Reward"].(float64),
				commodityName,
				json["Count"].(float64),
				currentCommanderName,
			}
		}
		//fmt.Printf("MissionAccepted, %v\n", missionId)
	}

	// check PIRATE Wing missions
	if _, ok := activePirateMissions[missionId]; !ok {

		if json["KillCount"] != nil && json["Faction"] != nil  && json["TargetFaction"] != nil {

			faction := json["Faction"].(string)

			//layout := "2021-04-08T12:08:50Z"
			timestamp, _ := time.Parse(time.RFC3339 /*layout*/, json["timestamp"].(string))

			activePirateMissions[missionId] = PirateMission{
				missionId,
				json["Reward"].(float64),
				faction,
				int64(json["KillCount"].(float64)),
				currentCommanderName,
				json["TargetFaction"].(string),
				timestamp.Unix(),
			}
		}
		//fmt.Printf("MissionAccepted, %v\n", json)
	}

	//todo other mission handlers ...
}

// When Written: when a mission is completed
func hMissionCompleted(json UnstructuredJson) {
	missionId := json["MissionID"].(float64)
	if _, ok := activeTradeMissions[missionId]; ok {
		delete(activeTradeMissions, missionId)
	}

	if _, ok := activePirateMissions[missionId]; ok {
		delete(activePirateMissions, missionId)
	}
}

// When Written: when a mission has been abandoned
func hMissionAbandoned(json UnstructuredJson) {
	missionId := json["MissionID"].(float64)
	if _, ok := activeTradeMissions[missionId]; ok {
		delete(activeTradeMissions, missionId)
	}

	if _, ok := activePirateMissions[missionId]; ok {
		delete(activePirateMissions, missionId)
	}
}

// When Written: when a mission has failed
func hMissionFailed(json UnstructuredJson) {
	missionId := json["MissionID"].(float64)
	if _, ok := activeTradeMissions[missionId]; ok {
		delete(activeTradeMissions, missionId)
	}

	if _, ok := activePirateMissions[missionId]; ok {
		delete(activePirateMissions, missionId)
	}
}

// When written: when collecting or delivering cargo for a wing mission, or if a wing member updates progress
func hCargoDepot(json UnstructuredJson) {
	missionId := json["MissionID"].(float64)
	if _, ok := activeTradeMissions[missionId]; ok {

		//fmt.Println(json)

		if val, ok := json["UpdateType"]; ok && val == "WingUpdate" {
			activeTradeMissions[missionId].Count = json["TotalItemsToDeliver"].(float64) - json["ItemsDelivered"].(float64)
		}

		if val, ok := json["UpdateType"]; ok && val == "Deliver" {
			//fmt.Println(json)
			activeTradeMissions[missionId].Count -= json["Count"].(float64)
		}

		//if json["UpdateType"].(string) == "WingUpdate" {
		//activeTradeMissions[missionId].Count += json["ItemsCollected"].(float64)
		//activeTradeMissions[missionId].Count -= json["ItemsDelivered"].(float64)
		//}
	}
}

func hCommander(json UnstructuredJson) {
	currentCommanderName = json["Name"].(string)
	if _, ok := bountiesCounts[currentCommanderName]; !ok {
		bountiesCounts[currentCommanderName] = 0 // initialize value
	}
}

func hLoadGame(json UnstructuredJson) {
	cmdrsCreditsCount[currentCommanderName] = int64(json["Credits"].(float64)) // always update
}

func hCarrierStats(json UnstructuredJson) {
	cmdrsCarriersCreditsCount[currentCommanderName] = int64(json["Finance"].(map[string]interface {})["CarrierBalance"].(float64)) // always update
}

func hBounty(json UnstructuredJson) {
	timestamp, _ := time.Parse(time.RFC3339 /*layout*/, json["timestamp"].(string))
	if (timestamp.Unix() > bountiesTimestamps[currentCommanderName].End) {
		//fmt.Println(timestamp.Unix(), json["VictimFaction"].(string))

		contains := false
		for _, val := range victimFactions {
			if strings.Contains(json["VictimFaction"].(string), val) {
				contains = true
				break
			}
		}

		if contains {
			bountiesCounts[currentCommanderName]++
		}
	}
}

func calcTradeMissions() {
	//print all active trade missions
	fmt.Println("Active Trade Wing missions")
	for _, v := range activeTradeMissions {
		fmt.Printf("%s, %v, %v\n", v.Commodity, v.Count, strconv.FormatFloat(v.Reward, 'f', -1, 64))
	}
	fmt.Println("")

	//calc active trade wing missions demand
	fmt.Println("Total Trade Wing Demand")
	totalactiveTradeMissionsDemand := make(map[string]float64)
	for _, v := range activeTradeMissions {
		if _, ok := totalactiveTradeMissionsDemand[v.Commodity]; ok {
			// append
			totalactiveTradeMissionsDemand[v.Commodity] = totalactiveTradeMissionsDemand[v.Commodity] + v.Count
		} else {
			// just insert
			totalactiveTradeMissionsDemand[v.Commodity] = v.Count
		}
	}

	for k, v := range totalactiveTradeMissionsDemand {
		fmt.Printf("%s = %v\n", k, v)
	}
}

func calcPirateMissions() {
	//calc active pirate wing missions demand
	fmt.Println("")
	fmt.Println("--- --- --- --- --- --- --- --- --- --- --- ---")
	fmt.Println("")

	fmt.Println("Total Pirate KillCount Demand.")
	fmt.Println("")

	overAllRemains := make(map[string]int64)
	overAllCompleted := make(map[string]int64)
	overAllMaxKillCount := make(map[string]int64)
	overallTotalMissions := 0
	overallTotalRewardX4 := 0.0

	cmdrs := getCmdrsList(activePirateMissions)
	for _, cmdr := range cmdrs {
		fmt.Println("--- CMDR", cmdr)
		totalPirateactiveTradeMissionsDemand := make(map[string]PFields)
		for _, v := range activePirateMissions {

			if cmdr != v.CommanderName {
				continue
			}

			if _, ok := totalPirateactiveTradeMissionsDemand[v.CommanderName+"_"+v.Faction]; ok {
				// append
				pfield := totalPirateactiveTradeMissionsDemand[v.CommanderName+"_"+v.Faction]
				pfield.KillCount = pfield.KillCount + v.KillCount
				pfield.Reward = pfield.Reward + v.Reward
				pfield.MissionCount++
				pfield.AllRewards = pfield.AllRewards + " / " + FormatNumber(v.Reward)[:4] + " (" + strconv.FormatInt(v.KillCount, 10) + ")"
				totalPirateactiveTradeMissionsDemand[v.CommanderName+"_"+v.Faction] = pfield
			} else {
				// just insert
				totalPirateactiveTradeMissionsDemand[v.CommanderName+"_"+v.Faction] = PFields{
					v.KillCount,
					v.Reward,
					v.CommanderName,
					v.Faction,
					1,
					FormatNumber(v.Reward)[:4] + " (" + strconv.FormatInt(v.KillCount, 10) + ")",
				}
			}
		}

		//sort keys by value
		keys := make([]string, 0, len(totalPirateactiveTradeMissionsDemand))
		for key := range totalPirateactiveTradeMissionsDemand {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool { return totalPirateactiveTradeMissionsDemand[keys[i]].KillCount < totalPirateactiveTradeMissionsDemand[keys[j]].KillCount })

		fmt.Printf("%34s, %4s, %4s, %15s, %69s\n", "Faction", "Kill", "Mssn", "Total", "Per Mission: millions (tgt cnt)")

		totalReward := 0.0
		var totalKillCount int64 = 0
		totalMissions := 0
		var maxKillCount int64 = 0
		for _, key := range keys {
			v := totalPirateactiveTradeMissionsDemand[key]
			fmt.Printf("%34s, %4v, %4v, %15s, %69s\n", v.Faction, v.KillCount, v.MissionCount, FormatNumber(v.Reward), v.AllRewards)
			totalReward += v.Reward
			totalKillCount += v.KillCount
			totalMissions += v.MissionCount

			if int64(v.KillCount) > maxKillCount {
				maxKillCount = int64(v.KillCount)
			}
		}
		fmt.Printf("%34s, %4v  %4v, %15s\n", "Total", "", totalMissions, FormatNumber(totalReward))
		fmt.Printf("%34s, %4s  %4s  %15s\n", "Total*4", "", "", FormatNumber(totalReward*4))
		fmt.Printf("%34s, %4v  %4s  %15s\n", "Done", bountiesCounts[cmdr], "", "")
		fmt.Printf("%34s, %4v  %4s  %15s\n", "Remnain", maxKillCount-int64(bountiesCounts[cmdr]), "", "")
		fmt.Println("")
		overAllRemains[cmdr] = maxKillCount - int64(bountiesCounts[cmdr])
		overAllCompleted[cmdr] = bountiesCounts[cmdr]
		overAllMaxKillCount[cmdr] = maxKillCount
		overallTotalMissions = overallTotalMissions + totalMissions
		overallTotalRewardX4 = overallTotalRewardX4 + totalReward*4
	}

	fmt.Println("--- Total over all")
	fmt.Printf("%34s, %4s, %4s, %6s, %32s, %5s\n", "CMDR", "Max", "Done", "Remain", "Missions collecting time", "Aprox")
	for _, cmdr := range cmdrs {
		t1 := time.Unix(bountiesTimestamps[cmdr].Start, 0)
		t1f := fmt.Sprintf("%02d-%02d %02d:%02d", t1.Month(), t1.Day(), t1.Hour(), t1.Minute())

		t2 := time.Unix(bountiesTimestamps[cmdr].End, 0)
		t2f := fmt.Sprintf("%02d-%02d %02d:%02d", t2.Month(), t2.Day(), t2.Hour(), t2.Minute())

		collectingTime := strconv.FormatFloat(t2.Sub(t1).Minutes(), 'f', 0, 64)
		remainingTime := float64(overAllRemains[cmdr]) * onePirateTimeInSecs / 60
		if remainingTime < 0 {
			remainingTime = 0
		}
		remainingTimeF :=  strconv.FormatFloat(remainingTime, 'f', 0, 64)
		fmt.Printf("%34s, %4v, %4v, %6v, %11s -> %11s, %3vm, %4vm\n", cmdr, overAllMaxKillCount[cmdr], overAllCompleted[cmdr], overAllRemains[cmdr], t1f, t2f, collectingTime, remainingTimeF)
	}
	fmt.Println("")
	fmt.Printf("%34s, %4v, %6v\n", "Total missn/reward", overallTotalMissions, FormatNumber(overallTotalRewardX4))
	fmt.Println("")
}

func recalcStats() {
	cmdrs := getCmdrsList(activePirateMissions)

	var totalCmdrs int64 = 0
	var totalCarriers int64 = 0
	for _, cmdr := range cmdrs {
		fmt.Printf("%34s, %15v, %15v\n", cmdr, FormatNumberInt(cmdrsCreditsCount[cmdr]), FormatNumberInt(cmdrsCarriersCreditsCount[cmdr]))
		totalCmdrs += cmdrsCreditsCount[cmdr]
		totalCarriers += cmdrsCarriersCreditsCount[cmdr]
	}
	fmt.Printf("%34s, %15v, %15v\n", "Total", FormatNumberInt(totalCmdrs), FormatNumberInt(totalCarriers))
	fmt.Printf("%34s, %15v\n", "Overall Total", FormatNumberInt(totalCmdrs + totalCarriers))
}

func recalcAllMissions() {

	//handlers https://elite-journal.readthedocs.io
	handlers := map[string] HandlerFunction {
		"MissionAccepted": hMissionAccepted,
		"MissionCompleted": hMissionCompleted,
		"MissionAbandoned": hMissionAbandoned,
		"MissionFailed": hMissionFailed,
		"CargoDepot": hCargoDepot,
		"Commander": hCommander,
		"LoadGame": hLoadGame,
		"CarrierStats": hCarrierStats,
	}

	handleEvents(handlers, false)

	retrieveBountyTimestamps(activePirateMissions)
	retrieveVictimFactions(activePirateMissions)

	handlers = map[string]HandlerFunction{
		"Commander": hCommander,
		"Bounty":    hBounty,
	}

	handleEvents(handlers, false)

	//calcTradeMissions()

	calcPirateMissions()

	recalcStats()
}

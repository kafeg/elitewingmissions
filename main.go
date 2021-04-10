package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// --- consts
const eliteSavesDir = "C:\\Users\\v3133\\Saved Games\\Frontier Developments\\Elite Dangerous"
type UnstructuredJson map[string]interface{}
type HandlerFunction func (json UnstructuredJson)


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

// --- data storages
var activeWingMissions = make(map[float64]*TradeMission)       // active mission struct indexed by the mission ids
var activePirateWingMissions = make(map[float64]PirateMission) // active mission struct indexed by the mission ids
var pirateBountiesCount = 0

var victimFactions []string // append works on nil slices.
var currentCommanderName string
var currentEarlierTimestampForBounties = time.Now().UTC().Unix()

// parser
func handleEvents(handlers map[string] HandlerFunction) {
	//parse each file and call handler for row if it exists
	items, _ := ioutil.ReadDir(eliteSavesDir)
	for _, item := range items {
		if strings.HasPrefix(item.Name(), "Journal") && strings.HasSuffix(item.Name(), ".log") {
			//fmt.Println(item.Name())

			inFile, _ := os.Open(eliteSavesDir + "\\" + item.Name())
			defer inFile.Close()
			scanner := bufio.NewScanner(inFile)
			scanner.Split(bufio.ScanLines)

			for scanner.Scan() {

				// optimize to prevent parse each event json
				contains := false
				for k, _ := range handlers {
					if strings.Contains(scanner.Text(), k) {
						contains = true
						break
					}
				}

				if !contains {
					continue
				}
				// end of optimize

				var result map[string]interface{}
				json.Unmarshal([]byte(scanner.Text()), &result)
				eventType := result["event"].(string)
				if _, ok := handlers[eventType]; ok {
					handlers[eventType](result)
					//fmt.Println(eventType)
				}
			}
		}
	}
}

// --- event handlers

// When Written: when starting a mission
func hMissionAccepted(json UnstructuredJson) {
	missionId := json["MissionID"].(float64)

	// check Trade Wing missions
	if _, ok := activeWingMissions[missionId]; !ok {
		if json["Commodity"] != nil && json["Reward"] != nil {

			commodityName := json["Commodity"].(string)
			commodityName = strings.ReplaceAll(commodityName, "$", "")
			commodityName = strings.ReplaceAll(commodityName, "_Name;", "")

			activeWingMissions[missionId] = &TradeMission{
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
	if _, ok := activePirateWingMissions[missionId]; !ok {

		if json["KillCount"] != nil && json["Faction"] != nil  && json["TargetFaction"] != nil {

			faction := json["Faction"].(string)

			//layout := "2021-04-08T12:08:50Z"
			timestamp, _ := time.Parse(time.RFC3339 /*layout*/, json["timestamp"].(string))
			currentEarlierTimestampForBounties = timestamp.UTC().Unix()

			//if err != nil {
			//	fmt.Println("ERR", err)
			//}
			//fmt.Println(json["timestamp"].(string), timestamp, timestamp.Unix())

			activePirateWingMissions[missionId] = PirateMission{
				missionId,
				json["Reward"].(float64),
				faction,
				json["KillCount"].(float64),
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
	if _, ok := activeWingMissions[missionId]; ok {
		delete(activeWingMissions, missionId)
	}

	if _, ok := activePirateWingMissions[missionId]; ok {
		delete(activePirateWingMissions, missionId)
	}
}

// When Written: when a mission has been abandoned
func hMissionAbandoned(json UnstructuredJson) {
	missionId := json["MissionID"].(float64)
	if _, ok := activeWingMissions[missionId]; ok {
		delete(activeWingMissions, missionId)
	}

	if _, ok := activePirateWingMissions[missionId]; ok {
		delete(activePirateWingMissions, missionId)
	}
}

// When Written: when a mission has failed
func hMissionFailed(json UnstructuredJson) {
	missionId := json["MissionID"].(float64)
	if _, ok := activeWingMissions[missionId]; ok {
		delete(activeWingMissions, missionId)
	}

	if _, ok := activePirateWingMissions[missionId]; ok {
		delete(activePirateWingMissions, missionId)
	}
}

// When written: when collecting or delivering cargo for a wing mission, or if a wing member updates progress
func hCargoDepot(json UnstructuredJson) {
	missionId := json["MissionID"].(float64)
	if _, ok := activeWingMissions[missionId]; ok {

		//fmt.Println(json)

		if val, ok := json["UpdateType"]; ok && val == "WingUpdate" {
			activeWingMissions[missionId].Count = json["TotalItemsToDeliver"].(float64) - json["ItemsDelivered"].(float64)
		}

		if val, ok := json["UpdateType"]; ok && val == "Deliver" {
			//fmt.Println(json)
			activeWingMissions[missionId].Count -= json["Count"].(float64)
		}

		//if json["UpdateType"].(string) == "WingUpdate" {
			//activeWingMissions[missionId].Count += json["ItemsCollected"].(float64)
			//activeWingMissions[missionId].Count -= json["ItemsDelivered"].(float64)
		//}
	}
}

func hCommander(json UnstructuredJson) {
	currentCommanderName = json["Name"].(string)
}

func hBounty(json UnstructuredJson) {
	timestamp, _ := time.Parse(time.RFC3339 /*layout*/, json["timestamp"].(string))
	if (timestamp.Unix() > currentEarlierTimestampForBounties) {
		//fmt.Println(timestamp.Unix(), json["VictimFaction"].(string))

		contains := false
		for _, val := range victimFactions {
			if strings.Contains(json["VictimFaction"].(string), val) {
				contains = true
				break
			}
		}

		if contains {
			pirateBountiesCount++
		}
	}
}

func calcTradeMissions() {
	//print all active trade missions
	fmt.Println("Active Trade Wing missions")
	for _, v := range activeWingMissions {
		fmt.Printf("%s, %v, %v\n", v.Commodity, v.Count, strconv.FormatFloat(v.Reward, 'f', -1, 64))
	}
	fmt.Println("")

	//calc active trade wing missions demand
	fmt.Println("Total Trade Wing Demand")
	totalActiveWingMissionsDemand := make(map[string]float64)
	for _, v := range activeWingMissions {
		if _, ok := totalActiveWingMissionsDemand[v.Commodity]; ok {
			// append
			totalActiveWingMissionsDemand[v.Commodity] = totalActiveWingMissionsDemand[v.Commodity] + v.Count
		} else {
			// just insert
			totalActiveWingMissionsDemand[v.Commodity] = v.Count
		}
	}

	for k, v := range totalActiveWingMissionsDemand {
		fmt.Printf("%s = %v\n", k, v)
	}
}

func calcPirateMissions() {
	//calc active pirate wing missions demand
	fmt.Println("")
	fmt.Println("---")
	fmt.Println("Total Pirate KillCount Demand")
	type PFields struct {
		KillCount float64
		Reward float64
		CommanderName string
		MissionCount int
	}
	totalPirateActiveWingMissionsDemand := make(map[string]PFields)
	for _, v := range activePirateWingMissions {
		if _, ok := totalPirateActiveWingMissionsDemand[v.Faction]; ok {
			// append
			pfield := totalPirateActiveWingMissionsDemand[v.Faction]
			pfield.KillCount = pfield.KillCount + v.KillCount
			pfield.Reward = pfield.Reward + v.Reward
			pfield.MissionCount++
			totalPirateActiveWingMissionsDemand[v.Faction] = pfield
		} else {
			// just insert
			totalPirateActiveWingMissionsDemand[v.Faction] = PFields{v.KillCount, v.Reward, v.CommanderName, 1}
		}
	}

	//sort keys by value
	keys := make([]string, 0, len(totalPirateActiveWingMissionsDemand))
	for key := range totalPirateActiveWingMissionsDemand {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return totalPirateActiveWingMissionsDemand[keys[i]].KillCount < totalPirateActiveWingMissionsDemand[keys[j]].KillCount })

	totalReward := 0.0
	totalKillCount := 0.0
	totalMissions := 0
	for _, key := range keys {
		v := totalPirateActiveWingMissionsDemand[key]
		fmt.Printf("%13s, %33s, %4v, %2v, %v\n", v.CommanderName, key, v.KillCount, v.MissionCount, strconv.FormatFloat(v.Reward, 'f', -1, 64))
		totalReward += v.Reward
		totalKillCount += v.KillCount
		totalMissions += v.MissionCount
	}
	fmt.Printf("%13s, %33s, %4s, %2v, %v\n", "Total", "", "", totalMissions, strconv.FormatFloat(totalReward * 4, 'f', -1, 64))
	fmt.Printf("%13s, %33s, %4v, %2s, %s\n", "Bounties done", "", pirateBountiesCount, "", "")
}

func main() {

	//handlers https://elite-journal.readthedocs.io
	handlers := map[string] HandlerFunction {
		"MissionAccepted": hMissionAccepted,
		"MissionCompleted": hMissionCompleted,
		"MissionAbandoned": hMissionAbandoned,
		"MissionFailed": hMissionFailed,
		"CargoDepot": hCargoDepot,
		"Commander": hCommander,
	}

	handleEvents(handlers)

	//get earlier timestamp for active missions
	for _, v := range activePirateWingMissions {
		if v.Timestamp < currentEarlierTimestampForBounties {
			currentEarlierTimestampForBounties = v.Timestamp
		}
	}
    //fmt.Println("Earlier Mission:", time.Unix(currentEarlierTimestampForBounties,0))

	//calc victim factions
	for _, v := range activePirateWingMissions {
		contains := false
		for _, val := range victimFactions {
			if strings.Contains(v.TargetFaction, val) {
				contains = true
				break
			}
		}

		if !contains {
			victimFactions = append(victimFactions, v.TargetFaction)
		}
	}
	//fmt.Println("victimFactions", victimFactions)

	handlers = map[string] HandlerFunction {
		"Bounty": hBounty,
	}

	handleEvents(handlers)

	calcTradeMissions()

	calcPirateMissions()
}

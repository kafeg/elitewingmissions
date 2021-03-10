package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"io/ioutil"
	"strings"
)

// --- consts
const eliteSavesDir = "C:\\Users\\v3133\\Saved Games\\Frontier Developments\\Elite Dangerous"
type UnstructuredJson map[string]interface{}
type HandlerFunction func (json UnstructuredJson)


// --- data structs
type ActiveMissions struct {
	MissionID float64
	Reward float64
	Commodity string
	Count float64
}

// --- data storages
var activeWingMissions = make(map[float64]ActiveMissions) // active mission stricts indexed by the mission ids

// --- event handlers
func hMissionAccepted(json UnstructuredJson) {
	missionId := json["MissionID"].(float64)

	// check Trade Wing missions
	if _, ok := activeWingMissions[missionId]; !ok {
		if json["Commodity"] != nil && json["Reward"] != nil {

			commodityName := json["Commodity"].(string)
			commodityName = strings.ReplaceAll(commodityName, "$", "")
			commodityName = strings.ReplaceAll(commodityName, "_Name;", "")

			activeWingMissions[missionId] = ActiveMissions{
				missionId,
				json["Reward"].(float64),
				commodityName,
				json["Count"].(float64),
			}
		}
		//fmt.Printf("MissionAccepted, %v\n", missionId)
	}

	//todo other missions ...
}

func hMissionCompleted(json UnstructuredJson) {
	missionId := json["MissionID"].(float64)
	if _, ok := activeWingMissions[missionId]; ok {
		delete(activeWingMissions, missionId)
	}
}

func hMissionAbandoned(json UnstructuredJson) {
	missionId := json["MissionID"].(float64)
	if _, ok := activeWingMissions[missionId]; ok {
		delete(activeWingMissions, missionId)
	}
}

func hMissionFailed(json UnstructuredJson) {
	missionId := json["MissionID"].(float64)
	if _, ok := activeWingMissions[missionId]; ok {
		delete(activeWingMissions, missionId)
	}
}

func main() {
	//fmt.Println("Starting...")

	//handlers
	handlers := map[string] HandlerFunction {
		"MissionAccepted": hMissionAccepted,
		"MissionCompleted": hMissionCompleted,
		"MissionAbandoned": hMissionAbandoned,
		"MissionFailed": hMissionFailed,
	}

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
				if strings.Contains(scanner.Text(), "\"event\":\"Mission") {
					//fmt.Println(scanner.Text())

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

	//print all active missions
	//for k, v := range activeWingMissions {
	//	fmt.Printf("key[%s] value[%s]\n", k, v)
	//}

	//calc active wing missions demand
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
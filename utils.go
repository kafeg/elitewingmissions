package main

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

func retrieveVictimFactions(pirateMissions map[float64]PirateMission) []string {

	if len(victimFactions) > 0 {
		return victimFactions
	}

	//calc victim factions
	for _, v := range activePirateMissions {
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
	return victimFactions
}

func getCmdrsList(pirateMissions map[float64]PirateMission) []string {
	//add uniq cmdr to list
	var cmdrs []string
	for _, v := range pirateMissions {
		//fmt.Println(v.CommanderName, v.Faction, v.Reward, v.KillCount)

		contains := false
		for _, val := range cmdrs {
			if strings.Contains(v.CommanderName, val) {
				contains = true
				break
			}
		}

		if !contains {
			cmdrs = append(cmdrs, v.CommanderName)
		}
	}

	sort.Strings(cmdrs)
	return cmdrs
}

func retrieveBountyTimestamps(pirateMissions  map[float64]PirateMission) map[string]MissionPackTimestamp {

	//get timestamp for active missions
	cmdrs := getCmdrsList(pirateMissions)
	for _, cmdr := range cmdrs {
		bountiesTimestamps[cmdr] = MissionPackTimestamp {time.Now().Unix(), 0}
	}

	for _, v := range pirateMissions {
		ts := bountiesTimestamps[v.CommanderName]
		if v.Timestamp < bountiesTimestamps[v.CommanderName].Start {
			ts.Start = v.Timestamp
		}
		if v.Timestamp > bountiesTimestamps[v.CommanderName].End {
			ts.End = v.Timestamp
		}
		bountiesTimestamps[v.CommanderName] = ts
	}

	return bountiesTimestamps
}

func FormatNumber(n float64) string {
    return FormatNumberInt(int64(n))
}

func FormatNumberInt(n int64) string {
	in := strconv.FormatInt(n, 10)
	numOfDigits := len(in)
	if n < 0 {
		numOfDigits-- // First character is the - sign (not a digit)
	}
	numOfCommas := (numOfDigits - 1) / 3

	out := make([]byte, len(in)+numOfCommas)
	if n < 0 {
		in, out[0] = in[1:], '-'
	}

	for i, j, k := len(in)-1, len(out)-1, 0; ; i, j = i-1, j-1 {
		out[j] = in[i]
		if i == 0 {
			return string(out)
		}
		if k++; k == 3 {
			j, k = j-1, 0
			out[j] = ','
		}
	}
}

// parser
func handleEvents(handlers map[string] HandlerFunction, transient bool) {

	maxModTime := time.Now().Unix() - readMissionsIntervalInSecs
	if transient {
		maxModTime = time.Now().Unix() - readTransientDataIntervalInSecs
	}

	for _, dir := range eliteDirs() {

		//parse each file and call handler for row if it exists
		items, _ := ioutil.ReadDir(dir)
		for _, item := range items {
			if item.ModTime().Unix() > maxModTime && strings.HasPrefix(item.Name(), "Journal") && strings.HasSuffix(item.Name(), ".log") {
				//fmt.Println(item.Name())

				inFile, _ := os.Open(dir + "\\" + item.Name())
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
}
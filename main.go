package main

import (
	"encoding/json"
	"github.com/runi95/wts-parser/models"
	"github.com/runi95/wts-parser/parser"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var baseUnitMap map[string]*models.SLKUnit
var unitFuncMap map[string]*models.UnitFunc
var builders []*UnitRaw

type Unit struct {
	Name        string
	Icon        string
	Description string
	Builds      []*Unit
	Upgrades    []*Unit
}

type UnitRaw struct {
	SLKUnit  *models.SLKUnit
	UnitFunc *models.UnitFunc
	Upgrades []*UnitRaw
	Builds   []*UnitRaw
}

func main() {
	if len(os.Args) == 3 {
		inputFolder := os.Args[1]
		outputFolder := os.Args[2]
		loadSLK(inputFolder)
		findBuilders()
		writeToJson(outputFolder)
	} else {
		log.Printf("Expected 2 arguments (input, output), but got %d\n", len(os.Args)-1)
	}
}

func writeToJson(outputFolder string) {
	bytes, err := json.MarshalIndent(builders, "", " ")
	if err != nil {
		log.Println(err)
	}

	err = ioutil.WriteFile(outputFolder, bytes, 0644)
	if err != nil {
		log.Println(err)
	}
}

func buildUnit(unitId string) *Unit {
	unitFuncBuild := unitFuncMap[unitId]
	var upgrades []*Unit
	if unitFuncBuild.Upgrade.Valid {
		unitUpgrades := unitFuncBuild.Upgrade.String
		upgradeSplit := strings.Split(unitUpgrades, ",")
		for _, unitUpgrade := range upgradeSplit {
			upgrades = append(upgrades, buildUnit(unitUpgrade))
		}
	}

	return &Unit{unitFuncBuild.Name.String, unitFuncBuild.Art.String, unitFuncBuild.Description.String, nil, upgrades}
}

func buildRawUnit(unitId string) *UnitRaw {
	unitFuncBuild := unitFuncMap[unitId]
	baseUnitBuild := baseUnitMap[unitId]
	var upgrades []*UnitRaw
	if unitFuncBuild.Upgrade.Valid {
		unitUpgrades := unitFuncBuild.Upgrade.String
		upgradeSplit := strings.Split(unitUpgrades, ",")
		for _, unitUpgrade := range upgradeSplit {
			if len(unitUpgrade) > 0 {
				upgrades = append(upgrades, buildRawUnit(unitUpgrade))
			}
		}
	}

	return &UnitRaw{baseUnitBuild, unitFuncBuild, upgrades, nil}
}

func findBuilders() {
	for key := range baseUnitMap {
		// log.Println(baseUnitMap[key].UnitBalance.Type.String)
		unitTypes := baseUnitMap[key].UnitBalance.Type.String
		split := strings.Split(strings.Trim(unitTypes, "\""), ",")
		for _, unitType := range split {
			if strings.ToLower(unitType) == "peon" {
				unitFunc := unitFuncMap[key]
				slkUnit := baseUnitMap[key]
				var builds []*UnitRaw
				unitBuilds := unitFunc.Builds.String
				buildSplit := strings.Split(strings.Trim(unitBuilds, "\""), ",")
				for _, unitBuild := range buildSplit {
					if len(unitBuild) > 0 {
						builds = append(builds, buildRawUnit(unitBuild))
					}
				}

				builders = append(builders, &UnitRaw{slkUnit, unitFunc, nil, builds})
			}
		}
	}
}

func loadSLK(inputFolder string) {
	log.Println("Reading UnitAbilities.slk...")

	unitAbilitiesBytes, err := ioutil.ReadFile(filepath.Join(inputFolder, "UnitAbilities.slk"))
	if err != nil {
		log.Println(err)
		os.Exit(10)
	}

	unitAbilitiesMap := parser.SlkToUnitAbilities(unitAbilitiesBytes)

	log.Println("Reading UnitData.slk...")

	unitDataBytes, err := ioutil.ReadFile(filepath.Join(inputFolder, "UnitData.slk"))
	if err != nil {
		log.Println(err)
		os.Exit(10)
	}

	unitDataMap := parser.SlkToUnitData(unitDataBytes)

	log.Println("Reading UnitUI.slk...")

	unitUIBytes, err := ioutil.ReadFile(filepath.Join(inputFolder, "UnitUI.slk"))
	if err != nil {
		log.Println(err)
		os.Exit(10)
	}

	unitUIMap := parser.SLKToUnitUI(unitUIBytes)

	log.Println("Reading UnitWeapons.slk...")

	unitWeaponsBytes, err := ioutil.ReadFile(filepath.Join(inputFolder, "UnitWeapons.slk"))
	if err != nil {
		log.Println(err)
		os.Exit(10)
	}

	unitWeaponsMap := parser.SLKToUnitWeapons(unitWeaponsBytes)

	log.Println("Reading UnitBalance.slk...")

	unitBalanceBytes, err := ioutil.ReadFile(filepath.Join(inputFolder, "UnitBalance.slk"))
	if err != nil {
		log.Println(err)
		os.Exit(10)
	}

	unitBalanceMap := parser.SLKToUnitBalance(unitBalanceBytes)

	log.Println("Reading CampaignUnitFunc.txt...")

	campaignUnitFuncBytes, err := ioutil.ReadFile(filepath.Join(inputFolder, "CampaignUnitFunc.txt"))
	if err != nil {
		log.Println(err)
		os.Exit(10)
	}

	unitFuncMap = parser.TxtToUnitFunc(campaignUnitFuncBytes)

	baseUnitMap = make(map[string]*models.SLKUnit)
	i := 0
	for k := range unitDataMap {
		slkUnit := new(models.SLKUnit)
		slkUnit.UnitAbilities = unitAbilitiesMap[k]
		slkUnit.UnitData = unitDataMap[k]
		slkUnit.UnitUI = unitUIMap[k]
		slkUnit.UnitWeapons = unitWeaponsMap[k]
		slkUnit.UnitBalance = unitBalanceMap[k]

		unitId := strings.Replace(k, "\"", "", -1)
		baseUnitMap[unitId] = slkUnit
		i++
	}
}

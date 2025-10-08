package scan

import (
	"log"

	"src/pkg/zapScanner"
)

const scanPassedTreshold = 8

// This maps CWE code to score to decide whether to block
// Will be in DB
var vulnerabilityScore = map[string]int{
	"89":  8, // SQL injection
	"527": 8, // .git or svn source code exposure
	"98":  8, // Local File Include
	"548": 4, // Directory Listing
}

func CheckScan(r zapScanner.AScanResult) (bool, error) {
	score := 0
	for _, s := range r.Sites {
		log.Printf("[Scan results] Site: %+v\n", s)
		for _, a := range s.Alerts {
			log.Printf("[Finding] %+v\n", a)
			score += vulnerabilityScore[a.Cweid]
		}
	}
	return score < scanPassedTreshold, nil
}

package main

import (
	"flag"
	"io/ioutil"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var path string
var fanTemp int
var timeout int

func init() {
	flag.StringVar(&path, "path", "/sys/class/thermal/thermal_zone*", "path to sensors")
	flag.IntVar(&fanTemp, "fan-temp", 75000, "temp to turn on fan")
	flag.IntVar(&timeout, "timeout", 60, "temp reading timeout")
}

func main() {
	flag.Parse()

	matches, err := filepath.Glob(path)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("/sys/class/hwmon/hwmon0/automatic", []byte("0"), 0666)
	if err != nil {
		log.Fatalf("failed to disable automatic: %v", err)
	}
	err = ioutil.WriteFile("/sys/class/hwmon/hwmon0/pwm1", []byte("0"), 0666)
	if err != nil {
		log.Fatalf("failed to write temp value: %v", err)
	}

	var turnedOn bool

	for {
		for _, m := range matches {
			fullpath := filepath.Join(m, "temp")

			b, err := ioutil.ReadFile(fullpath)
			if err != nil {
				log.Fatalf("failed to read temp from file: %v", err)
			}
			s := string(b)
			s = strings.TrimSpace(s)

			tm, err := strconv.Atoi(s)
			if err != nil {
				log.Fatalf("failed to read temp values as int: %v", err)
			}
			log.Printf("temps: %d", tm)

			if tm > fanTemp && !turnedOn {
				err := ioutil.WriteFile("/sys/class/hwmon/hwmon0/pwm1", []byte("120"), 0666)
				if err != nil {
					log.Fatalf("failed to write temp value: %v", err)
				}
				log.Printf("Temp above fan trigger temp: %d", tm)
				turnedOn = true
				break
			}

			if tm < fanTemp-10000 && turnedOn {
				err := ioutil.WriteFile("/sys/class/hwmon/hwmon0/pwm1", []byte("0"), 0666)
				if err != nil {
					log.Fatalf("failed to write temp value: %v", err)
				}
				log.Printf("Temp below fan trigger temp: %d", tm)
				turnedOn = false
				break
			}

		}

		time.Sleep(time.Duration(timeout) * time.Second)
	}
}

package main

import (
	"bufio"
	"log"
	"os/exec"
)

func main() {
	cmd := exec.Command("sleep", "10")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	reader := bufio.NewReader(stdout)
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	log.Print("DATA:")
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		log.Print(text)
	}

	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}

}

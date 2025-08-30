package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func getMaps(port string) []string {
	// Lots of credit to https://github.com/BishopFox/sliver/blob/master/implant/sliver/procdump/dump_linux.go

	var targetMaps []string

	memFile := fmt.Sprintf("/proc/%s/maps", port)

	file, err := os.OpenFile(memFile, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Printf("\r[!] Error opening file: %v\r\n", memFile)
		os.Exit(1)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " ")

		// The vvar region is for shared kernel variables
		// Other regions reserved by the kernel proved problematic, vdso and vsyscall
		region := parts[len(parts)-1]
		if region == "[vvar]" || region == "[vdso]" || region == "[vsyscall]" {
			continue
		}
		isFile := parts[3]

		// Then this is not a file
		if isFile == "00:00" {
			continue
		}

		// fmt.Println("Debug region:", region)
		maps := parts[0]
		targetMaps = append(targetMaps, maps)
	}

	return targetMaps

}

func getStackHeap(pid string) []string {
	var target_regions []string

	memFile := fmt.Sprintf("/proc/%s/maps", pid)
	file, err := os.OpenFile(memFile, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Printf("\r[!] Error Opening file: %v\r\n", memFile)
		os.Exit(1)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " ")
		region := parts[len(parts)-1]
		if region == "[stack]" || region == "[heap]" {
			target_regions = append(target_regions, parts[0])
		}
	}
	return target_regions
}

func getStartStop(memRange string) (int64, int64) {
	rangeMem := strings.Split(memRange, "-") // - dividing the start and the stop addresses
	memStart := rangeMem[0]
	memStop := rangeMem[1]
	memStartNum, _ := strconv.ParseInt(memStart, 16, 64)
	memEndNum, _ := strconv.ParseInt(memStop, 16, 64)

	return memStartNum, memEndNum
}


func doDump(memStart int64, memEnd int64, pid int) {

	memFileName := fmt.Sprintf("/proc/%v/mem", pid)
	pwd, err := os.Getwd()

	if err != nil {
		fmt.Printf("\r[!] Failed to get pwd: %v\r\n", err)
		os.Exit(1)
	}

	outputName := fmt.Sprintf("%v/dump.%v", pwd, pid)

	outFile, err := os.OpenFile(outputName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("\r[!] Failed to create dumpfile: %v\r\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	memFile, err := os.OpenFile(memFileName, os.O_RDONLY, 0)
	if err != nil {
		fmt.Printf("\r[!] Failed to open memFile: %v\r\n", memFile)
		fmt.Printf("\t--> Are you the root user...\r\n")
		os.Exit(1)
	}

	_, err = memFile.Seek(memStart, io.SeekStart)
	if err != nil {
		fmt.Printf("\r[!] Failed to seek to file start: %v\r\n", err)
		os.Exit(1)
	}

	bufferSize := int64(4096)
	buffer := make([]byte, bufferSize)

	bytesRemaining := memEnd - memStart
	for bytesRemaining > 0 {
		bytesToRead := bufferSize
		if bytesRemaining < bufferSize {
			bytesToRead = bytesRemaining
		}

		// Read the memory into the buffer
		n, err := memFile.Read(buffer[:bytesToRead])
		if err != nil && err != io.EOF {

			fmt.Printf("\r[!] Skipping, failed to read memory: %x-%x: %v\r\n", memStart, memEnd, err)
			return
		}

		// Write the buffer content to the output file
		if _, err := outFile.Write(buffer[:n]); err != nil {
			fmt.Printf("\r[!] Failed to write to output file: %v\r\n", err)
			return
		}

		bytesRemaining -= int64(n)
		if err == io.EOF {
			break
		}
	}
}

func main() {

	pid := flag.String("p", "", "The pid of the process to memory dump\r\n")
	singleShotRange := flag.String("r", "", "[Optional] The single memory range to target -> 77535b8d5000-77535b8d7000\r\n")
	justStackHeap := flag.Bool("s", false, "[Optional] Just dump stack and heap data (ignores -r)")
	flag.Parse()
	fmt.Printf("\r[+] goDumper started\r\n")

	if *pid == "" {
		fmt.Printf("\r[!] PID must be specified in order to dump process memory\r\n")
		flag.Usage()
		os.Exit(1)
	}

	pidInt, err := strconv.Atoi(*pid) // Why the hell do we do this, pid never needs to be an integer
	// Other than the fact that we abitrarily make the sprintf in doDump Require an int
	if err != nil {
		fmt.Printf("\r[!] Error converting target pid to int: %v\r\n", err)
		os.Exit(1)
	}
	fmt.Printf("\r[+] Target PID: %v\r\n", *pid)

	// full memory dump of the target pid
	if *singleShotRange == "" && *justStackHeap == false {
		// Getting maps
		targetMaps := getMaps(*pid)

		for _, line := range targetMaps {
			memStart, memEnd := getStartStop(line)
			doDump(memStart, memEnd, pidInt)
		}
		fmt.Printf("\r[+] Successful memory dump for pid: %v\r\n", pidInt)
	} else if *justStackHeap == true {
	// Just a dump of the stack and the heap
		targetMaps := getStackHeap(*pid)
		for _, line := range targetMaps {
			memStart, memEnd := getStartStop(line)
			doDump(memStart, memEnd, pidInt)
		}
		fmt.Printf("\r[+] Successful stack/heap dump for pid :%v\r\n", pidInt)
	} else if *singleShotRange != "" {
	// Just a dump of user specified memory range
		memStart, memEnd := getStartStop(*singleShotRange)
		doDump(memStart, memEnd, pidInt)
		fmt.Printf("\r[+] Successful memory dump for pid: %v\r\n", pidInt)
	}



}

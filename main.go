package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/shirou/gopsutil/v4/process"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: go run main.go <pid> <interval_in_seconds> <duration_in_seconds>")
		return
	}

	// 获取传入的PID和统计间隔、持续时间
	pid := os.Args[1]
	intervalSec, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatalf("Invalid interval: %v\n", err)
	}
	durationSec, err := strconv.Atoi(os.Args[3])
	if err != nil {
		log.Fatalf("Invalid duration: %v\n", err)
	}

	// 转换pid为整数
	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		log.Fatalf("Invalid PID: %v\n", err)
	}

	// 获取进程信息
	proc, err := process.NewProcess(int32(pidInt))
	if err != nil {
		log.Fatalf("Error retrieving process info: %v\n", err)
	}

	// 初始化监控数据
	var totalCPU float64
	var maxMemory uint64
	var totalReadOps, totalWriteOps uint64
	var totalReadBytes, totalWriteBytes uint64
	samples := 0

	// 记录起始I/O数据用于累加
	startIOCounters, err := proc.IOCounters()
	if err != nil {
		log.Fatalf("Error retrieving initial I/O counters: %v\n", err)
	}

	// 开始采集数据
	endTime := time.Now().Add(time.Duration(durationSec) * time.Second)
	for time.Now().Before(endTime) {
		// 获取 CPU 使用情况
		cpuPercent, err := proc.CPUPercent()
		if err != nil {
			log.Printf("Error retrieving CPU usage: %v\n", err)
		} else {
			totalCPU += cpuPercent
		}

		// 获取内存信息并更新最大内存占用
		memInfo, err := proc.MemoryInfo()
		if err != nil {
			log.Printf("Error retrieving memory info: %v\n", err)
		} else {
			if memInfo.RSS > maxMemory {
				maxMemory = memInfo.RSS
			}
		}

		// 获取 I/O 信息
		ioCounters, err := proc.IOCounters()
		if err != nil {
			log.Printf("Error retrieving I/O counters: %v\n", err)
		} else {
			totalReadOps = ioCounters.ReadCount - startIOCounters.ReadCount
			totalWriteOps = ioCounters.WriteCount - startIOCounters.WriteCount
			totalReadBytes = ioCounters.ReadBytes - startIOCounters.ReadBytes
			totalWriteBytes = ioCounters.WriteBytes - startIOCounters.WriteBytes
		}

		samples++

		// 等待下一个采集周期
		time.Sleep(time.Duration(intervalSec) * time.Second)
	}

	// 计算平均 CPU 占用率
	avgCPU := totalCPU / float64(samples)

	// 输出最终统计信息
	fmt.Printf("\nMonitoring completed over %d seconds (interval: %d seconds):\n", durationSec, intervalSec)
	fmt.Printf("Average CPU Usage: %.2f%%\n", avgCPU)
	fmt.Printf("Maximum Memory Usage: %v KiB\n", maxMemory/1024)
	fmt.Printf("Total Read Operations: %v\n", totalReadOps)
	fmt.Printf("Total Write Operations: %v\n", totalWriteOps)
	fmt.Printf("Total Read Bytes: %v Bytes(%v KiB)\n", totalReadBytes, totalReadBytes/1024)
	fmt.Printf("Total Write Bytes: %v Bytes(%v KiB)\n", totalWriteBytes, totalWriteBytes/1024)
}

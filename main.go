package main

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type RingBuffer struct {
	array    []int
	position int
	size     int
	m        sync.Mutex
}

func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{make([]int, size), -1, size, sync.Mutex{}}
}

func (r *RingBuffer) Push(el int) {
	r.m.Lock()
	defer r.m.Unlock()
	if r.position == r.size-1 {
		for i := 1; i <= r.size-1; i++ {
			r.array[i-1] = r.array[i]
		}
		r.array[r.position] = el
	} else {
		r.position++
		r.array[r.position] = el
	}
}

func (r *RingBuffer) Get() []int {
	if r.position <= 0 {
		return nil
	}
	r.m.Lock()
	defer r.m.Unlock()
	output := r.array[:r.position+1]
	r.position = -1
	return output
}

func read(next chan<- int, done chan bool) {
	scanner := bufio.NewScanner(os.Stdin)
	var data string
	for scanner.Scan() {
		data = scanner.Text()
		if strings.EqualFold(data, "exit") {
			log.Println("Программа завершила работу")
			close(done)
			return
		}
		i, err := strconv.Atoi(data)
		if err != nil {
			log.Println("Программа обрабатывает только целые числа")
			continue
		}
		next <- i
	}

}

func negativFilterStageInt(prevStageChan <-chan int, nextStageChan chan<- int, done <-chan bool) {
	for {
		select {
		case data := <-prevStageChan:
			if data > 0 {
				log.Printf("Проверка числа %v -> число подходит, оно больше 0, передано в следующую стадию пайплайна", data)
				nextStageChan <- data
			} else {
				log.Printf("Проверка числа %v -> число не подходит, оно меньше или равно 0", data)
			}
		case <-done:
			return

		}
	}
}

func notDivThreeFunc(prevStageChan <-chan int, nextStageChan chan<- int, done <-chan bool) {
	for {
		select {
		case data := <-prevStageChan:
			if data%3 == 0 {
				log.Printf("Проверка числа %v -> число подходит, оно кратно 3, передано в следующую стадию пайплайна", data)
				nextStageChan <- data
			} else {
				log.Printf("Проверка числа %v -> число не подходит, оно не кратно 3", data)
			}
		case <-done:
			return

		}
	}
}

func bufferStageFunc(prevStageChan <-chan int, nextStageChan chan<- int, done <-chan bool, size int, interval time.Duration) {
	buffer := NewRingBuffer(size)

	for {
		select {
		case data := <-prevStageChan:
			log.Printf("Число %v помещено в буфер", data)
			buffer.Push(data)
		case <-time.After(interval):
			bufferData := buffer.Get()
			if bufferData != nil {
				log.Printf("Прошло %v, начинается опустошение буфера", interval)
				for _, data := range bufferData {
					nextStageChan <- data
				}
			}
		case <-done:
			return

		}
	}
}

func main() {
	input := make(chan int)
	done := make(chan bool)
	go read(input, done)
	negativeFilterChan := make(chan int)
	go negativFilterStageInt(input, negativeFilterChan, done)
	notDivThreeChan := make(chan int)
	go notDivThreeFunc(negativeFilterChan, notDivThreeChan, done)
	bufferedIntChan := make(chan int)

	bufferSize := 10
	bufferDrainInt := 5 * time.Second
	go bufferStageFunc(notDivThreeChan, bufferedIntChan, done, bufferSize, bufferDrainInt)

	for {
		select {
		case data := <-bufferedIntChan:
			log.Println("Получены, ", data)
		case <-done:
			return

		}
	}
}

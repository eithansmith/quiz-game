package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

type Quiz struct {
	Questions  []Question
	Correct    int
	Incorrect  int
	Unanswered int
}

type Question struct {
	Prompt string
	Answer string
}

func main() {
	err := Run()
	if err != nil {
		log.Fatal(err)
	}
}

func Run() error {
	filename := flag.String("filename", "problems.csv", "The file to read")
	timeLimit := flag.Int("timeLimit", 30, "Time limit in seconds")
	random := flag.String("random", "Y", "Randomize question order")
	flag.Parse()

	file, err := os.Open(*filename)
	if err != nil {
		return fmt.Errorf("os open: %w", err)
	}
	defer func(file *os.File) {
		closeErr := file.Close()
		if closeErr != nil {
			err = fmt.Errorf("file close: %w", closeErr)
		}
	}(file)

	csvReader := csv.NewReader(file)
	records, err := csvReader.ReadAll()
	if err != nil {
		return fmt.Errorf("csv read: %w", err)
	}
	if len(records) < 1 {
		return fmt.Errorf("no records found")
	}

	if *random == "Y" {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		r.Shuffle(len(records), func(i, j int) {
			records[i], records[j] = records[j], records[i]
		})
	}

	quiz := Quiz{
		Questions: make([]Question, 0, len(records)),
	}

	for _, record := range records {
		if len(record) < 2 {
			return fmt.Errorf("invalid quiz question: %v", record)
		}
		quiz.Questions = append(quiz.Questions, Question{Prompt: record[0], Answer: record[1]})
	}

	ioReader := bufio.NewReader(os.Stdin)
	fmt.Printf("Welcome to the Quiz. You have %d seconds to complete the quiz.\n", *timeLimit)
	fmt.Printf("Press any key to start the quiz.\n")
	_, _ = ioReader.ReadString('\n')

	startTime := time.Now()
	timer := time.NewTimer(time.Duration(*timeLimit) * time.Second)
	done := make(chan bool, 1)

	go func() {
		for i, q := range quiz.Questions {
			fmt.Printf("%d) %s\n", i+1, q.Prompt)
			text, _ := ioReader.ReadString('\n')
			if strings.TrimSpace(text) == q.Answer {
				quiz.Correct++
			} else {
				quiz.Incorrect++
			}
		}
		done <- true
	}()

	select {
	case <-timer.C:
		fmt.Println("You have run out of time to complete the quiz.")
		quiz.Unanswered = len(quiz.Questions) - (quiz.Correct + quiz.Incorrect)
	case <-done:
		fmt.Println("You have completed the quiz.")
	}

	fmt.Printf("Quiz Results:\n")
	fmt.Println("Total Questions: ", len(quiz.Questions))
	fmt.Println("Correct: ", quiz.Correct)
	fmt.Println("Incorrect: ", quiz.Incorrect)
	fmt.Println("Unanswered: ", quiz.Unanswered)
	fmt.Printf("Time Elapsed: %.2f seconds\n", time.Since(startTime).Seconds())
	fmt.Printf("Score: %.2f\n", (float64(quiz.Correct)/float64(len(quiz.Questions)))*100)

	return nil
}

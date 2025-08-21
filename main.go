package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Question struct {
	Prompt string
	Answer string
}

type Quiz struct {
	Questions   []Question
	Correct     int
	Incorrect   int
	NotAnswered int
}

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	filenamePtr := flag.String("file", "problems.csv", "the quiz csv file")
	timeLimitPtr := flag.Int("limit", 30, "the time limit in seconds for the quiz")
	flag.Parse()

	filename := *filenamePtr
	timeLimit := *timeLimitPtr

	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	fileExtension := filepath.Ext(filename)

	if fileExtension != ".csv" {
		return fmt.Errorf("%s is not a .csv file", filename)
	}

	//goland:noinspection GoUnhandledErrorResult
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	quiz := Quiz{
		Questions: make([]Question, 0, len(records)),
		Correct:   0,
		Incorrect: 0,
	}

	for _, record := range records {
		if len(record) != 2 {
			return fmt.Errorf("expected 2 fields, got %d", len(record))
		}
		question := Question{
			Prompt: record[0],
			Answer: record[1],
		}
		quiz.Questions = append(quiz.Questions, question)
	}

	//fmt.Printf("%#v", quiz)

	ioReader := bufio.NewReader(os.Stdin)

	fmt.Println("Welcome to the Quiz game!")
	fmt.Println("Press any key to start the quiz.")
	_, _ = ioReader.ReadString('\n')

	startTime := time.Now()
	timer := time.NewTimer(time.Duration(timeLimit) * time.Second)

	done := make(chan bool, 1)

	go func() {
		for i, question := range quiz.Questions {
			fmt.Printf("%d) %s\n", i+1, question.Prompt)

			answer, _ := ioReader.ReadString('\n')
			answer = strings.Replace(answer, "\n", "", -1)
			answer = strings.TrimSpace(answer)

			expected := question.Answer
			expected = strings.TrimSpace(expected)

			if answer == expected {
				quiz.Correct++
			} else {
				quiz.Incorrect++
			}
		}
		done <- true
	}()

	// wait for the quiz to finish or the time to expire
	select {
	case <-timer.C:
		fmt.Println("Time limit exceeded!")
		quiz.NotAnswered = len(quiz.Questions) - (quiz.Correct + quiz.Incorrect)
	case <-done:
		fmt.Println("Quiz finished!")
	}

	fmt.Println("Total Questions: ", len(quiz.Questions))
	fmt.Printf("Correct: %d\n", quiz.Correct)
	fmt.Printf("Incorrect: %d\n", quiz.Incorrect)
	fmt.Printf("Not Answered: %d\n", quiz.NotAnswered)
	fmt.Printf("Time elapsed: %.2f seconds \n", time.Since(startTime).Seconds())
	fmt.Printf("Score: %.2f\n", (float64(quiz.Correct)/float64(len(quiz.Questions)))*100)

	return nil
}

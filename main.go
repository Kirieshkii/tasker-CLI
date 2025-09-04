package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

func main() {

	for {
		fmt.Println("Введите команду (Чтобы узнать список команд, введите: help)")
		request, err := bufio.NewReader(os.Stdin).ReadString('\n') // считываем строку
		if err != nil {
			fmt.Println("Ошибка при чтении ввода:", err)
			break //сделать fatal err
		}

		request = strings.TrimSpace(request) //удаляем пробелы

		var command string
		if len(request) > 1 {
			command = cutFirstWord(&request) //вырезать первое слово
		} else {
			fmt.Println("Ошибка при чтении пустой строки")
			continue //повторить цикл
		}

		switch command {
		case "list": //вывести список задач
			if len(request) > 0 {
				switch commandList := strings.Fields(request)[0]; commandList {
				case "todo", "in-progress", "done":
					searchTask(commandList)
				default:
					unknownCommand()
				}
			} else {
				fmt.Println("Общий список:")
				tasks := readerTaskJSON()
				printTasks(tasks)
			}

		case "add": //добавить задачу
			addTask(request)

		case "update": //обновить задачу
			updateTask(&request)

		case "mark-in-progress", "mark-done": //поменять статус
			if len(request) > 0 {
				taskNumberStr := cutFirstWord(&request)           //вырезаем первое слово (число) из запроса
				taskNumberInt, err := strconv.Atoi(taskNumberStr) //форматируем строку в int
				if err != nil {
					log.Fatalf("Ошибка: не верно введён номер задачи: %v", err)
				}
				newStatus := command[5:]
				changeStatus(newStatus, taskNumberInt)

			} else {
				fmt.Println("Ошибка: не указан номер задачи")
			}

		case "delete": // удалить задачу
			deleteTask(request)

		case "break": // завершение программы
			fmt.Println("Завершение программы")
			return

		case "help":
			help()

		default:
			unknownCommand()
		}

	}
}

func help() {
	fmt.Println("list - вывести все задачи")
	fmt.Println("list todo - вывести все запланированные задачи")
	fmt.Println("list in-progress - вывести все задачи в процессе")
	fmt.Println("list done - вывести все завершенные задачи")
	fmt.Println("add \"текст_задачи\"- добавить задачу")
	fmt.Println("update id \"текст_задачи\" - обновить задачу №id")
	fmt.Println("mark-in-progress id - задача №id в процессе")
	fmt.Println("mark-done id - завершить задачу №id")
	fmt.Println("delete id - удалить задачу №id")
	fmt.Println("break - завершить программу")
	fmt.Println("help - вывести список команд")
}

func unknownCommand() {
	fmt.Println("Неизвестная команда, попробуйте еще раз")
}

func cutFirstWord(pfullString *string) (firstWord string) {
	*pfullString = strings.TrimSpace(*pfullString)             // удаляем пробелы и переносы строк
	firstWord = strings.Fields(*pfullString)[0]                // извлекаем первое слово из строки запроса
	*pfullString = strings.TrimPrefix(*pfullString, firstWord) //удаляем первое слово из запроса
	*pfullString = strings.TrimSpace(*pfullString)             // удаляем пробелы и переносы строк
	return firstWord
}

type TaskCLI struct {
	ID          int
	Description string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func readerTaskJSON() []TaskCLI {
	file, err := os.OpenFile("taskList.json", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err = file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	var tasks []TaskCLI

	decoder := json.NewDecoder(file)
	if err = decoder.Decode(&tasks); err != nil && err != io.EOF {
		log.Fatal(err)
	}
	return tasks
}

func writerTaskJSON(tasks []TaskCLI) {
	file, err := os.OpenFile("taskList.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err = file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	encoder := json.NewEncoder(file)
	if err = encoder.Encode(tasks); err != nil {
		log.Fatal(err)
	}
}

func maxIDcounter(tasks []TaskCLI) int { //поиск максимального ID
	maxID := 0
	for i := range tasks {
		if tasks[i].ID > maxID {
			maxID = tasks[i].ID
		}
	}
	maxID++
	return maxID
}

func searchTask(status string) {
	var listTask []TaskCLI
	var taskFlag bool = false
	var counter int16 = 0

	tasks := readerTaskJSON()
	for i := range tasks {
		if tasks[i].Status == status {
			listTask = append(listTask, tasks[i])
			taskFlag = true
			counter++
		}
	}

	if taskFlag == true {
		fmt.Printf("Задачи в статусе %s:\n", status)
		printTasks(listTask)
	} else {
		fmt.Printf("Задачи в статусе %s отсутствуют\n", status)
	}
}

func printTasks(tasks []TaskCLI) {
	for _, v := range tasks {
		fmt.Printf("ID: %d, status:%s, Description:\"%s\" \n", v.ID, v.Status, v.Description)
	}
}

func printTask(task TaskCLI) {
	fmt.Printf("ID: %d, status:%s, Description:\"%s\" \n", task.ID, task.Status, task.Description)
}

func changeStatus(status string, taskNumber int) {
	tasks := readerTaskJSON()

	var checkNum bool = false

	for i := range tasks {
		if tasks[i].ID == taskNumber {
			checkNum = true
			if tasks[i].Status != status {
				tasks[i].Status = status
				tasks[i].UpdatedAt = time.Now()
				fmt.Printf("Статус задачи №%v успешно обновлен\n", taskNumber)
				writerTaskJSON(tasks)
			} else {
				fmt.Printf("Статус задачи №%v уже является таким\n", taskNumber)
			}

		}
	}

	if checkNum == false {
		fmt.Printf("Задача номер %v отсутствует\n", taskNumber)
	}
}

func addTask(request string) {
	taskText := strings.Trim(request, `"`) //убираем кавычки
	if len(taskText) > 2 {

		tasks := readerTaskJSON()
		taskCount := maxIDcounter(tasks) // счётчик id

		newTask := TaskCLI{
			ID:          taskCount,
			Description: taskText,
			Status:      "todo",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		tasks = append(tasks, newTask)

		writerTaskJSON(tasks)
		fmt.Println("Задача успешно добавлена:")
		printTask(newTask)

	} else {
		fmt.Println("Ошибка: пустая задача")
	}
}

func updateTask(request *string) {
	if len(*request) > 0 {
		taskNumberStr := cutFirstWord(request)            //вырезаем первое слово (число) из запроса
		taskNumberInt, err := strconv.Atoi(taskNumberStr) //форматируем строку в int
		if err != nil {
			log.Fatalf("Ошибка: не верно введён номер задачи: %v\n", err)
		}

		taskText := strings.Trim(*request, `"`) //убираем кавычки
		if len(taskText) > 0 {

			tasks := readerTaskJSON()

			var checkNum bool = false
			for i := range tasks {
				if tasks[i].ID == taskNumberInt {
					checkNum = true
					tasks[i].Description = taskText
					tasks[i].UpdatedAt = time.Now()
					writerTaskJSON(tasks)
					fmt.Printf("Задача номер %v успешно обновлена\n", taskNumberInt)
				}
			}

			if checkNum == false {
				fmt.Printf("Задача номер %v отсутствует\n", taskNumberInt)
			}
		} else {
			fmt.Println("Ошибка: не указано описание задачи")
		}
	} else {
		fmt.Println("Ошибка: не указан номер задачи")
	}
}

func deleteTask(request string) {
	if len(request) > 0 {
		taskNumberStr := cutFirstWord(&request)           //вырезаем первое слово (число) из запроса
		taskNumberInt, err := strconv.Atoi(taskNumberStr) //форматируем строку в int
		if err != nil {
			log.Fatalf("Ошибка: не верно введён номер задачи: %v", err)
		}
		tasks := readerTaskJSON()
		var checkNum bool = false
		var counter int
		for i := range tasks {
			if tasks[i].ID == taskNumberInt {
				checkNum = true
				counter = i
			}
		}
		if checkNum == true {
			tasks = slices.Delete(tasks, counter, counter+1)
			fmt.Printf("Задача №%v успешно удалена\n", taskNumberInt)
			writerTaskJSON(tasks)
		} else {
			fmt.Printf("Задача №%v не найдена или была удалена ранее\n", taskNumberInt)
		}
	} else {
		fmt.Println("Ошибка: не указан номер задачи")
	}
}

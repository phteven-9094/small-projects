package main

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const (
	WouldLikeToContinue = "Would you like to continue? (y/n): "
	ThankYou            = "Thank you for using PyDo List! Goodbye!"
	Invalid             = "Invalid option. Please try again."
	InvalidNumberInput  = "Invalid input. Please enter a number."
)

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./todo.db")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	mainMenu()
}

func mainMenu() {
	for {
		fmt.Println("Welcome to PyDo List!")
		fmt.Println("Please select an option:")
		fmt.Println(`
1. Create New Todo List
2. Open Existing Todo List
3. List All Todo Lists
4. Delete Todo List
        `)

		var listOption int
		if !scanInt(&listOption) {
			fmt.Println(InvalidNumberInput)
			continue
		}

		switch listOption {
		case 1:
			createNewTodoList()
			fmt.Println("New todo list created successfully.")
		case 2:
			openExistingTodoList()
		case 3:
			listAllTodoLists()
		case 4:
			deleteTodoList()
		default:
			fmt.Println(Invalid)
		}

		if !askToContinue() {
			fmt.Println(ThankYou)
			break
		}
	}
}

func createNewTodoList() {
	fmt.Print("Please provide a name for your new todo list: ")
	var tableName string
	fmt.Scan(&tableName)

	if !isValidTableName(tableName) {
		fmt.Println("Invalid table name. Please use only alphanumeric characters and underscores.")
		return
	}

	_, err := db.Exec(fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            task TEXT NOT NULL,
            completed INTEGER DEFAULT 0
        )
    `, tableName))
	if err != nil {
		fmt.Printf("Failed to create table '%s': %v\n", tableName, err)
		return
	}
	fmt.Printf("Table '%s' created successfully.\n", tableName)
}

func openExistingTodoList() {
	fmt.Print("Please select a todo list to open: ")
	var tableName string
	fmt.Scan(&tableName)

	if !tableExists(tableName) {
		fmt.Println("Todo list not found. Please try again.")
		return
	}

	handleTaskOptions(tableName)
}

func listAllTodoLists() {
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table';")
	if err != nil {
		fmt.Printf("Failed to list tables: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Println("Available Todo Lists:")
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			fmt.Printf("Error reading table name: %v\n", err)
			continue
		}
		fmt.Println(tableName)
	}
}

func deleteTodoList() {
	fmt.Print("Please select a todo list to delete: ")
	var tableName string
	fmt.Scan(&tableName)

	if !isValidTableName(tableName) {
		fmt.Println("Invalid table name. Please use only alphanumeric characters and underscores.")
		return
	}

	_, err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName))
	if err != nil {
		fmt.Printf("Failed to delete table '%s': %v\n", tableName, err)
		return
	}
	fmt.Printf("Table '%s' deleted successfully.\n", tableName)
}

func handleTaskOptions(tableName string) {
	for {
		fmt.Println(`
1. Add Task
2. Mark Task as Completed
3. Delete Task
4. List Tasks
        `)

		var taskOption int
		if !scanInt(&taskOption) {
			fmt.Println(InvalidNumberInput)
			continue
		}

		switch taskOption {
		case 1:
			addTask(tableName)
		case 2:
			markTaskAsCompleted(tableName)
		case 3:
			deleteTask(tableName)
		case 4:
			listTasks(tableName)
		default:
			fmt.Println(Invalid)
		}

		if !askToContinue() {
			break
		}
	}
}

func addTask(tableName string) {
	fmt.Print("Please provide a task to add: ")
	var task string
	fmt.Scan(&task)

	_, err := db.Exec(fmt.Sprintf("INSERT INTO %s (task) VALUES (?)", tableName), task)
	if err != nil {
		fmt.Printf("Failed to add task to '%s': %v\n", tableName, err)
		return
	}
	fmt.Printf("Task '%s' added to '%s'.\n", task, tableName)
}

func markTaskAsCompleted(tableName string) {
	fmt.Print("Please provide the ID of the task to mark as completed: ")
	var taskID int
	if !scanInt(&taskID) {
		fmt.Println(InvalidNumberInput)
		return
	}

	_, err := db.Exec(fmt.Sprintf("UPDATE %s SET completed = 1 WHERE id = ?", tableName), taskID)
	if err != nil {
		fmt.Printf("Failed to mark task as completed in '%s': %v\n", tableName, err)
		return
	}
	fmt.Printf("Task with ID '%d' marked as completed in '%s'.\n", taskID, tableName)
}

func deleteTask(tableName string) {
	fmt.Print("Please provide the ID of the task to delete: ")
	var taskID int
	if !scanInt(&taskID) {
		fmt.Println(InvalidNumberInput)
		return
	}

	_, err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = ?", tableName), taskID)
	if err != nil {
		fmt.Printf("Failed to delete task from '%s': %v\n", tableName, err)
		return
	}
	fmt.Printf("Task with ID '%d' deleted from '%s'.\n", taskID, tableName)
}

func listTasks(tableName string) {
	rows, err := db.Query(fmt.Sprintf("SELECT * FROM %s", tableName))
	if err != nil {
		fmt.Printf("Failed to list tasks in '%s': %v\n", tableName, err)
		return
	}
	defer rows.Close()

	fmt.Printf("Tasks in '%s':\n", tableName)
	for rows.Next() {
		var id int
		var task string
		var completed int
		if err := rows.Scan(&id, &task, &completed); err != nil {
			fmt.Printf("Error reading task: %v\n", err)
			continue
		}
		status := "Not Completed"
		if completed == 1 {
			status = "Completed"
		}
		fmt.Printf("%d: %s - %s\n", id, task, status)
	}
}

func askToContinue() bool {
	fmt.Print(WouldLikeToContinue)
	var response string
	fmt.Scan(&response)
	return strings.ToLower(response) == "y"
}

func tableExists(tableName string) bool {
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name=?;", tableName)
	if err != nil {
		fmt.Printf("Failed to check if table exists: %v\n", err)
		return false
	}
	defer rows.Close()

	return rows.Next()
}

func isValidTableName(tableName string) bool {
	// Allow only alphanumeric characters and underscores
	validName := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	return validName.MatchString(tableName)
}

func scanInt(target *int) bool {
	_, err := fmt.Scan(target)
	return err == nil
}

import sqlite3

# Constants
WOULD_LIKE_TO_CONTINUE = "Would you like to continue? (y/n): "
THANK_YOU = "Thank you for using PyDo List! Goodbye!"
INVALID = "Invalid option. Please try again."
INVALID_NUMBER_INPUT = "Invalid input. Please enter a number."

# Database connection
conn = sqlite3.connect('todo.db')
cursor = conn.cursor()


def create_table(table_name: str) -> None:
    """Creates a new table for a todo list."""
    cursor.execute(f'''
        CREATE TABLE IF NOT EXISTS {table_name} (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            task TEXT NOT NULL,
            completed INTEGER DEFAULT 0
        )
    ''')
    conn.commit()
    print(f"Table '{table_name}' created successfully.")


def list_tasks(table_name: str) -> None:
    """Lists all tasks in a given todo list."""
    cursor.execute(f'SELECT * FROM {table_name}')
    tasks = cursor.fetchall()
    if tasks:
        print(f"Tasks in '{table_name}':")
        for task in tasks:
            print(f"{task[0]}: {task[1]} - {'Completed' if task[2] else 'Not Completed'}")
    else:
        print(f"No tasks found in '{table_name}'.")


def list_tables() -> list:
    """Returns a list of all todo lists (tables)."""
    cursor.execute("SELECT name FROM sqlite_master WHERE type='table';")
    tables = cursor.fetchall()
    return [table[0] for table in tables]


def delete_table(table_name: str) -> None:
    """Deletes a todo list (table)."""
    cursor.execute(f'DROP TABLE IF EXISTS {table_name}')
    conn.commit()
    print(f"Table '{table_name}' deleted successfully.")


def add_task(table_name: str, task: str) -> None:
    """Adds a new task to a todo list."""
    cursor.execute(f'INSERT INTO {table_name} (task) VALUES (?)', (task,))
    conn.commit()
    print(f"Task '{task}' added to '{table_name}'.")


def complete_task(table_name: str, task_id: int) -> None:
    """Marks a task as completed in a todo list."""
    cursor.execute(f'UPDATE {table_name} SET completed = 1 WHERE id = ?', (task_id,))
    conn.commit()
    print(f"Task with ID '{task_id}' marked as completed.")


def delete_task(table_name: str, task_id: int) -> None:
    """Deletes a task from a todo list."""
    cursor.execute(f'DELETE FROM {table_name} WHERE id = ?', (task_id,))
    conn.commit()
    print(f"Task with ID '{task_id}' deleted from '{table_name}'.")


def ask_to_continue() -> bool:
    """Asks the user if they want to continue."""
    response = input(WOULD_LIKE_TO_CONTINUE).strip().lower()
    return response == 'y'


def handle_task_options(table_name: str) -> None:
    """Handles task-related operations for a selected todo list."""
    print("""
        1. Add Task
        2. Mark Task as Completed
        3. Delete Task
        4. List Tasks
    """)
    try:
        task_options = int(input("What would you like to do? "))
    except ValueError:
        print(INVALID_NUMBER_INPUT)
        return

    if task_options == 1:
        task_to_add = input("Please provide a task to add: ").strip()
        add_task(table_name, task_to_add)
    elif task_options == 2:
        try:
            task_id = int(input("Please provide the ID of the task to mark as completed: "))
            complete_task(table_name, task_id)
        except ValueError:
            print(INVALID_NUMBER_INPUT)
    elif task_options == 3:
        try:
            task_id = int(input("Please provide the ID of the task to delete: "))
            delete_task(table_name, task_id)
        except ValueError:
            print(INVALID_NUMBER_INPUT)
    elif task_options == 4:
        list_tasks(table_name)
    else:
        print(INVALID)


def main_menu() -> None:
    """Displays the main menu and handles user input."""
    while True:
        print("Welcome to PyDo List!")
        print("Please select an option: ")
        try:
            list_options = int(input("""
                1. Create New Todo List
                2. Open Existing Todo List
                3. List All Todo Lists
                4. Delete Todo List
                """))
        except ValueError:
            print("Invalid input. Please enter a number.")
            continue

        if list_options == 1:
            new_table_name = input("Please provide a name for your new todo list: ").strip()
            create_table(new_table_name)
        elif list_options == 2:
            table_to_open = input("Please select a todo list to open: ").strip()
            if table_to_open not in list_tables():
                print("Todo list not found. Please try again.")
                continue
            print(f"You've selected the following todo list: {table_to_open}")
            handle_task_options(table_to_open)
        elif list_options == 3:
            print("Available Todo Lists:")
            for table in list_tables():
                print(table)
        elif list_options == 4:
            table_to_delete = input("Please select a todo list to delete: ").strip()
            delete_table(table_to_delete)
        else:
            print(INVALID)

        if not ask_to_continue():
            print(THANK_YOU)
            break


if __name__ == "__main__":
    try:
        main_menu()
    finally:
        conn.close()
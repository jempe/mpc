# MPC - My Private Collection

MPC is a web based media player that allows you to manage and stream your video library. It includes features such as video scanning, and remote control.

## Features

- Manage and stream your video library
- Video scanning
- Remote control
- Web server to serve video content and handle requests

## Requirements

- Go 1.16 or higher

## Installation

1. Clone the repository:

    ```sh
    git clone https://github.com/jempe/mpc.git
    cd mpc
    ```

2. Build the project:

    ```sh
    go build -o mpc main.go
    ```

3. Run the project:

    ```sh
    ./mpc -path="/path/to/your/videos" -config="/path/to/config/folder"
    ```

## Configuration

- `-path`: Define the path of your videos folder.
- `-config`: Define the path of the config folder.

## Usage

1. Start the server:

    ```sh
    ./mpc -path="/path/to/your/videos" -config="/path/to/config/folder"
    ```

2. Access the server in your web browser at `http://<local_ip>:3000`.

## Project Structure

- `auth`: Handles user authentication.
- `library`: Manages the video library.
- `remote`: Manages remote control functionality.
- `server`: Handles HTTP server and routes.
- `storage`: Manages storage and database operations.
- `users`: Manages user data and operations.
- `utils`: Contains utility functions.
- `tmpl`: Contains HTML templates.
- `html`: Contains static files like JavaScript, CSS, fonts, and images.

## License

This project is licensed under the GPLv3 License. See the [LICENSE](LICENSE) file for details.

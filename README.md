# gimer

gimer is a CLI timer tool written in Go.
It allows users to create a timer for a specified duration and plays a notification sound when the timer ends.

![image](https://github.com/user-attachments/assets/204edf33-8f29-414f-9426-32b9302bb8ab)
![image](https://github.com/user-attachments/assets/5a70576f-b668-41c6-b065-f2a748cb5d01)

## Features

- Set time duration (seconds, minutes, hours)
- Pause and resume the timer
- Play a notification sound when the timer ends
- Save and reuse timers
- Add descriptions to timers

## Installation

There are two ways to install gimer.

### 1. Clone from GitHub

```sh
# Clone gimer from GitHub
git clone git@github.com:kyaoi/gimer.git
cd gimer

# Install dependencies
go mod tidy

# Build gimer
go build -o gimer
```

### 2. Install using go install

```sh
go install github.com/kyaoi/gimer@latest
```

## Usage

### Create a new timer

```sh
# Example: Create a timer for 1 hour and 30 minutes
gimer start -H 1 -M 30
```

### Stop a timer

You can either run `gimer stop` in a separate terminal or press the q key on the timer screen.

### Pause and resume a timer

Press the space key on the timer screen to pause or resume the timer.

### Other commands

To see detailed usage instructions, run the help command:

```sh
gimer -h
gimer start -h
gimer stop -h
gimer status -h
```

[日本語はこちら (Japanese version)](README.ja.md)

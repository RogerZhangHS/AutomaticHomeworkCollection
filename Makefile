
SHELL = /bin/sh

# define the source/target folders
SRC_ROOT = $(shell pwd)
CLIENT   = $(SRC_ROOT)/client
SERVER   = $(SRC_ROOT)/server

# Raspberry Pi binaries
PiScannerStudent: $(CLIENT)/scannerStudent.go
	go build -o $(CLIENT)/PiScannerStudent $^

WebAppStudent: $(CLIENT)/webappStudent.go
	go build -o $(CLIENT)/WebAppStudent $^

PI_TARGETS = PiScannerStudent WebAppStudent

clients: $(PI_TARGETS)

# Server binary
APIServer: $(SERVER)/main.go
	go build -o $(SERVER)/APIServer $^

# all components
all: clients APIServer

clean:
	rm -f $(addprefix $(CLIENT)/, $(PI_TARGETS))
	rm -f $(addprefix $(SERVER)/, APIServer)

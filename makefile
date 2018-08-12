#!make
include .env

DIR:=${SYNC_DIR}

# ensure SYNC_DIR is set
ifndef DIR
  $(error SYNC_DIR not defined)
endif

# extra checks to prevent terrrible things from happening
ifeq ($(DIR),$(filter $(DIR),/ ~/ ~))
  $(error invalid SYNC_DIR value)
endif

all: clean build run
clean:
	cd $(DIR) && rm -rf $(DIR)/* && cd -
build:
	go build -o client
run:
	./client

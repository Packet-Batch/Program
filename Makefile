BUILD_DIR = build

all: program

program:
	go build -o $(BUILD_DIR)/pb

install:
	cp -f $(BUILD_DIR)/pb /usr/bin

clean:
	find $(BUILD_DIR) -type f ! -name ".*" -exec rm -f {} +

.PHONY: program
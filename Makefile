# PkgConfig
# ------------------------------
PKG_CONF = pkg-config

ifeq ($(shell which $(PKG_CONF)),)
$(error "Package config not found. Please install it on server.")
endif

# ------------------------------
# PkgConfig End

# Common
# ------------------------------
# Compiler (intended for Clang, lol)
CC = clang

# Directories.
BUILD_DIR := build
SRC_DIR := src
MODULES_DIR := modules

MAIN_OUT := pcktbatch
MAIN_SRC := main.c

SRC_DIR_COMMON = $(SRC_DIR)/common
COMMON_BUILD_DIR = $(BUILD_DIR)/common

# Sources and objects
COMMON_SRCS := $(wildcard $(SRC_DIR_COMMON)/*.c)
COMMON_OBJS := $(patsubst $(SRC_DIR_COMMON)/%.c,$(COMMON_BUILD_DIR)/%.o,$(COMMON_SRCS))

# Flags
GLOBAL_FLAGS := -O2 -g
MAIN_FLAGS := -pthread -lelf -lz

INCS := -I $(SRC_DIR) -I $(MODULES_DIR)/libbpf/src

# ------------------------------
# Common End

# Tech
# ------------------------------
TECH_SRC_DIR = $(SRC_DIR)/tech
TECH_BUILD_DIR = $(BUILD_DIR)/tech

TECH_SRCS := $(shell find $(TECH_SRC_DIR) -name '*.c')
TECH_OBJS := $(patsubst $(TECH_SRC_DIR)/%.c,$(TECH_BUILD_DIR)/%.o,$(TECH_SRCS))

# ------------------------------
# Tech End

# LibBPF
# ------------------------------
# Directories.
LIBBPF_DIR := $(MODULES_DIR)/libbpf

LIBBPF_SRC_DIR := $(LIBBPF_DIR)/src

# Objects.
LIBBPF_OBJS_DIR := $(LIBBPF_SRC_DIR)/staticobjs

LIBBPF_OBJS = $(addprefix $(LIBBPF_OBJS_DIR)/, $(notdir $(wildcard $(LIBBPF_OBJS_DIR)/*.o)))
# ------------------------------
# LibBPF End

# JSON-C
# ------------------------------
JSONC_DIR := $(MODULES_DIR)/json-c
# ------------------------------
# JSON-C End


# Chains
# ------------------------------
all: main

jsonc:
	@mkdir -p $(JSONC_DIR)/build
	cd $(JSONC_DIR)/build && cmake ../

jsonc_install:
	cd $(JSONC_DIR)/build && make && make install

libbpf:
	$(MAKE) -C $(LIBBPF_SRC_DIR)

$(COMMON_BUILD_DIR)/%.o: $(SRC_DIR_COMMON)/%.c
	@mkdir -p $(dir $@)
	$(CC) $(GLOBAL_FLAGS) $(INCS) -c $< -o $@

$(TECH_BUILD_DIR)/%.o: $(TECH_SRC_DIR)/%.c
	@mkdir -p $(dir $@)
	$(CC) $(GLOBAL_FLAGS) $(INCS) -c $< -o $@

main: $(COMMON_OBJS) $(TECH_OBJS) libbpf
	@mkdir -p $(BUILD_DIR)
	$(CC) $(INCS) $(GLOBAL_FLAGS) $(MAIN_FLAGS) -o $(BUILD_DIR)/$(MAIN_OUT) $(shell $(PKG_CONF) --libs json-c) $(LIBBPF_OBJS) $(COMMON_OBJS) $(TECH_OBJS) $(SRC_DIR)/$(MAIN_SRC)

clean:
	$(MAKE) -C $(LIBBPF_SRC_DIR)/ clean
	$(MAKE) -C $(JSONC_DIR)/build clean
	
	rm -rf $(BUILD_DIR)

install:
	cp -f $(BUILD_DIR)/$(MAIN_OUT) /usr/bin/$(MAIN_OUT)

.PHONY: all libbpf clean install

.DEFAULT: all
# ------------------------------
# Chains End
SHELL := /bin/sh

CORE_ROOT := $(abspath $(CURDIR)/../..)
TONSUO_PREFIX := $(CORE_ROOT)/libs/bin/tongsuo
TONSUO_INCLUDE := $(TONSUO_PREFIX)/include
TONSUO_LIB := $(TONSUO_PREFIX)/libs

CC ?= gcc
CFLAGS := -Wall -Wextra -O2 -I$(TONSUO_INCLUDE)
LDFLAGS := -L$(TONSUO_LIB) -lcrypto -lssl -Wl,-rpath,$(TONSUO_LIB)

SRCS := $(wildcard src/*.c)
BINS := $(patsubst src/%.c,bin/%,$(SRCS))

.PHONY: all clean

all: $(BINS)

bin/%: src/%.c
	@mkdir -p bin
	$(CC) $(CFLAGS) -o $@ $< $(LDFLAGS)

clean:
	rm -rf bin

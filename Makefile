GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
BINARY_NAME=proto-parse-tcaplus
BIN_NAME=bin
TEST_HOME=testdata
all: build
build:
	$(GOBUILD) -o $(BINARY_NAME)
	mkdir -p ${BIN_NAME}
	mv ${BINARY_NAME} ${BIN_NAME}
	cp -r ${TEST_HOME} ${BIN_NAME}
	cp -r config ${BIN_NAME}/
clean:
	$(GOCLEAN)
	rm -rf ${BIN_NAME}

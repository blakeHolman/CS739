set working-directory := '..'

# list project 1 just recipes
default:
    @just --list p1 --unsorted

# install extra dependencies of your KV code
deps:
    echo "Installing Go and Protobuf..." ; \
    sudo apt update && sudo apt install -y golang-go protobuf-compiler ; \
    sudo ln -sf /usr/lib/go/bin/go /usr/bin/go ; \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest ; \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest ; \
    export PATH=$PATH:$(go env GOPATH)/bin ; \


# build your executables in release mode
build:
    if [ ! -f go.mod ]; then \
        echo "Initializing Go modules..." ; \
        go mod init madkv ; \
        go mod tidy ; \
    fi ; \
    export PATH=$PATH:$(go env GOPATH)/bin ; \
    mkdir -p kvstore ; \
    protoc --go_out=.. --go-grpc_out=.. --proto_path=src --experimental_allow_proto3_optional src/kvstore.proto ; \
    go mod tidy ; \
    go build -o server src/server.go ; \
    go build -o client src/client.go

# clean the build of your executables
clean:
    # FIXME: your clean commands here
    rm -f server client src/kvstore.pb.go src/kvstore_grpc.pb.go

# run your KV store server
server listen="0.0.0.0:3777":
    ./server --listen "{{listen}}"

# run you KV store client in stdin/out interface mode
client server="127.0.0.1:3777":
    ./client "{{server}}"

# run your KV store testcase 1 client
test1 server="127.0.0.1:3777":
    # Test 1 checks to ensure basic commands work for single client
    sh -c 'printf "PUT name Alice\n\
    GET name\n\
    SWAP name Bob\n\
    GET name\n\
    DELETE name\n\
    SCAN a z\n\
    STOP\n" | ./client "{{server}}"'

# run your KV store testcase 2 client
test2 server="127.0.0.1:3777":
    # Tests multiple commands from single client, ensures keys are not found
    sh -c 'printf "DELETE name\n\
    GET name\n\
    SWAP name Ted\n\
    PUT name Jim\n\
    DELETE name\n\
    PUT name Bob\n\
    GET name\n\
    DELETE name\n\
    GET name\n\
    DELETE name\n\
    SCAN A Z\n\
    PUT name Bill\n\
    PUT address 1234\n\
    PUT car Aston\n\
    PUT name James\n\
    GET name\n\
    GET car\n\
    SWAP name Bond\n\
    GET name\n\
    SCAN a z\n\
    STOP\n" | ./client "{{server}}"'

# run your KV store testcase 3 client
test3 server="127.0.0.1:3777":
    # Test 3 tests if multiple clients can issue non-conflicting requests
    { sh -c '( \
        printf "PUT key1 Value1\n" ;\
    	sleep 0.2; \
    	printf "GET key1\n" ; \
    	sleep 0.2; \
    	printf "PUT key2 Value2\n" ; \
    	sleep 0.2; \
    	printf "GET key2\n" ; \
    	sleep 0.2; \
    	printf "SCAN a z\n" ; \
    	sleep 0.2; \
    	printf "STOP\n" \
    ) | ./client "{{server}}"'; } &

    { sh -c '( \
    	printf "\nPUT keyA DataA\n" ; \
    	sleep 0.2 ; \
    	printf "GET keyA\n" ; \
    	sleep 0.2 ; \
    	printf "PUT keyB DataB\n" ; \
    	sleep 0.2 ; \
    	printf "GET keyB\n" ; \
    	sleep 0.2 ; \
    	printf "SCAN a z\n" ; \
    	sleep 0.2 ; \
    	printf "STOP\n" \
    ) | ./client "{{server}}"'; } & 

    wait

    sleep 1
    exit 0

# run your KV store testcase 4 client(s)
test4 server="127.0.0.1:3777":
    # Test4: This test shows conflicting requests from 2 clients
    # It shows a phantom read in SCAN due to its read lock
    { sh -c '( \
        printf "PUT shared_key c1_A\n"; \
        sleep 0.1; \
        printf "GET shared_key\n"; \
        sleep 0.1; \
        printf "SWAP shared_key C1_B\n"; \
        sleep 0.1; \
        printf "STOP\n" \
    ) | ./client "{{server}}"'; } &
    { sh -c '( \
        printf "PUT shared_key C2_X\n"; \
        sleep 0.1; \
        printf "GET shared_key\n"; \
        sleep 0.1; \
        printf "SCAN a z"; \
        sleep 0.3 ; \
        printf "DELETE shared_key\n"; \
        sleep 0.1; \
        printf "STOP\n" \
    ) | ./client "{{server}}"'; } &
    wait
    sleep 1
    exit 0

# run your KV store testcase 5 client(s)
test5 server="127.0.0.1:3777":
    # FIXME: your testcase 5 client(s) run commands here
    { sh -c '( \
        printf "PUT key1 C1_1\n"; \
        sleep 0.3; \
        printf "GET key2\n"; \
        sleep 0.3; \
        printf "SWAP key1 C1_2\n"; \
        sleep 0.3; \
        printf "GET key3\n"; \
        sleep 0.3; \
        printf "DELETE key2\n"; \
        sleep 0.3; \
        printf "SCAN a z\n"; \
        sleep 0.3; \
        printf "STOP\n" \
    ) | ./client "{{server}}"'; } &

    { sh -c '( \
        printf "PUT key2 C2_X\n"; \
        sleep 0.3; \
        printf "GET key1\n"; \
        sleep 0.3; \
        printf "SWAP key3 C2_Y\n"; \
        sleep 0.3; \
        printf "GET key2\n"; \
        sleep 0.3; \
        printf "DELETE key3\n"; \
        sleep 0.3; \
        printf "SCAN a z\n"; \
        sleep 0.3; \
        printf "STOP\n" \
    ) | ./client "{{server}}"'; } &

    wait
    sleep 3
    exit 0

# kill all processes of your KV store system
kill:
    # Kills server and client
    -pkill -f "./server" &> /dev/null
    -pkill -f "./client" &> /dev/null
    sleep 1

# NOTE: feel free to add more recipes as you see fit...
#       also feel free to add extra parameters to the recipes as you see fit,
#       but don't change the existing parameters

# launch a long-running KV store service
service listen="0.0.0.0:3777":
    just p1::build
    just utils::build
    cargo run -p runner -r --bin service -- \
        --just-args p1::server "{{listen}}"

# ensure a subdir under 'tmp/' exists
tmpdir subdir:
    mkdir -p "/tmp/madkv-p1/{{subdir}}"

# run a student-provided testcase
testcase num server="127.0.0.1:3777": (tmpdir "tests")
    just p1::build
    just "p1::test{{num}}" "{{server}}" \
        | tee "/tmp/madkv-p1/tests/test{{num}}.log"
    just p1::kill

# run a fuzz testing scenario
fuzz nclis conflict="no" server="127.0.0.1:3777": (tmpdir "fuzz")
    just p1::build
    just utils::build
    cargo run -p runner -r --bin fuzzer -- \
        --num-clis "{{nclis}}" \
        {{ if conflict == "no" { "" } else { "--conflict" } }} \
        --just-args p1::client "{{server}}" \
        | tee "/tmp/madkv-p1/fuzz/fuzz-{{nclis}}-{{conflict}}.log"
    just p1::kill

# run a YCSB benchmark workload
bench nclis wload server="127.0.0.1:3777": (tmpdir "bench")
    just p1::build
    just utils::build
    just utils::ycsb
    cargo run -p runner -r --bin bencher -- \
        --num-clis "{{nclis}}" \
        --workload "{{wload}}" \
        --just-args p1::client "{{server}}" \
        | tee "/tmp/madkv-p1/bench/bench-{{nclis}}-{{wload}}.log"
    just p1::kill

# generate .md report template from existing results (wip)
report:
    python3 sumgen/proj1.py

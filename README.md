# Project 1 Spring CS739

## Environment Setup

Ensure the following is run on each machine:

```
git clone https://github.com/josehu07/madkv.git
cd madkv
git checkout main
git pull
git checkout proj
git merge main

curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh

curl https://pyenv.run | bash
tee -a $HOME/.bashrc <<EOF
export PYENV_ROOT="\$HOME/.pyenv"
command -v pyenv >/dev/null || export PATH="\$PYENV_ROOT/bin:\$PATH"
eval "\$(pyenv init -)"
EOF
source $HOME/.bashrc

sudo apt update
sudo apt install libssl-dev zlib1g-dev
pyenv install 3.12
pyenv global 3.12

# pip packages
pip3 install numpy matplotlib termcolor

# add just gpg
wget -qO - 'https://proget.makedeb.org/debian-feeds/prebuilt-mpr.pub' | gpg --dearmor | sudo tee /usr/share/keyrings/prebuilt-mpr-archive-keyring.gpg 1> /dev/null
echo "deb [arch=all,$(dpkg --print-architecture) signed-by=/usr/share/keyrings/prebuilt-mpr-archive-keyring.gpg] https://proget.makedeb.org prebuilt-mpr $(lsb_release -cs)" | sudo tee /etc/apt/sources.list.d/prebuilt-mpr.list

# apt install
sudo apt update
sudo apt install tree just default-jre liblog4j2-java

```

Install dependencies and build files:
```
just p1::deps
just p1::build
```


## Server

To run the server, ensure the following code is uploaded on the server machine located in the src directory:

**server.cpp**
**kvstore.proto**

Next, find the ip address of the server machine using:
```
ip a
```

*Note: If running both client server on a single machine, this step is not needed. The just functions default to 0.0.0.0 and loopback address.*

Then, the server can be started using the command:
```
just p1::server <ip>:<port>
```

## Client 

To run the client, ensure the following code is uploaded on the client machine located in the src directory:

**client.cpp**
**kvstore.proto**

To run the client, first ensure the server is running. Then, run the below command to start an interactive shell.

```
just p1::client <server ip>
```

*Note: If no server IP is provided, it will default to 127.0.0.1 *

## Testing

There are a number of different tests you can run to test the operations as well as the concuncurrency.

### Single Client Tests
Test 1 and 2 are single client tests. This ensures basic operations are working.
Test 1 ensures basic commands work. It can be run with the command:
```
just p1::test1 <server ip>
```
Test 2 ensures the server returns null values or not founds when approriate. This can be run with the following:
```
just p1::test2 <server ip>
```

### Multi-Client Tests
Test 3, 4, and 5 are multi-client tests. This ensures that multiple clients can make concurrent calls without causing a conflict.
Test 3 focuses on non-conflicting calls. This ensures clients can access the server and run commands without issue. It can be run with:
```
just p1::test3 <server ip>
```
Test 4 focuses on read-read and read-write conflicts. It can be run using the command:
```
just p1::test4 <server ip>
```
Test 5 focuses on write-write conflicts. It can be run using the command:
```
just p1::test5 <server ip>
```

*Note: To log the outputs of the tests, you can run the provided function below. This will create a log in the /tmp/madkv-p1/tests/ directory.*

```
just testcase <num> <server ip>
```

### Fuzz Testing and Benchmarking

The provided fuzz testing and YCSB benchmarking can be run with the commands:
```
just fuzz <num_clients> <conflict yes/no> <server ip>
just bench <num_clients> <workload> <server ip>
```

## Project Report

Our project report is located in the report directory as **proj1.pdf**.

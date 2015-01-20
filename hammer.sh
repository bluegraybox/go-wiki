#!/usr/local/bin/bash

# Start X processes in parallel, each making Y requests in sequence.

numProcs=${1:-10}
numReqs=${2:-10}

url=http://localhost:8888/

echo "Starting $numProcs processes in parallel, each making $numReqs requests in sequence to $url"

function reqSeq() { 
    echo "starting $i";
    for j in `seq $numReqs`;
    do
        curl -sL $url > /dev/null;
    done;
    echo "$i done"
}

function startProcs() 
{ 
    for i in `seq $numProcs`;
    do
        reqSeq &
    done;
    wait
}

time startProcs

package examples

default is_positive = {"result": false}

is_positive = {"result": true} {
    input.number > 0
}

default is_big = {"result": false}

is_big = {"result": true} {
    input.number > 100
}
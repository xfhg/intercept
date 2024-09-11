package policies

default allow = false

allow {
    input.user == "alice"
    input.action == "read"
}

allow {
    input.user == "bob"
    input.action == "write"
}
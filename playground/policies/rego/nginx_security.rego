package nginx

import future.keywords

default allow = false

# Helper function to check if a string contains a substring
string_contains(str, substr) {
    contains(str, substr)
}

# Main rule to allow the configuration if all checks pass
allow {
    count(violations) == 0
}

# Violations report
violations[msg] {
    server_tokens_on
    msg := "server_tokens should be set to off"
}

violations[msg] {
    weak_ssl_protocols
    msg := "Weak SSL/TLS protocols should be disabled"
}

violations[msg] {
    not https_redirection
    msg := "HTTPS redirection is missing"
}

violations[msg] {
    not directory_listing_off
    msg := "Directory listing should be disabled"
}

violations[msg] {
    not security_headers_present
    msg := "Security headers are missing"
}

violations[msg] {
    not unsafe_methods_restricted
    msg := "Unsafe HTTP methods should be restricted"
}

violations[msg] {
    not access_logging_enabled
    msg := "Access logging should be enabled"
}

violations[msg] {
    not client_max_body_size_set
    msg := "client_max_body_size should be set"
}

violations[msg] {
    not strong_ssl_ciphers
    msg := "Weak SSL ciphers should be disabled"
}

violations[msg] {
    not allowed_server_names
    msg := "Server name not in the list of allowed server names"
}

# Helper rules
server_tokens_on {
    some line in input.lines
    contains(line, "server_tokens on")
}

weak_ssl_protocols {
    some line in input.lines
    contains(line, "ssl_protocols")
    not contains(line, "TLSv1.2")
    not contains(line, "TLSv1.3")
}

https_redirection {
    some line in input.lines
    string_contains(line, "return 301 https://")
}

directory_listing_off {
    not string_contains(input.content, "autoindex on")
}

security_headers_present {
    headers := [
        "Strict-Transport-Security",
        "X-Frame-Options",
        "X-Content-Type-Options",
        "Content-Security-Policy"
    ]
    
    count([h | some h in headers; some line in input.lines; string_contains(line, h)]) == count(headers)
}

unsafe_methods_restricted {
    some block in input.blocks
    block.directive == "location"
    some line in block.block
    string_contains(line, "limit_except GET POST")
    not string_contains(line, "allow all")
}

access_logging_enabled {
    not string_contains(input.content, "access_log off")
}

client_max_body_size_set {
    some line in input.lines
    regex.match(`client_max_body_size\s+\d+[mMkK];`, line)
}

strong_ssl_ciphers {
    some line in input.lines
    string_contains(line, "ssl_ciphers")
    not string_contains(line, "DES-CBC3-SHA")
    not string_contains(line, "RC4-MD5")
    not string_contains(line, "RC4-SHA")
}

allowed_server_names {
    some line in input.lines
    contains(line, "server_name")
    server_name := trim(split(line, "server_name")[1])
    some allowed_name in data.allowed_server_names
    server_name == allowed_name
}

# Helper function to trim whitespace
trim(s) = t {
    t := regex.replace(s, "^\\s+|\\s+$", "")
}
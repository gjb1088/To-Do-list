module github.com/gjb1088/To-Do-list

go 1.18

require (
    github.com/gorilla/sessions v1.3.0
    golang.org/x/crypto v0.0.0
)

replace github.com/gorilla/sessions => ./stubs/gorilla/sessions
replace golang.org/x/crypto => ./stubs/xcrypto

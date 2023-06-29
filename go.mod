module GoShell

go 1.20

replace github.com/leaanthony/winc => ./winc

require (
	github.com/leaanthony/winc v0.0.0-20220323084916-ea5df694ec1f
	github.com/parsiya/golnk v0.0.0-20221103095132-740a4c27c4ff
	golang.org/x/sys v0.9.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
)

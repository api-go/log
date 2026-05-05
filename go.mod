module apigo.cc/go/log

go 1.25.0

require (
	apigo.cc/go/cast v1.2.6
	apigo.cc/go/config v1.0.5
	apigo.cc/go/shell v1.0.4
)

require (
	apigo.cc/go/encoding v1.0.4 // indirect
	apigo.cc/go/file v1.0.5 // indirect
	apigo.cc/go/rand v1.0.5 // indirect
	apigo.cc/go/safe v1.0.4 // indirect
	golang.org/x/crypto v0.50.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace apigo.cc/go/cast => ../cast

replace apigo.cc/go/config => ../config

replace apigo.cc/go/shell => ../shell

replace apigo.cc/go/file => ../file

replace apigo.cc/go/encoding => ../encoding

replace apigo.cc/go/safe => ../safe

replace apigo.cc/go/rand => ../rand

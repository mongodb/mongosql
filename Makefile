SUITES = integration

default: generate test 

run:
	go run ./main/sqlproxy.go --schema ./testdata/resources/schema -vvv

generate:
	go run ./testdata/bin/generate.go -suites "$(SUITES)"

test:
	go test . ./catalog ./collation ./evaluator ./internal/config ./mongodb ./parser ./schema ./server ./variable

test-restore-data:
	go test -restoreData "$(SUITES)"
  
shell:
	mysql -P3307
  
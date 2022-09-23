TEST_MYSQL=magicpool_pool_test_mysql
TEST_REDIS=magicpool_pool_test_redis

reset-test-containers:
	docker kill $(TEST_MYSQL) $(TEST_REDIS) || true
	docker rm -f $(TEST_MYSQL) $(TEST_REDIS) || true
	docker run --rm --name $(TEST_MYSQL) -p 3544:3306 -e MYSQL_ROOT_PASSWORD=secret -e MYSQL_DATABASE=pooldb -d mysql:8.0
	docker run --rm --name $(TEST_REDIS) -p 3545:6379 -d redis:6-alpine

fmt:
	go fmt ./... ./tests

unit:
	go test ./...

mod-tidy:
	go get github.com/btcsuite/btcd; go get github.com/btcsuite/btcd/btcec/v2; go get github.com/ethereum/go-ethereum; go mod tidy

integration:
	make reset-test-containers
	go test ./... -p 1 -tags=integration
	docker rm -f $(TEST_MYSQL) $(TEST_REDIS)

pool:
	go build -o magicpool-pool ./svc/pool

worker:
	go build -o magicpool-worker ./svc/worker

api:
	go build -o magicpool-api ./svc/api

keygen:
	go build -o magicpool-keygen ./cmd/keygen

loadtest:
	go build -o magicpool-loadtest ./cmd/loadtest

clean:
	rm -rf magicpool-pool magicpool-worker magicpool-api magicpool-keygen magicpool-loadtest
	docker rm -f $(TEST_MYSQL) $(TEST_REDIS)

.PHONY: reset-test-containers fmt unit integration pool worker api keygen loadtest clean

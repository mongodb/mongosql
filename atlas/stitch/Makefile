
clean:
	rm -rf bic-metrics
	git checkout HEAD -- bic-metrics

verify-committed:
	git add bic-metrics
	git diff-index --cached --quiet HEAD -- bic-metrics

download: verify-committed
	rm -rf bic-metrics
	stitch-cli export --app-id bic-metrics-fsdyb --output bic-metrics --include-hosting

upload:
	stitch-cli import --path bic-metrics --strategy=merge --include-hosting --reset-cdn-cache

deploy: upload clean download

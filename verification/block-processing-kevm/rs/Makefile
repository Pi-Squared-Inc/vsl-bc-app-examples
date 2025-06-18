.PHONY: unit-test
unit-test:
	@echo "Run unit tests" && \
		docker build --build-arg GITHUB_ACTOR=${GITHUB_ACTOR} --build-arg GITHUB_TOKEN=${GITHUB_TOKEN} --build-context root=../../../ -t kevm-test -f Dockerfile.test . && \
		echo "Unit tests completed successfully"

FROM ubuntu:latest

RUN apt-get update && apt-get install -y jq curl git \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /opt/tester
COPY ./integration-test.sh .
COPY ./check-test-result-geth.sh .
COPY ./check-test-result-reth.sh .
COPY ./check-test-result-btc.sh .

CMD [ "./integration-test.sh" ]
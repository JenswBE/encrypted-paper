# This script is intended to be sourced like "source lint.sh" or ". lint.sh".
golangci-lint run \
    --config <(curl --silent https://raw.githubusercontent.com/JenswBE/setup/main/programming_configs/golang/.golangci.yml) \
    --disable gochecknoinits,err113,forbidigo,wrapcheck

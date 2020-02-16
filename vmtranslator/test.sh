#!/usr/bin/env bash
set -eu

main() {
  echo "Running unit tests..."
	go test -race ./... "$@"

	# Print blank line
	echo

  echo "Running E2E tests..."
  e2e_test
}

e2e_test() {
  local test_dirs=$(find .. -type d -regex '../projects/0[78]/.*/.*' | grep -v FibonacciElement)

  for dir in ${test_dirs[@]}; do
		if [[ "$dir" == *"FibonacciElement"* ]]; then
			e2e_test_one "$dir" "true"
		else
			e2e_test_one "$dir" "false"
		fi
  done
}

e2e_test_one() {
	local dir="$1"
	local bootstrap="$2"

	echo "$dir:"
	echo "  ==> Compiling $dir..."
	echo -n "  ==> "
	go run main.go -bootstrap="$bootstrap" "$dir"
	local test="$(find $dir -regex '.*[^(VME)]\.tst')"
  echo "  ==> Running $test..."
	echo -n "  ==> "
	../tools/CPUEmulator.sh $test
}

main "$@"

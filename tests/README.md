# run all tests
./run_tests.sh

# run a test in container using dlv
docker run --rm -e CGO_ENABLED=0 --security-opt="apparmor=unconfined" --cap-add=SYS_PTRACE -p 40000:40000 -v $PWD:/app -it envy-${distro} dlv test --listen=:40000 --headless=true --api-version=2 --accept-multiclient <location of test file>
docker run --rm -e CGO_ENABLED=0 --security-opt="apparmor=unconfined" --cap-add=SYS_PTRACE -p 40000:40000 -v $PWD:/app -it envy-alpine dlv test --listen=:40000 --headless=true --api-version=2 --accept-multiclient ./tests/*.go -- -test.run ^TestTaskInstallsDepsCorrectly$ 

# run cmd in container using dlv
docker run --rm --security-opt="apparmor=unconfined" --cap-add=SYS_PTRACE -p 40000:40000 -v $PWD:/app -it envy-${distro} dlv debug --listen=:40000 --headless=true --api-version=2 --accept-multiclient main.go
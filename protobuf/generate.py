from subprocess import run, CalledProcessError
import os

PROCESS = 'protoc'
ARGUMENTS = [
    "--go_out=.",
    "--go_opt=paths=source_relative",
    "--go-grpc_out=.",
    "--go-grpc_opt=paths=source_relative",
]
DEPS = [
    "google.golang.org/protobuf/cmd/protoc-gen-go@v1.28",
    "google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2"
]

def main(files: list[str]) -> None:
    for file in files:
        print(f"Generating protobuf files for {file}")
        proc = run([PROCESS, *ARGUMENTS, f"{file}"])
        try:
            proc.check_returncode()
        except CalledProcessError as e:
            print(f"Error: {e}")
            os._exit(1)


def install_deps() -> None:
    for dep in DEPS:
        run(["go", "install", dep])

def files() -> list[str]:
    files = [file for file in os.listdir() if file.endswith(".proto")]
    return files

if __name__ == '__main__':
    install_deps()
    main(files())

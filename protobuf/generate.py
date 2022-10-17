# Generate protobuf files by running protoc

# import function that runs protoc
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

def main(folders: list[str]) -> None:
    for folder in folders:
        print(f"Generating protobuf files for {folder}")
        proc = run([PROCESS, *ARGUMENTS, f"{folder}/{folder}.proto"])
        try:
            proc.check_returncode()
        except CalledProcessError as e:
            print(f"Error: {e}")
            os.exit(1)


def install_deps() -> None:
    for dep in DEPS:
        run(["go", "install", dep])

def folders() -> list[str]:
    folders = os.listdir('.')
    folders = [folder for folder in folders if os.path.isdir(folder)]
    
    return folders

if __name__ == '__main__':
    install_deps()
    main(folders())

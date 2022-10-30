import { promisify } from "node:util";
import { exec } from "node:child_process";
import { readdir } from "node:fs/promises";

enum Language {
  TS = "ts",
  GO = "go",
}

type IDeps = {
  [key in Language]: string[];
};

const isWindows = () => (process.platform === "win32" ? ".ps1" : "");

const run = promisify(exec);

const PROCESS = "protoc";
const ARGUMENTS = [
  "--go_out=.",
  "--go_opt=paths=source_relative",
  "--go-grpc_out=.",
  "--go-grpc_opt=paths=source_relative",
  `--plugin=./node_modules/.bin/protoc-gen-ts_proto${isWindows()} --ts_proto_out=.`,
];
const DEPS: IDeps = {
  go: [
    "google.golang.org/protobuf/cmd/protoc-gen-go@v1.28",
    "google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2",
  ],
  ts: ["ts-proto"],
};

type ProcStep = { [key in Language]: (dep: string) => ProcOut };
type ProcOut = Promise<{ stdout: string; stderr: string }>;

const INSTALL_PRG: ProcStep = {
  ts: async (dep) => run(`npm install ${dep}`),
  go: async (dep) => run(`go install ${dep}`),
};

(async () => {
  for (const [typ, l] of Object.entries(DEPS)) {
    for (const dep of l) {
      console.log(`Installing ${typ} - ${dep}`);
      await INSTALL_PRG[typ as Language](dep);
    }
  }

  const files = await (
    await readdir(process.cwd())
  ).filter((file) => file.endsWith(".proto"));

  for (const file of files) {
    console.log(`Generating protobuf files for ${file}`);
    const cmd = `${PROCESS} ${ARGUMENTS.join(" ")} ${file}`;
    console.log(cmd);
    await run(cmd);
  }

  console.log("Done!");
})();

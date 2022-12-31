import { sh } from "https://deno.land/x/drake@v1.6.0/mod.ts";

const GOPROCESS = "protoc";
const TSPROCESS = `${Deno.env.get("PB_BINARY") ?? "pb"} gen ts`;

const ARGUMENTS: string[] = [
  "--go_out=.",
  "--go_opt=paths=source_relative",
  "--go-grpc_out=.",
  "--go-grpc_opt=paths=source_relative",
];

const DEPS: string[] = [
  "google.golang.org/protobuf/cmd/protoc-gen-go@v1.28",
  "google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2",
];

(async () => {
  for (const dep of DEPS) {
    console.log(`Installing - ${dep}`);
    await sh(`go install ${dep}`);
  }

  for (const { name } of Deno.readDirSync(Deno.cwd())) {
    if (!name.endsWith(".proto")) continue;

    console.log(`Generating protobuf files for ${name}`);
    const cmd = `${GOPROCESS} ${ARGUMENTS.join(" ")} ${name}`;
    console.log("Running", cmd);
    await sh(cmd);
  }

  console.log("Running", TSPROCESS);
  await sh(`${TSPROCESS} --entry-path . `);
  console.log('Typescript definitions generated to "out" folder!');

  console.log("Done!");
})();

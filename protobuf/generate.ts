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

const moveDir = async (src: string, dst: string) => {
  const Remove = (path: string) => Deno.remove(path, { recursive: true });

  const exists = await Deno.stat(dst)
    .then(() => true)
    .catch(() => false);

  if (exists) await Remove(dst);
  //   if (exists) await Deno.remove(dst, { recursive: true });

  await Deno.mkdir(dst, { recursive: true });

  const files = Deno.readDirSync(dst);

  for (const file of files) {
    const filePath = src + "/" + file.name;
    const newFilePath = dst + "/" + file.name;

    await Deno.rename(filePath, newFilePath);
  }

  await Remove(src);
};

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

  console.log("Moving files to markov-generator");

  const outDir = Deno.cwd() + "/out";
  const markovDir = Deno.cwd() + "/../markov-generator/src/protobuf";

  await moveDir(outDir, markovDir);

  console.log("Done!");
})();

import { EnvironementVariableSchemaType } from "~/api/services/schema/services";
import { PatchSchemaType } from "~/pages/namespace/Explorer/Service/ServiceEditor/schema";

const createPatchesYaml = (patches: PatchSchemaType[]) =>
  patches
    .map(
      (item) =>
        `\n  - op: "${item.op}"\n    path: "${item.path}"\n    value: "${item.value}"`
    )
    .join("");

const createEnvsYaml = (envs: EnvironementVariableSchemaType[]) =>
  envs
    .map((item) => `\n  - name: "${item.name}"\n    value: "${item.value}"`)
    .join("");

type Service = {
  image: string;
  scale: number;
  size: string;
  cmd: string;
  patches: PatchSchemaType[];
  envs: EnvironementVariableSchemaType[];
};

export const createServiceYaml = ({
  image,
  scale,
  size,
  cmd,
  patches,
  envs,
}: Service) => `direktiv_api: "service/v1"
image: "${image}"
scale: ${scale}
size: "${size}"
cmd: "${cmd}"
patches:${createPatchesYaml(patches)}
envs:${createEnvsYaml(envs)}`;

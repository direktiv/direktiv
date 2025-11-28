import { EnvVarSchemaType } from "~/api/services/schema/services";
import { PatchSchemaType } from "~/pages/namespace/Explorer/Service/ServiceEditor/schema";
import { createFile } from "e2e/utils/files";

const createPatchesYaml = (patches?: PatchSchemaType[]) =>
  patches
    ? patches
        .map(
          (item) =>
            `\n  - op: ${item.op}\n    path: ${item.path}\n    value: ${item.value}`
        )
        .join("")
    : "[]";

const createEnvsYaml = (envs?: EnvVarSchemaType[]) =>
  envs
    ? envs
        .map((item) => `\n  - name: ${item.name}\n    value: ${item.value}`)
        .join("")
    : "[]";

type Service = {
  name: string;
  image: string;
  scale: number;
  size: string;
  cmd: string;
  patches?: PatchSchemaType[];
  envs?: EnvVarSchemaType[];
};

export const createServiceYaml = ({
  image,
  scale,
  size,
  cmd,
  patches,
  envs,
}: Service) => {
  // Build the YAML lines individually
  const lines: string[] = [
    "direktiv_api: service/v1",
    `image: ${image}`,
    `scale: ${scale}`,
    `size: ${size}`,
    `cmd: ${cmd}`,
  ];

  if (patches && patches.length > 0) {
    lines.push(`patches:${createPatchesYaml(patches)}`);
  }

  if (envs && envs.length > 0) {
    lines.push(`envs:${createEnvsYaml(envs)}`);
  }

  // Join everything with newlines
  return lines.join("\n");
};

export const createService = async (namespace: string, service: Service) => {
  const content = createServiceYaml(service);

  await createFile({
    name: service.name,
    namespace,
    type: "service",
    mimeType: "application/json",
    content,
  });
};

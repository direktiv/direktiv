import { EnvVarSchemaType } from "~/api/services/schema/services";
import { PatchSchemaType } from "~/pages/namespace/Explorer/Service/ServiceEditor/schema";
import { createWorkflow } from "~/api/tree/mutate/createWorkflow";
import { headers } from "e2e/utils/testutils";

const createPatchesYaml = (patches?: PatchSchemaType[]) =>
  patches
    ? patches
        .map(
          (item) =>
            `\n  - op: "${item.op}"\n    path: "${item.path}"\n    value: "${item.value}"`
        )
        .join("")
    : "[]";

const createEnvsYaml = (envs?: EnvVarSchemaType[]) =>
  envs
    ? envs
        .map((item) => `\n  - name: "${item.name}"\n    value: "${item.value}"`)
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
}: Service) => `direktiv_api: "service/v1"
image: "${image}"
scale: ${scale}
size: "${size}"
cmd: "${cmd}"
patches: ${createPatchesYaml(patches)}
envs: ${createEnvsYaml(envs)}`;

export const createService = async (namespace: string, service: Service) => {
  const payload = createServiceYaml(service);

  await createWorkflow({
    payload,
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name: service.name,
    },
    headers,
  });
};

import { InstancesInputSchema } from "../schema";
import { apiFactory } from "~/api/apiFactory";

export const getInput = apiFactory({
  url: ({
    namespace,
    baseUrl,
    instanceId,
  }: {
    baseUrl?: string;
    namespace: string;
    instanceId: string;
  }) =>
    `${
      baseUrl ?? ""
    }/api/namespaces/${namespace}/instances/${instanceId}/input`,
  method: "GET",
  schema: InstancesInputSchema,
});

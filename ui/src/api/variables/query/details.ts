import { VarContentSchema } from "../schema";
import { apiFactory } from "~/api/apiFactory";

export const getVariableContent = apiFactory({
  url: ({ namespace, variableID }: { namespace: string; variableID: string }) =>
    `/api/v2/namespaces/${namespace}/variables/${variableID}`,
  method: "GET",
  schema: VarContentSchema,
});

export type VarContentType = Awaited<ReturnType<typeof getVariableContent>>;

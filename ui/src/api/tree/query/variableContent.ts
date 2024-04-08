import type { QueryFunctionContext } from "@tanstack/react-query";
import type { ResponseParser } from "../../apiFactory";
import { WorkflowVariableContentSchema } from "../schema/obsoleteWorkflowVariable";
import { apiFactory } from "../../apiFactory";
import { forceLeadingSlash } from "~/api/files/utils";
import { treeKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useQuery } from "@tanstack/react-query";

const variableResponseParser: ResponseParser = async ({ res, schema }) => {
  // Different from the default responseParser, for varibles we
  // don't want to try to parse the response as JSON because
  // it will always be treated as text to be displayed or edited
  // (even if it is JSON).
  const textResult = await res.text();
  if (!textResult) return schema.parse(null);

  // If response is not null, return it as 'body',
  // and also add the response headers (content type is needed later)
  // This is a workaround, in the new API we should return the
  // content type as part of the JSON response.
  const headers = Object.fromEntries(res.headers);
  return schema.parse({ body: textResult, headers });
};

const getVariableContent = apiFactory({
  url: ({
    namespace,
    path,
    name,
  }: {
    namespace: string;
    path: string;
    name: string;
  }) =>
    `/api/namespaces/${namespace}/tree${forceLeadingSlash(
      path
    )}?op=var&var=${name}`,
  method: "GET",
  schema: WorkflowVariableContentSchema,
  responseParser: variableResponseParser,
});

const fetchVariableContent = async ({
  queryKey: [{ name, apiKey, namespace, path }],
}: QueryFunctionContext<
  ReturnType<(typeof treeKeys)["workflowVariableContent"]>
>) =>
  getVariableContent({
    apiKey,
    urlParams: { namespace, path, name },
  });

export const useWorkflowVariableContent = (name: string, path: string) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery({
    queryKey: treeKeys.workflowVariableContent(namespace, {
      apiKey: apiKey ?? undefined,
      name,
      path,
    }),
    queryFn: fetchVariableContent,
    enabled: !!namespace,
  });
};

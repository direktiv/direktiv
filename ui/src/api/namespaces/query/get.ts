import {
  NamespaceListSchema,
  NamespaceListSchemaType,
} from "../schema/namespace";

import type { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { namespaceKeys } from "..";
import { sortByName } from "~/api/files/utils";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useQuery } from "@tanstack/react-query";

export const getNamespaces = apiFactory({
  url: ({ baseUrl }: { baseUrl?: string }) =>
    `${baseUrl ?? ""}/api/v2/namespaces`,
  method: "GET",
  schema: NamespaceListSchema,
});

const fetchNamespaces = async ({
  queryKey: [{ apiKey }],
}: QueryFunctionContext<ReturnType<(typeof namespaceKeys)["all"]>>) =>
  getNamespaces({
    apiKey,
    urlParams: {},
  });

const useNamespaces = <T>({
  filter,
}: {
  filter: (apiResponse: NamespaceListSchemaType) => T;
}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery({
    queryKey: namespaceKeys.all(apiKey ?? undefined),
    queryFn: fetchNamespaces,
    enabled: !!namespace,
    select: (data) => filter(data),
  });
};

export const useListNamespaces = () =>
  useNamespaces({
    filter: (apiResponse) => ({ data: apiResponse.data.sort(sortByName) }),
  });

export const useNamespaceDetail = () => {
  const namespace = useNamespace();

  return useNamespaces({
    filter: (apiResponse) =>
      apiResponse.data.find((namespaceObj) => namespaceObj.name === namespace),
  });
};

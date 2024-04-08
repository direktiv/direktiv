import { NamespaceListSchema } from "../schema";
import type { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { namespaceKeys } from "../";
import { sortByName } from "~/api/files/utils";
import { useApiKey } from "~/util/store/apiKey";
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

export const useListNamespaces = () => {
  const apiKey = useApiKey();

  return useQuery({
    queryKey: namespaceKeys.all(apiKey ?? undefined),
    queryFn: fetchNamespaces,
    select(data) {
      if (data) {
        return {
          data: data.data.sort(sortByName),
        };
      }
      return data;
    },
  });
};

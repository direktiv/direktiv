import { NamespaceListSchema, NamespaceListSchemaType } from "../schema";

import type { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { namespaceKeys } from "../";
import { sortByName } from "~/api/tree/utils";
import { useApiKey } from "~/util/store/apiKey";
import { useQuery } from "@tanstack/react-query";

type newType = {
  name: string;
};

export const getNamespaces = apiFactory({
  url: ({ baseUrl }: { baseUrl?: string }) => `${baseUrl ?? ""}/api/namespaces`,
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
  });
};

export const useSortedNamespaces = () => {
  const { data } = useListNamespaces();

  if (data == undefined) return null;

  const list = data.results;

  return { data, results: data.results.sort(sortByName) };
};

export const useSorted = () => {
  const data = {
    results: [
      { name: "zebra", other: "abc" },
      { name: "apple", other: "cde" },
      { name: "okay" },
    ],
  };

  return { results: data.results.sort(sortByName) };
};

export const useSort = () => {
  const apple: newType = {
    name: "apple",
  };

  const banana: newType = {
    name: "banana",
  };

  const colibri: newType = {
    name: "colibri",
  };

  const zebra: newType = {
    name: "zebra",
  };

  // const results = [zebra, apple].sort(sortByName);

  // return results;

  return { results: [zebra, apple, colibri, banana].sort(sortByName) };
};

//         results: [...oldResults, data.results[0]].sort(sortByName),

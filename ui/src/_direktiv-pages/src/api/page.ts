import { queryOptions, useQuery } from "@tanstack/react-query";

import { apiFactory } from "~/api/apiFactory";
import { removeTrailingSlash } from "~/api/files/utils";
import z from "zod";

const getPage = apiFactory({
  url: ({ path }: { path: string }) => `${removeTrailingSlash(path)}/page.json`,
  method: "GET",
  schema: z.unknown(),
});

const pageQueryOptions = (path: string) =>
  queryOptions({
    queryKey: ["page", path],
    queryFn: () => getPage({ urlParams: { path } }),
  });

export const usePage = (path: string) => useQuery(pageQueryOptions(path));

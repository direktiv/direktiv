import { queryOptions, useQuery } from "@tanstack/react-query";

import { DirektivPagesSchema } from "~/pages/namespace/Explorer/Page/poc/schema";
import { apiFactory } from "~/api/apiFactory";
import { removeTrailingSlash } from "~/api/files/utils";

const getPage = apiFactory({
  url: ({ path }: { path: string }) => `${path}/page.json`,
  method: "GET",
  schema: DirektivPagesSchema,
});

const pageQueryOptions = (path: string) =>
  queryOptions({
    queryKey: ["page", removeTrailingSlash(path)],
    queryFn: () => getPage({ urlParams: { path } }),
  });

export const usePage = (path: string) => useQuery(pageQueryOptions(path));

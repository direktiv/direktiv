import { QueryFunctionContext, useQuery } from "@tanstack/react-query";

import { DirektivPagesSchema } from "~/pages/namespace/Explorer/Page/poc/schema";
import { apiFactory } from "~/api/apiFactory";
import { removeTrailingSlash } from "~/api/files/utils";

export const getPage = apiFactory({
  url: ({ path }: { path: string }) => `${path}/page.json`,
  method: "GET",
  schema: DirektivPagesSchema,
});

export const pageKeys = {
  page: (path: string) => [{ scope: "page", path }] as const,
};

const fetchPage = async ({
  queryKey: [{ path }],
}: QueryFunctionContext<ReturnType<(typeof pageKeys)["page"]>>) =>
  getPage({ urlParams: { path } });

export const usePage = (path: string) =>
  useQuery({
    queryKey: pageKeys.page(removeTrailingSlash(path)),
    queryFn: fetchPage,
  });

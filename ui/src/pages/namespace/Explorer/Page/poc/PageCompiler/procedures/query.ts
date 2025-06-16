import { queryOptions, useSuspenseQueries } from "@tanstack/react-query";

import { QueryType } from "../../schema/procedures/query";
import { useTranslation } from "react-i18next";

export const usePageSuspenseQueries = (queries: QueryType[]) => {
  const { t } = useTranslation();
  return useSuspenseQueries({
    queries: queries.map((query) => {
      // TODO: implement query params
      const { id, baseUrl, queryParams } = query;
      return queryOptions({
        queryKey: [id, baseUrl],
        queryFn: async () => {
          const response = await fetch(baseUrl);
          if (!response.ok) {
            throw new Error(
              t("direktivPage.error.query.queryFailed", {
                id,
                url: baseUrl,
                status: response.status,
              })
            );
          }
          try {
            return await response.json();
          } catch (e) {
            throw new Error(
              t("direktivPage.error.query.invalidJson", {
                id,
                url: baseUrl,
              })
            );
          }
        },
      });
    }),
  });
};

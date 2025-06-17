import { queryOptions, useSuspenseQueries } from "@tanstack/react-query";

import { QueryType } from "../../schema/procedures/query";
import { useTranslation } from "react-i18next";
import { useUrlGenerator } from "./utils";

export const usePageSuspenseQueries = (queries: QueryType[]) => {
  const { t } = useTranslation();
  const getUrlFn = useUrlGenerator();
  return useSuspenseQueries({
    queries: queries.map((query) => {
      const url = getUrlFn(query);
      const { id } = query;
      return queryOptions({
        queryKey: [id, url],
        queryFn: async () => {
          const response = await fetch(url);
          if (!response.ok) {
            throw new Error(
              t("direktivPage.error.query.queryFailed", {
                id,
                url,
                status: response.status,
              })
            );
          }
          try {
            return await response.json();
          } catch (e) {
            throw new Error(
              t("direktivPage.error.query.invalidJson", { id, url })
            );
          }
        },
      });
    }),
  });
};

import { queryOptions, useSuspenseQueries } from "@tanstack/react-query";

import { QueryType } from "../../schema/procedures/query";
import { getUrl } from "./utils";
import { useTranslation } from "react-i18next";

export const usePageSuspenseQueries = (queries: QueryType[]) => {
  const { t } = useTranslation();
  return useSuspenseQueries({
    queries: queries.map((query) => {
      const { id } = query;
      const url = getUrl(query);
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

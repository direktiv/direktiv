import { JqQueryResult, JqQueryResultType } from "../schema";

import { apiFactory } from "~/api/apiFactory";
import { useApiKey } from "~/util/store/apiKey";
import { useMutation } from "@tanstack/react-query";

export const executeJquery = apiFactory({
  url: ({ baseUrl }: { baseUrl?: string }) => `${baseUrl ?? ""}/api/jq`,
  method: "POST",
  schema: JqQueryResult,
});

export const useExecuteJQuery = ({
  onSuccess,
}: {
  onSuccess?: (data: JqQueryResultType) => void;
} = {}) => {
  const apiKey = useApiKey();

  return useMutation({
    mutationFn: ({ query, inputJSON }: { query: string; inputJSON: string }) =>
      executeJquery({
        apiKey: apiKey ?? undefined,
        urlParams: {},
        payload: {
          query,
          data: btoa(inputJSON),
        },
      }),
    onSuccess: (res) => {
      onSuccess?.(res);
    },
    onError: () => {
      // TODO: handle the error in the UI
    },
  });
};

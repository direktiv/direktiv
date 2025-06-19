import {
  keyValueArrayToObject,
  useKeyValueArrayResolver,
} from "../primitives/keyValue/utils";
import { useMutation, useQueryClient } from "@tanstack/react-query";

import { MutationType } from "../../schema/procedures/mutation";
import { useUrlGenerator } from "./utils";

export const usePageMutation = (mutation: MutationType) => {
  const { method, requestBody, requestHeaders } = mutation;
  const queryClient = useQueryClient();

  const url = useUrlGenerator()(mutation);
  const resolveKeyValueArray = useKeyValueArrayResolver();

  const requestBodyResolved = resolveKeyValueArray(requestBody ?? []);
  const body = JSON.stringify(keyValueArrayToObject(requestBodyResolved));

  const requestHeadersResolved = resolveKeyValueArray(requestHeaders ?? []);
  const headers = keyValueArrayToObject(requestHeadersResolved);

  return useMutation({
    mutationFn: async () => {
      const response = await fetch(url, {
        method,
        body,
        headers,
      });
      if (!response.ok) {
        throw new Error(`${response.status}: ${response.statusText}`);
      }

      return;
    },
    onSuccess: () => {
      queryClient.invalidateQueries();
    },
  });
};

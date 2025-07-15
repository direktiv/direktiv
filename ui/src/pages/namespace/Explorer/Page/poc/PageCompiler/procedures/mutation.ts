import {
  keyValueArrayToObject,
  useKeyValueArrayResolver,
} from "../primitives/keyValue/utils";
import { useMutation, useQueryClient } from "@tanstack/react-query";

import { MutationType } from "../../schema/procedures/mutation";
import { useUrlGenerator } from "./utils";

export const usePageMutation = (mutation: MutationType) => {
  const { method, requestHeaders } = mutation;
  const queryClient = useQueryClient();
  const generateUrl = useUrlGenerator();
  const resolveKeyValueArray = useKeyValueArrayResolver();

  const url = generateUrl(mutation);

  // TODO: implement parsing the body with the new schema
  // const requestBodyResolved = resolveKeyValueArray(requestBody ?? []);
  // const body = JSON.stringify(keyValueArrayToObject(requestBodyResolved));

  const requestHeadersResolved = resolveKeyValueArray(requestHeaders ?? []);
  const headers = keyValueArrayToObject(requestHeadersResolved);

  return useMutation({
    mutationFn: async () => {
      const response = await fetch(url, {
        method,
        body: undefined,
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

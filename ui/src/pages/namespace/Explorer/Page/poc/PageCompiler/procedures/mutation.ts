import {
  keyValueArrayToObject,
  useKeyValueArrayResolver,
} from "../primitives/keyValue/utils";
import { useMutation, useQueryClient } from "@tanstack/react-query";

import { MutationType } from "../../schema/procedures/mutation";
import { useUrlGenerator } from "./utils";

export const usePageMutation = () => {
  const queryClient = useQueryClient();
  const generateUrl = useUrlGenerator();
  const resolveKeyValueArray = useKeyValueArrayResolver();

  return useMutation({
    mutationFn: async ({
      mutation,
      formData,
    }: {
      mutation: MutationType;
      formData: FormData;
    }) => {
      const { method, requestBody, requestHeaders } = mutation;
      const requestBodyResolved = resolveKeyValueArray(requestBody ?? [], {
        formData,
      });
      const body = JSON.stringify(keyValueArrayToObject(requestBodyResolved));

      const requestHeadersResolved = resolveKeyValueArray(
        requestHeaders ?? [],
        { formData }
      );
      const headers = keyValueArrayToObject(requestHeadersResolved);

      const url = generateUrl(mutation);

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

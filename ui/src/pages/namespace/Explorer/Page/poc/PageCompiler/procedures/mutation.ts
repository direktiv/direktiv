import {
  keyValueArrayToObject,
  useKeyValueArrayResolver,
} from "../primitives/keyValue/utils";

import { MutationType } from "../../schema/procedures/mutation";
import { useMutation } from "@tanstack/react-query";
import { useUrlGenerator } from "./utils";

export const usePageMutation = (mutation: MutationType) => {
  const { method, requestBody, requestHeaders } = mutation;

  const url = useUrlGenerator()(mutation);
  const resolveKeyValueArray = useKeyValueArrayResolver();

  const requestBodyResolved = resolveKeyValueArray(requestBody ?? []);
  const body = JSON.stringify(keyValueArrayToObject(requestBodyResolved));
  const headers = keyValueArrayToObject(requestHeaders ?? []);

  return useMutation({
    mutationFn: async () => {
      const response = await fetch(url, {
        method,
        body,
        headers,
      });
      if (!response.ok) {
        throw new Error("Something went wrong.");
      }
      return await response.json();
    },
  });
};

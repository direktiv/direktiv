import { MutationType } from "../../schema/procedures/mutation";
import { useKeyValueArrayResolver } from "../primitives/keyValue/utils";
import { useMutation } from "@tanstack/react-query";
import { useUrlGenerator } from "./utils";

export const usePageMutation = (mutation: MutationType) => {
  const { method, requestBody, requestHeaders } = mutation;

  const url = useUrlGenerator()(mutation);
  const resolveKeyValueArray = useKeyValueArrayResolver();

  const requestBodyResolved = resolveKeyValueArray(requestBody ?? []);

  return useMutation({
    mutationFn: async () => {
      const response = await fetch(url, {
        method,
        body: JSON.stringify({ some: "JSON" }),
        headers: { "Content-Type": "application/json" },
      });
      if (!response.ok) {
        throw new Error("Something went wrong.");
      }
      return await response.json();
    },
  });
};

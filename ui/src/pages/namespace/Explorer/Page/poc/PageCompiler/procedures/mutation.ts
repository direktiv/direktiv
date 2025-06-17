import { MutationType } from "../../schema/procedures/mutation";
import { useGetUrl } from "./utils";
import { useMutation } from "@tanstack/react-query";

export const usePageMutation = (mutation: MutationType) => {
  const { method } = mutation;

  const url = useGetUrl()(mutation);

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

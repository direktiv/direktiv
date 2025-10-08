import { WorkflowValidationSchemaType } from "./schema";
import { useQuery } from "@tanstack/react-query";
import { validationKeys } from ".";

export const sha1 = async (data: string) => {
  const encoder = new TextEncoder();
  const hashBuffer = await crypto.subtle.digest("SHA-1", encoder.encode(data));
  const hashArray = Array.from(new Uint8Array(hashBuffer));
  const hashHex = hashArray
    .map((b) => b.toString(16).padStart(2, "0"))
    .join("");
  return hashHex;
};

export const useValidate = async ({ data }: { data: string }) => {
  const hash = await sha1(data);
  return useQuery<WorkflowValidationSchemaType>({
    queryKey: validationKeys.messagesList({
      hash,
    }),
    /**
     * This hook is only used to subscribe to the correct cache key. Data for this key
     * is currently added through mutations (subject to change when we get a dedicated
     * endpoint for ts-workflow validation).
     */
    enabled: false,
  });
};

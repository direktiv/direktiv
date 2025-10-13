import { MonacoMarkerSchemaType } from "./schema/markers";
import { useQuery } from "@tanstack/react-query";
import { validationKeys } from ".";

export const useValidate = ({ hash }: { hash: string | null }) =>
  useQuery<MonacoMarkerSchemaType | undefined>({
    queryKey: validationKeys.messagesList({
      hash: hash ?? "",
    }),
    queryFn: async () => {
      if (!hash) {
        return [];
      }
      return undefined;
    },
    enabled: false,
    /**
     * This hook is only used to subscribe to the correct cache key. Data for this key
     * is currently added through mutations (subject to change when we get a dedicated
     * endpoint for ts-workflow validation).
     */
  });

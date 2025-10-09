import { useEffect, useState } from "react";

import { MonacoMarkerSchemaType } from "./schema/markers";
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

export const useSha1Hash = (data: string) => {
  const [hash, setHash] = useState<string | null>(null);

  useEffect(() => {
    let canceled = false;

    (async () => {
      const result = await sha1(data);
      if (!canceled) setHash(result);
    })();

    return () => {
      canceled = true;
    };
  }, [data]);

  return hash;
};

export const useValidate = ({ hash }: { hash: string | null }) =>
  useQuery<MonacoMarkerSchemaType>({
    queryKey: validationKeys.messagesList({
      hash: hash ?? "",
    }),
    queryFn: async () => {
      if (!hash) {
        return [] as MonacoMarkerSchemaType;
      }
      return undefined as unknown as MonacoMarkerSchemaType;
    },
    enabled: false,
    /**
     * This hook is only used to subscribe to the correct cache key. Data for this key
     * is currently added through mutations (subject to change when we get a dedicated
     * endpoint for ts-workflow validation).
     */
  });

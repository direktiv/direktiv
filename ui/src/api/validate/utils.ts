import { useEffect, useState } from "react";

import { MonacoMarkerSchema } from "./schema/markers";
import { SaveFileResponseSchemaType } from "../files/schema";
import { decode } from "js-base64";
import { editor } from "monaco-editor";
import queryClient from "~/util/queryClient";
import { validationKeys } from ".";

const sha1 = async (data: string) => {
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

    sha1(data).then((result) => {
      if (!canceled) setHash(result);
    });

    return () => {
      canceled = true;
    };
  }, [data]);

  return hash;
};

export const updateValidationCache = async ({
  response,
}: {
  response: SaveFileResponseSchemaType;
}): Promise<void> => {
  const { errors, data } = response.data;
  if (data === undefined) {
    return;
  }
  const content = decode(data);
  const hash = await sha1(content);
  const markers = MonacoMarkerSchema.parse(errors);
  if (markers) {
    queryClient.setQueryData<editor.IMarkerData[]>(
      validationKeys.messagesList({
        hash,
      }),
      () => markers
    );
  }
};

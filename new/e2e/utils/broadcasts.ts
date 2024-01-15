import {
  BroadcastsSchemaKeys,
  BroadcastsSchemaType,
} from "~/api/broadcasts/schema";

import { headers } from "./testutils";
import { updateBroadcasts } from "~/api/broadcasts/mutate/updateBroadcasts";

export const createBroadcasts = async (namespace: string) => {
  const broadcasts: { [key: string]: boolean } = {};
  BroadcastsSchemaKeys.map((key: keyof BroadcastsSchemaType) => {
    broadcasts[key] = Math.random() > 0.5;
  });
  return await updateBroadcasts({
    payload: { broadcast: broadcasts as BroadcastsSchemaType },
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
    },
    headers: {
      ...headers,
      "content-type": "application/json",
    },
  });
};

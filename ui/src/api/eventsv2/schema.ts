import { z } from "zod";

/**
 * example event object
 * {
    "event": {
        "specversion": "1.0",
        "id": "3b5f436a-aae8-46cb-a67c-1e1edca4e2f6",
        "source": "https://github.com/cloudevents/spec/pull",
        "type": "com.github.pull.create",
        "subject": "123",
        "datacontenttype": "text/xml",
        "time": "2018-04-05T17:31:00Z",
        "data": "\u003cmuch wow=\"xml\"/\u003e",
        "comexampleextension1": "value",
        "comexampleothervalue": 5
    },
    "namespace": "3c33a775-b90f-4fbf-901f-3c9bd0cc68e5",
    "namespaceName": "sebxian",
    "receivedAt": "2024-04-29T09:26:32.212915Z",
    "serialID": 1
  }
 */

const EventObject = z.object({
  namespaceName: z.string(),
  receivedAt: z.string(),
  event: z.object({
    id: z.string(),
    type: z.string(),
    time: z.string(),
  }),
});

/**
 * example response
 * 
 * {
    "meta": {
        "previousPage": "2024-04-29T09:26:32.212915Z",
        "startingFrom": "2024-04-29T09:38:27.358208093Z"
    },
    "data": EventObject[]
  }
 */

export const MetaData = z.object({
  previousPage: z.string(),
  startingFrom: z.string(),
});

export const EventsListResponseSchema = z.object({
  meta: MetaData,
  data: z.array(EventObject),
});

export type EventsListResponseSchemaType = z.infer<
  typeof EventsListResponseSchema
>;

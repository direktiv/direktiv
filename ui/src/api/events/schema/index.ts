import { z } from "zod";

/**
 * example event object
 * {
    "event": {
      "specversion": "1.0",
      "id": "4",
      "source": "https://direktiv.io/test",
      "type": "testerDuplicate"
    },
    "namespace": "foo",
    "namespaceId": "3c33a775-b90f-4fbf-901f-3c9bd0cc68e5",
    "receivedAt": "2024-04-29T09:26:32.212915Z",
    "serialID": 1
  }
 */

/**
 * the declared properties are mandatory (cloudevent spec);
 * additional properties may exist.
 */
const EventSchema = z
  .object({
    id: z.string(),
    specversion: z.string(),
    source: z.string(),
    type: z.string(),
  })
  .passthrough();

export const EventListItem = z.object({
  namespace: z.string(), // currently namespaceName
  receivedAt: z.string(),
  event: EventSchema,
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

const MetaData = z.object({
  previousPage: z.string().nullable(),
  startingFrom: z.string().nullable(),
});

export const EventsListResponseSchema = z.object({
  meta: MetaData,
  data: z.array(EventListItem),
});

export const NewEventSchema = z.object({
  body: z.string(),
});

export const EventReplayedSchema = z.null();
export const EventCreatedSchema = z.null();

export type EventSchemaType = z.infer<typeof EventSchema>;
export type EventListItemType = z.infer<typeof EventListItem>;
export type EventsListResponseSchemaType = z.infer<
  typeof EventsListResponseSchema
>;
export type EventReplayedSchemaType = z.infer<typeof EventReplayedSchema>;
export type EventCreatedSchemaType = z.infer<typeof EventCreatedSchema>;
export type NewEventSchemaType = z.infer<typeof NewEventSchema>;

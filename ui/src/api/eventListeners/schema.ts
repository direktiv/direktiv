/**
 * example response
 * 
 * {
    "meta": {
        "previousPage": "2024-04-29T12:19:30.814328Z",
        "startingFrom": ""
    },
    "data": [
      {
        "id": "46cac590-c9c3-4d50-af49-352fb0257acf",
        "createdAt": "2024-04-29T12:19:30.814328Z",
        "updatedAt": "2024-04-29T12:19:30.814328Z",
        "namespace": "foo",
        "listeningForEventTypes": [
            "greetingcloudevent"
        ],
        "triggerType": "WaitSimple",
        "triggerInstance": "886bf89a-9ea0-4bea-b7fc-f29ac88c7e9f"
        "eventContextFilters": [
          {
              "type": "take.event.two",
              "context": {
                  "foo": "bar"
              }
          },
        ]
      }
    ]
  }
 */

import { z } from "zod";

const triggerTypes = z.enum([
  "StartAnd",
  "WaitAnd",
  "StartSimple",
  "WaitSimple",
  "StartOR",
  "WaitOR",
]);

const EventContextFiltersSchema = z.array(
  z.object({
    type: z.string(),
    context: z.record(z.string(), z.string()),
  })
);

const EventListenerSchema = z.object({
  id: z.string(),
  createdAt: z.string(),
  updatedAt: z.string(),
  namespace: z.string(),
  listeningForEventTypes: z.array(z.string()),
  triggerType: triggerTypes,
  triggerInstance: z.string().optional(), // instance id
  triggerWorkflow: z.string().optional(), // workflow path
  eventContextFilters: EventContextFiltersSchema,
});

export const EventListenersResponseSchema = z.object({
  meta: z.object({
    total: z.number(),
  }),
  data: z.array(EventListenerSchema),
});

export type EventListenerSchemaType = z.infer<typeof EventListenerSchema>;
export type EventListenersResponseSchemaType = z.infer<
  typeof EventListenersResponseSchema
>;

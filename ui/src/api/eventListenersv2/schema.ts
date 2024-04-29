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
        }
    ]
  }
 */

import { z } from "zod";

const EventListenerSchema = z.object({
  id: z.string(),
  createdAt: z.string(),
  updatedAt: z.string(),
  namespace: z.string(),
  listeningForEventTypes: z.array(z.string()),
  triggerType: z.string(), // enum?
  triggerInstance: z.string().optional(), // TBD: clarify
  triggerWorkflow: z.string().optional(), // TBD: clarify
});

export const EventListenerResponseSchema = z.object({
  meta: z.object({
    previousPage: z.string().nullable(), // Todo: should it be nullable?
    startingFrom: z.string().nullable(), // Todo: should it be nullable?
  }),
  data: z.array(EventListenerSchema),
});

export type EventListenerSchemaType = z.infer<typeof EventListenerSchema>;
export type EventListenerResponseSchemaType = z.infer<
  typeof EventListenerResponseSchema
>;

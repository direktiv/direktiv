import { PageinfoSchema } from "../schema";
import { z } from "zod";

/**
 * One item in example API response
 * {
      
      "updatedAt":  "2023-08-09T13:56:10.364435Z",
      "mode":  "simple",
      "events":  [
        {
          "type":  "greetingcloudevent",
          "filters":  {}
        }
      ],
      "createdAt":  "2023-08-09T13:56:10.364435Z"
    }
 */

const EventDefinition = z.object({
  type: z.string(),
  filters: z.object({}),
});

const EventListenerSchema = z.object({
  createdAt: z.string(),
  workflow: z.string(), // "/listener.yaml",
  instance: z.string(), // ""
  mode: z.string(), // "simple" - TODO: is this an ENUM?
  events: z.array(EventDefinition),
});

export const EventListenersListSchema = z.object({
  namespace: z.string(),
  pageInfo: PageinfoSchema,
  results: z.array(EventListenerSchema),
});

export type EventListenersListSchemaType = z.infer<
  typeof EventListenersListSchema
>;

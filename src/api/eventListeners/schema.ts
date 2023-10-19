import { PageinfoSchema } from "../schema";
import { z } from "zod";

/**
 * One item in example API response
  {
    "workflow":  "/listener.yaml",
    "instance":  "",
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

// either listener or instance is defined (an event listener either
// starts the workflow or resumes the specified instance)
const EventListenerSchema = z.object({
  createdAt: z.string(),
  workflow: z.string(),
  instance: z.string(),
  mode: z.string(),
  events: z.array(EventDefinition),
});

export const EventListenersListSchema = z.object({
  namespace: z.string(),
  pageInfo: PageinfoSchema,
  results: z.array(EventListenerSchema),
});

export type EventListenerSchemaType = z.infer<typeof EventListenerSchema>;
export type EventListenersListSchemaType = z.infer<
  typeof EventListenersListSchema
>;

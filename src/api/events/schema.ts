import { PageinfoSchema } from "../schema";
import { z } from "zod";

const EventSchema = z.object({
  receivedAt: z.string(),
  id: z.string(),
  source: z.string(), // "https://github.com/cloudevents/spec/pull"
  type: z.string(), // "com.github.pull.create"
  cloudevent: z.string(), // base64
});

export const EventsListSchema = z.object({
  namespace: z.string(),
  events: z.object({
    pageInfo: PageinfoSchema,
    results: z.array(EventSchema),
  }),
});

export type EventSchemaType = z.infer<typeof EventSchema>;

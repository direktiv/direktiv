import { z } from "zod";

/**
 * example response:
 * 
 * {
    "data": [
        {
            "type": "uninitialized_secrets",
            "issue": "secrets have not been initialized: [list of secrets]",
            "count": 9,
            "level": "warning"
        }
    ]
  }
 */

// "issue" prop is intended for human API users and not processed in the ui
const NotificationSchema = z.object({
  type: z.string(),
  count: z.number(),
  level: z.enum(["warning"]),
});

export const NotificationListSchema = z.object({
  data: z.array(NotificationSchema),
});

export type NotificationSchemaType = z.infer<typeof NotificationSchema>;

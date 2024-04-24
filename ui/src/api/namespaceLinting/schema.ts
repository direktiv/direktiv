import { z } from "zod";

/**
 * example response:
 * 
 * {
    "data": [
        {
            "type": "uninitialized_secrets",
            "issue": "secrets have not been initialized: [list of secrets]",
            "level": "warning"
        }
    ]
  }
 */

// "issue" prop is intended for human API users and not processed in the ui
const NotificationSchema = z.object({
  type: z.string(),
  level: z.enum(["warning"]),
});

export const NotificationListSchema = z.object({
  data: z.array(NotificationSchema),
});

export type NotificationSchemaType = z.infer<typeof NotificationSchema>;
export type NotificationListSchemaType = z.infer<typeof NotificationListSchema>;

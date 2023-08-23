import {
  MirrorActivityStatusSchemaType,
  MirrorActivityTypeSchemaType,
} from "~/api/tree/schema/mirror";

import Badge from "~/design/Badge";
import { ComponentProps } from "react";

type BadgeVariant = ComponentProps<typeof Badge>["variant"];

export const activityTypeToBadeVariant = (
  type: MirrorActivityTypeSchemaType
): BadgeVariant => {
  switch (type) {
    case "dry_run":
      return "secondary";
    case "init":
    case "sync":
      return "success";
    default:
      break;
  }
};

export const activityStatusToBadgeVariant = (
  status: MirrorActivityStatusSchemaType
): BadgeVariant => {
  switch (status) {
    case "cancelled":
    case "failed":
      return "destructive";
    case "complete":
      return "success";
    case "executing":
    case "pending":
      return "outline";
    default:
      break;
  }
};

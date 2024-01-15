import Badge, { BadgeProps } from "~/design/Badge";
import {
  MirrorActivityStatusSchemaType,
  MirrorActivityTypeSchemaType,
} from "~/api/tree/schema/mirror";

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

type StatusBadgeProps = {
  variant: BadgeProps["variant"];
  icon: BadgeProps["icon"];
};

export const activityStatusToBadgeProps = (
  status: MirrorActivityStatusSchemaType
): StatusBadgeProps => {
  switch (status) {
    case "cancelled":
    case "failed":
      return { variant: "destructive", icon: "failed" };
    case "complete":
      return { variant: "success", icon: "complete" };
    case "executing":
    case "pending":
      return { variant: "outline", icon: "pending" };
    default:
      return { variant: "outline", icon: undefined };
  }
};

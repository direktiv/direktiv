import { BadgeProps } from "~/design/Badge";
import { SyncStatusSchemaType } from "~/api/syncs/schema";

type StatusBadgeProps = {
  variant: BadgeProps["variant"];
  icon: BadgeProps["icon"];
};

export const activityStatusToBadgeProps = (
  status: SyncStatusSchemaType
): StatusBadgeProps => {
  switch (status) {
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

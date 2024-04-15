import { BadgeProps } from "~/design/Badge";
import { MirrorActivityStatusSchemaType } from "~/api/tree/schema/mirror";

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

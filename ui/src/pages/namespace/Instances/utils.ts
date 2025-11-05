import Badge from "~/design/Badge";
import { ComponentProps } from "react";
import { InstanceSchemaType } from "~/api/instances/schema";

type BadgeVariant = ComponentProps<typeof Badge>["variant"];
type InstanceStatus = InstanceSchemaType["status"];

export const statusToBadgeVariant = (status: InstanceStatus): BadgeVariant => {
  switch (status) {
    case "complete":
      return "success";
    case "cancelled":
    case "failed":
      return "destructive";
    case "pending":
    case "running":
      return undefined;
    default:
      break;
  }
};

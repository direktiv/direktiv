import Badge from "~/design/Badge";
import { ComponentProps } from "react";
import { StatusSchemaType } from "~/api/services/schema";

type BadgeVariant = ComponentProps<typeof Badge>["variant"];

export const statusToBadgeVariant = (
  status: StatusSchemaType
): BadgeVariant => {
  switch (status) {
    case "True":
      return "success";
    case "False":
      return "destructive";
    case "Unknown":
      return "outline";
    default:
      break;
  }
};

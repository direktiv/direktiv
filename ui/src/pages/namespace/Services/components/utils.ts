import Alert from "~/design/Alert";
import Badge from "~/design/Badge";
import { ComponentProps } from "react";
import { StatusSchemaType } from "~/api/services/schema/services";

type BadgeVariant = ComponentProps<typeof Badge>["variant"];
type BadgeIcon = ComponentProps<typeof Badge>["icon"];
type AlertVariant = ComponentProps<typeof Alert>["variant"];

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

export const statusToBadgeIcon = (status: StatusSchemaType): BadgeIcon => {
  switch (status) {
    case "True":
      return "complete";
    case "False":
      return "failed";
    case "Unknown":
      return undefined;
    default:
      break;
  }
};

export const statusToAlertVariant = (
  status: StatusSchemaType
): AlertVariant => {
  switch (status) {
    case "True":
      return "success";
    case "False":
      return "error";
    case "Unknown":
      return undefined;
    default:
      break;
  }
};

import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from "~/design/HoverCard";
import {
  statusToAlertVariant,
  statusToBadgeIcon,
  statusToBadgeVariant,
} from "./utils";

import Alert from "~/design/Alert";
import Badge from "~/design/Badge";
import { ComponentProps } from "react";
import { ConditionalWrapper } from "~/util/helpers";
import { StatusSchemaType } from "~/api/services/schema";

type BadgeProps = ComponentProps<typeof Badge>;

type StatusBadgeProps = BadgeProps & {
  title?: string;
  message?: string;
  status: StatusSchemaType;
};

export const StatusBadge = ({
  status,
  title,
  message,
  ...props
}: StatusBadgeProps) => (
  <ConditionalWrapper
    condition={!!title || !!message}
    wrapper={(children) => (
      <HoverCard>
        <HoverCardTrigger className="inline-flex">{children}</HoverCardTrigger>
        <HoverCardContent asChild noBackground className="">
          <Alert
            variant={statusToAlertVariant(status)}
            className="w-96 whitespace-pre-wrap break-all"
          >
            <span className="font-bold">{title}</span>
            <br />
            {message}
          </Alert>
        </HoverCardContent>
      </HoverCard>
    )}
  >
    <Badge
      variant={statusToBadgeVariant(status)}
      icon={statusToBadgeIcon(status)}
      className="inline-flex"
      {...props}
    />
  </ConditionalWrapper>
);

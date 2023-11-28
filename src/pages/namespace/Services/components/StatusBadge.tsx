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
import { StatusSchemaType } from "~/api/services/schema/services";

type BadgeProps = ComponentProps<typeof Badge>;

type StatusBadgeProps = BadgeProps & {
  message?: string;
  status: StatusSchemaType;
};

export const StatusBadge = ({
  status,
  message,
  ...props
}: StatusBadgeProps) => (
  <ConditionalWrapper
    condition={!!message}
    wrapper={(children) => (
      <HoverCard>
        <HoverCardTrigger className="inline-flex">{children}</HoverCardTrigger>
        <HoverCardContent asChild noBackground>
          <Alert
            variant={statusToAlertVariant(status)}
            className="w-96 whitespace-pre-wrap break-all"
          >
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

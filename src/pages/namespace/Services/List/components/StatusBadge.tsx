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
import { useTranslation } from "react-i18next";

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
        <HoverCardTrigger>{children}</HoverCardTrigger>
        <HoverCardContent asChild noBackground>
          <Alert variant={statusToAlertVariant(status)}>
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
      {...props}
    />
  </ConditionalWrapper>
);

import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from "~/design/HoverCard";
import { TableCell, TableRow } from "~/design/Table";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";

import Alert from "~/design/Alert";
import Badge from "~/design/Badge";
import { ConditionalWrapper } from "~/util/helpers";
import { InstanceSchemaType } from "~/api/instances/schema";
import TooltipCopyBadge from "~/design/TooltipCopyBadge";
import { statusToBadgeVariant } from "../../Instances/utils";
import { useTranslation } from "react-i18next";
import useUpdatedAt from "~/hooksNext/useUpdatedAt";

export const InstanceRow = ({ instance }: { instance: InstanceSchemaType }) => {
  const { t } = useTranslation();
  const updatedAt = useUpdatedAt(instance.updatedAt);
  return (
    <TooltipProvider>
      <TableRow>
        <TableCell>{instance.as}</TableCell>
        <TableCell>
          <TooltipCopyBadge value={instance.id} variant="outline">
            {instance.id.slice(0, 8)}
          </TooltipCopyBadge>
        </TableCell>
        <TableCell>
          <ConditionalWrapper
            condition={instance.status === "failed"}
            wrapper={(children) => (
              <HoverCard>
                <HoverCardTrigger>{children}</HoverCardTrigger>
                <HoverCardContent
                  asChild
                  noBackground
                  data-testid={`instance-row-state-error-tooltip-${instance.id}`}
                >
                  <Alert variant="error">
                    <span className="font-bold">{instance.errorCode}</span>
                    <br />
                    {instance.errorMessage}
                  </Alert>
                </HoverCardContent>
              </HoverCard>
            )}
          >
            <Badge
              data-testid={`instance-row-state-${instance.id}`}
              variant={statusToBadgeVariant(instance.status)}
              icon={instance.status}
            >
              {instance.status}
            </Badge>
          </ConditionalWrapper>
        </TableCell>
        <TableCell>
          <Tooltip>
            <TooltipTrigger
              data-testid={`instance-row-relative-updated-time-${instance.id}`}
            >
              {t("pages.instances.list.tableRow.realtiveTime", {
                relativeTime: updatedAt,
              })}
            </TooltipTrigger>
            <TooltipContent
              data-testid={`instance-row-absolute-updated-time-${instance.id}`}
            >
              {instance.updatedAt}
            </TooltipContent>
          </Tooltip>
        </TableCell>
      </TableRow>
    </TooltipProvider>
  );
};

import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from "~/design/HoverCard";
import { Link, useNavigate } from "react-router-dom";
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
import { FC } from "react";
import { InstanceSchemaType } from "~/api/instances/schema";
import TooltipCopyBadge from "~/design/TooltipCopyBadge";
import { pages } from "~/util/router/pages";
import { statusToBadgeVariant } from "../utils";
import { useTranslation } from "react-i18next";
import useUpdatedAt from "~/hooksNext/useUpdatedAt";

const InstanceTableRow: FC<{
  instance: InstanceSchemaType;
  namespace: string;
}> = ({ instance, namespace }) => {
  const [invoker, childInstance] = instance.invoker.split(":");
  const updatedAt = useUpdatedAt(instance.updatedAt);
  const createdAt = useUpdatedAt(instance.createdAt);
  const navigate = useNavigate();
  const { t } = useTranslation();
  const isChildInstance = invoker === "instance" && !!childInstance;

  return (
    <TooltipProvider>
      <TableRow
        data-testid={`instance-row-${instance.id}`}
        key={instance.id}
        onClick={() => {
          navigate(
            pages.instances.createHref({
              namespace,
              instance: instance.id,
            })
          );
        }}
        className="cursor-pointer"
      >
        <TableCell data-testid="instance-column-name">
          <Tooltip>
            <TooltipTrigger asChild>
              <Link
                onClick={(e) => {
                  e.stopPropagation(); // prevent the onClick on the row from firing when clicking the workflow link
                }}
                to={pages.explorer.createHref({
                  namespace,
                  path: instance.as,
                  subpage: "workflow",
                })}
                className="hover:underline"
              >
                {instance.as}
              </Link>
            </TooltipTrigger>
            <TooltipContent>
              {t("pages.instances.list.tableRow.openWorkflowTooltip", {
                name: instance.as,
              })}
            </TooltipContent>
          </Tooltip>
        </TableCell>
        <TableCell data-testid="instance-column-id">
          <TooltipCopyBadge value={instance.id} variant="outline">
            {instance.id.slice(0, 8)}
          </TooltipCopyBadge>
        </TableCell>
        <TableCell data-testid="instance-column-invoker">
          {isChildInstance ? (
            <TooltipCopyBadge value={childInstance} variant="outline">
              {invoker}
            </TooltipCopyBadge>
          ) : (
            <Badge data-testid="invoker-type-badge" variant="outline">
              {invoker}
            </Badge>
          )}
        </TableCell>
        <TableCell data-testid="instance-column-state">
          <ConditionalWrapper
            condition={instance.status === "failed"}
            wrapper={(children) => (
              <HoverCard>
                <HoverCardTrigger data-testid="tooltip-copy-trigger">
                  {children}
                </HoverCardTrigger>
                <HoverCardContent
                  asChild
                  noBackground
                  data-testid="tooltip-copy-content"
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
              variant={statusToBadgeVariant(instance.status)}
              icon={instance.status}
            >
              {instance.status}
            </Badge>
          </ConditionalWrapper>
        </TableCell>
        <TableCell data-testid="instance-column-created-time">
          <Tooltip>
            <TooltipTrigger data-testid="tooltip-trigger">
              {t("pages.instances.list.tableRow.realtiveTime", {
                relativeTime: createdAt,
              })}
            </TooltipTrigger>
            <TooltipContent data-testid="tooltip-content">
              {instance.createdAt}
            </TooltipContent>
          </Tooltip>
        </TableCell>
        <TableCell data-testid="instance-column-updated-time">
          <Tooltip>
            <TooltipTrigger data-testid="tooltip-trigger">
              {t("pages.instances.list.tableRow.realtiveTime", {
                relativeTime: updatedAt,
              })}
            </TooltipTrigger>
            <TooltipContent data-testid="tooltip-content">
              {instance.updatedAt}
            </TooltipContent>
          </Tooltip>
        </TableCell>
      </TableRow>
    </TooltipProvider>
  );
};

export default InstanceTableRow;

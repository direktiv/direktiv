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
import CopyButton from "~/design/CopyButton";
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
  const isChild = invoker === "instance" && !!childInstance;

  return (
    <TooltipProvider>
      <TableRow
        data-testid={`instance-row-wrap-${instance.id}`}
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
        <TableCell>
          <Tooltip>
            <TooltipTrigger
              data-testid={`instance-row-workflow-${instance.id}`}
            >
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
              {t("pages.instances.list.tableRow.openWorkflowTooltip")}
            </TooltipContent>
          </Tooltip>
        </TableCell>
        <TableCell>
          <TooltipCopyBadge value={instance.id} variant="outline">
            {instance.id.slice(0, 8)}
          </TooltipCopyBadge>
        </TableCell>
        <TableCell>
          <ConditionalWrapper
            condition={isChild}
            wrapper={(children) => (
              <Tooltip>
                <TooltipTrigger>{children}</TooltipTrigger>
                <TooltipContent className="flex gap-2 align-middle">
                  {childInstance}
                  <CopyButton
                    value={childInstance ?? ""}
                    buttonProps={{
                      size: "sm",
                      onClick: (e) => {
                        e.stopPropagation();
                      },
                    }}
                  />
                </TooltipContent>
              </Tooltip>
            )}
          >
            <Badge
              data-testid={`instance-row-invoker-${instance.id}`}
              variant="outline"
            >
              {invoker}
            </Badge>
          </ConditionalWrapper>
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
              data-testid={`instance-row-relative-created-time-${instance.id}`}
            >
              {t("pages.instances.list.tableRow.realtiveTime", {
                relativeTime: createdAt,
              })}
            </TooltipTrigger>
            <TooltipContent
              data-testid={`instance-row-absolute-created-time-${instance.id}`}
            >
              {instance.createdAt}
            </TooltipContent>
          </Tooltip>
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

export default InstanceTableRow;

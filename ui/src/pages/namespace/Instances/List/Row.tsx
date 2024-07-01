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
import { FC } from "react";
import { InstanceSchemaType } from "~/api/instances/schema";
import TooltipCopyBadge from "~/design/TooltipCopyBadge";
import { decode } from "js-base64";
import moment from "moment";
import { statusToBadgeVariant } from "../utils";
import { useNavigate } from "react-router-dom";
import { usePages } from "~/util/router/pages";
import { useTranslation } from "react-i18next";
import useUpdatedAt from "~/hooks/useUpdatedAt";

const InstanceTableRow: FC<{
  instance: InstanceSchemaType;
}> = ({ instance }) => {
  const pages = usePages();
  const [invoker, childInstance] = instance.invoker.split(":");
  const isValidDate = moment(instance.endedAt).isValid();
  const endedAt = useUpdatedAt(instance.endedAt);
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
              namespace: instance.namespace,
              instance: instance.id,
            })
          );
        }}
        className="cursor-pointer"
      >
        <TableCell data-testid="instance-column-name">
          {instance.path}
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
                    {instance.errorMessage && decode(instance.errorMessage)}
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
        <TableCell data-testid="instance-column-ended-time">
          <ConditionalWrapper
            condition={isValidDate}
            wrapper={(children) => (
              <Tooltip>
                <TooltipTrigger data-testid="tooltip-trigger">
                  {children}
                </TooltipTrigger>
                <TooltipContent data-testid="tooltip-content">
                  {instance.endedAt}
                </TooltipContent>
              </Tooltip>
            )}
          >
            <>
              {isValidDate ? (
                t("pages.instances.list.tableRow.realtiveTime", {
                  relativeTime: endedAt,
                })
              ) : (
                <span className="italic">
                  {t("pages.instances.list.tableRow.stillRunning")}
                </span>
              )}
            </>
          </ConditionalWrapper>
        </TableCell>
      </TableRow>
    </TooltipProvider>
  );
};

export default InstanceTableRow;

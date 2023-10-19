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
import { InstanceSchemaType } from "~/api/instances/schema";
import TooltipCopyBadge from "~/design/TooltipCopyBadge";
import { pages } from "~/util/router/pages";
import { statusToBadgeVariant } from "../../Instances/utils";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";
import useUpdatedAt from "~/hooksNext/useUpdatedAt";

export const InstanceRow = ({ instance }: { instance: InstanceSchemaType }) => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const updatedAt = useUpdatedAt(instance.updatedAt);
  const namespace = useNamespace();

  if (!namespace) return null;

  return (
    <TooltipProvider>
      <TableRow
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
        <TableCell className="grid pl-5">
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
                className="overflow-hidden text-ellipsis hover:underline md:w-auto"
              >
                {instance.as}
              </Link>
            </TooltipTrigger>
            <TooltipContent>
              {t("pages.monitoring.instances.openWorkflowTooltip", {
                name: instance.as,
              })}
            </TooltipContent>
          </Tooltip>
        </TableCell>
        <TableCell className="w-0">
          <TooltipCopyBadge value={instance.id} variant="outline">
            {instance.id.slice(0, 8)}
          </TooltipCopyBadge>
        </TableCell>
        <TableCell className="w-0">
          <ConditionalWrapper
            condition={instance.status === "failed"}
            wrapper={(children) => (
              <HoverCard>
                <HoverCardTrigger>{children}</HoverCardTrigger>
                <HoverCardContent asChild noBackground>
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
              <span className="max-lg:hidden ">{instance.status}</span>
            </Badge>
          </ConditionalWrapper>
        </TableCell>
        <TableCell className="w-0 pr-5">
          <Tooltip>
            <TooltipTrigger>
              {t("pages.instances.list.tableRow.realtiveTime", {
                relativeTime: updatedAt,
              })}
            </TooltipTrigger>
            <TooltipContent>{instance.updatedAt}</TooltipContent>
          </Tooltip>
        </TableCell>
      </TableRow>
    </TooltipProvider>
  );
};

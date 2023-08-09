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
import { Link } from "react-router-dom";
import TooltipCopyBadge from "~/design/TooltipCopyBadge";
import { pages } from "~/util/router/pages";
import { statusToBadgeVariant } from "../../Instances/utils";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";
import useUpdatedAt from "~/hooksNext/useUpdatedAt";

export const InstanceRow = ({ instance }: { instance: InstanceSchemaType }) => {
  const { t } = useTranslation();
  const updatedAt = useUpdatedAt(instance.updatedAt);
  const namespace = useNamespace();

  if (!namespace) return null;

  return (
    <TooltipProvider>
      <TableRow>
        <TableCell>
          <Tooltip>
            <TooltipTrigger>
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
              {t("pages.monitoring.instances.openWorkflowTooltip")}
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
              {instance.status}
            </Badge>
          </ConditionalWrapper>
        </TableCell>
        <TableCell>
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

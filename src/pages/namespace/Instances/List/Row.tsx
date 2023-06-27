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
import { pages } from "~/util/router/pages";
import { statusToBadgeVariant } from "./utils";
import { useTranslation } from "react-i18next";
import useUpdatedAt from "~/hooksNext/useUpdatedAt";

const InstanceTableRow: FC<{
  instance: InstanceSchemaType;
  namespace: string;
}> = ({ instance, namespace }) => {
  const [name, revision] = instance.as.split(":");
  const updatedAt = useUpdatedAt(instance.updatedAt);
  const createdAt = useUpdatedAt(instance.createdAt);
  const navigate = useNavigate();
  const { t } = useTranslation();

  const isLatestRevision = revision === "latest";

  return (
    <TooltipProvider>
      <TableRow
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
            <TooltipTrigger>
              <Link
                onClick={(e) => {
                  e.stopPropagation(); // prevent the onClick on the row from firing when clicking the workflow link
                }}
                to={pages.explorer.createHref({
                  namespace,
                  path: name,
                  subpage: isLatestRevision ? "workflow" : "workflow-revisions",
                  revision: isLatestRevision ? undefined : revision,
                })}
                className="hover:underline"
              >
                {name}
              </Link>
            </TooltipTrigger>
            <TooltipContent>
              {t("pages.instances.list.tableRow.openWorkflowTooltip")}
            </TooltipContent>
          </Tooltip>
        </TableCell>
        <TableCell className="w-32">
          <Badge variant="outline">{instance.id.slice(0, 8)}</Badge>
        </TableCell>
        <TableCell className="w-28">
          <Badge variant="outline">{revision}</Badge>
        </TableCell>
        <TableCell className="w-28">
          <Badge variant="outline">{instance.invoker}</Badge>
        </TableCell>
        <TableCell className="w-28">
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
            <Badge variant={statusToBadgeVariant(instance.status)}>
              {instance.status}
            </Badge>
          </ConditionalWrapper>
        </TableCell>
        <TableCell className="w-40">
          <Tooltip>
            <TooltipTrigger>
              {t("pages.instances.list.tableRow.realtiveTime", {
                relativeTime: createdAt,
              })}
            </TooltipTrigger>
            <TooltipContent>{instance.createdAt}</TooltipContent>
          </Tooltip>
        </TableCell>
        <TableCell className="w-40">
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

export default InstanceTableRow;

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
import moment from "moment";
import { statusToBadgeVariant } from "../../Instances/utils";
import { useNamespace } from "~/util/store/namespace";
import { useNavigate } from "react-router-dom";
import { usePages } from "~/util/router/pages";
import { useTranslation } from "react-i18next";
import useUpdatedAt from "~/hooks/useUpdatedAt";

export const InstanceRow = ({ instance }: { instance: InstanceSchemaType }) => {
  const pages = usePages();
  const { t } = useTranslation();
  const navigate = useNavigate();
  const isValidDate = moment(instance.endedAt).isValid();
  const endedAt = useUpdatedAt(instance.endedAt);
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
        <TableCell className="grid pl-5">{instance.path}</TableCell>
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
          <ConditionalWrapper
            condition={isValidDate}
            wrapper={(children) => (
              <Tooltip>
                <TooltipTrigger>{children}</TooltipTrigger>
                <TooltipContent>{instance.endedAt}</TooltipContent>
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

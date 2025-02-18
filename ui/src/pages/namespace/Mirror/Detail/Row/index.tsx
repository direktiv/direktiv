import { TableCell, TableRow } from "~/design/Table";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";

import Badge from "~/design/Badge";
import { SyncObjectSchemaType } from "~/api/syncs/schema";
import TooltipCopyBadge from "~/design/TooltipCopyBadge";
import { activityStatusToBadgeProps } from "../utils";
import { useNavigate } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";
import useUpdatedAt from "~/hooks/useUpdatedAt";

const Row = ({
  item,
  namespace,
}: {
  item: SyncObjectSchemaType;
  namespace: string;
}) => {
  const createdAt = useUpdatedAt(item.createdAt);

  const { t } = useTranslation();
  const navigate = useNavigate();

  const statusBadgeProps = activityStatusToBadgeProps(item.status);

  return (
    <TableRow
      data-testid="sync-row"
      onClick={() => {
        navigate({
          to: "/n/$namespace/mirror/logs/$sync",
          params: { namespace, sync: item.id },
        });
      }}
    >
      <TooltipProvider>
        <TableCell>
          <TooltipCopyBadge value={item.id} variant="outline">
            {item.id.slice(0, 8)}
          </TooltipCopyBadge>
        </TableCell>
        <TableCell>
          <Badge
            variant={statusBadgeProps.variant}
            icon={statusBadgeProps.icon}
          >
            {item.status}
          </Badge>
        </TableCell>
        <TableCell>
          <Tooltip>
            <TooltipTrigger data-testid="createdAt-relative">
              {t("pages.mirror.syncs.tableRow.realtiveTime", {
                relativeTime: createdAt,
              })}
            </TooltipTrigger>
            <TooltipContent data-testid="createdAt-full">
              {item.createdAt}
            </TooltipContent>
          </Tooltip>
        </TableCell>
      </TooltipProvider>
    </TableRow>
  );
};

export default Row;

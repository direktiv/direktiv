import { TableCell, TableRow } from "~/design/Table";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";
import {
  activityStatusToBadgeProps,
  activityTypeToBadeVariant,
} from "../utils";

import Badge from "~/design/Badge";
import { MirrorActivitySchemaType } from "~/api/tree/schema/mirror";
import TooltipCopyBadge from "~/design/TooltipCopyBadge";
import { pages } from "~/util/router/pages";
import { useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import useUpdatedAt from "~/hooks/useUpdatedAt";

const Row = ({
  item,
  namespace,
}: {
  item: MirrorActivitySchemaType;
  namespace: string;
}) => {
  const createdAt = useUpdatedAt(item.createdAt);

  const { t } = useTranslation();
  const navigate = useNavigate();

  const statusBadgeProps = activityStatusToBadgeProps(item.status);

  return (
    <TableRow
      onClick={() => {
        navigate(
          pages.mirror.createHref({
            namespace,
            activity: item.id,
          })
        );
      }}
    >
      <TooltipProvider>
        <TableCell>
          <TooltipCopyBadge value={item.id} variant="outline">
            {item.id.slice(0, 8)}
          </TooltipCopyBadge>
        </TableCell>
        <TableCell>
          <Badge variant={activityTypeToBadeVariant(item.type)}>
            {item.type}
          </Badge>
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
            <TooltipTrigger data-testid="activity-row-createdAt-relative">
              {t("pages.mirror.activities.tableRow.realtiveTime", {
                relativeTime: createdAt,
              })}
            </TooltipTrigger>
            <TooltipContent data-testid="activity-row-createdAt-full">
              {item.createdAt}
            </TooltipContent>
          </Tooltip>
        </TableCell>
      </TooltipProvider>
    </TableRow>
  );
};

export default Row;

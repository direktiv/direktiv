import { TableCell, TableRow } from "~/design/Table";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";

import type { MirrorActivitySchemaType } from "~/api/tree/schema";
import TooltipCopyBadge from "~/design/TooltipCopyBadge";
import { useTranslation } from "react-i18next";
import useUpdatedAt from "~/hooksNext/useUpdatedAt";

const Row = ({ item }: { item: MirrorActivitySchemaType }) => {
  const createdAt = useUpdatedAt(item.createdAt);

  const { t } = useTranslation();
  return (
    <TableRow>
      <TooltipProvider>
        <TableCell>{item.status}</TableCell>
        <TableCell>{item.type}</TableCell>
        <TableCell>
          <TooltipCopyBadge value={item.id} variant="outline">
            {item.id.slice(0, 8)}
          </TooltipCopyBadge>
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

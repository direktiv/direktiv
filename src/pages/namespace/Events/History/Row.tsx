import { TableCell, TableRow } from "~/design/Table";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";

import { EventSchemaType } from "~/api/events/schema";
import TooltipCopyBadge from "../../../../design/TooltipCopyBadge";
import { useTranslation } from "react-i18next";
import useUpdatedAt from "~/hooksNext/useUpdatedAt";

const Row = ({
  event,
  onClick,
}: {
  event: EventSchemaType;
  namespace: string;
  onClick: (value: EventSchemaType) => void;
}) => {
  const { t } = useTranslation();

  const receivedAt = useUpdatedAt(event.receivedAt);

  return (
    <TooltipProvider>
      <TableRow
        data-testid={`event-row-${event.id}`}
        onClick={() => onClick(event)}
      >
        <TableCell>{event.type}</TableCell>
        <TableCell>
          <TooltipCopyBadge value={event.id} variant="outline">
            {event.id.slice(0, 8)}
          </TooltipCopyBadge>
        </TableCell>
        <TableCell>{event.source}</TableCell>
        <TableCell>
          <Tooltip>
            <TooltipTrigger data-testid="receivedAt-tooltip-trigger">
              {t("pages.events.history.tableRow.realtiveTime", {
                relativeTime: receivedAt,
              })}
            </TooltipTrigger>
            <TooltipContent data-testid="receivedAt-tooltip-content">
              {event.receivedAt}
            </TooltipContent>
          </Tooltip>
        </TableCell>
      </TableRow>
    </TooltipProvider>
  );
};

export default Row;

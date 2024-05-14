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
import useUpdatedAt from "~/hooks/useUpdatedAt";

const Row = ({
  receivedAt: propReceivedAt,
  event,
  onClick,
}: {
  receivedAt: string;
  event: EventSchemaType;
  namespace: string;
  onClick: (value: EventSchemaType) => void;
}) => {
  const { t } = useTranslation();

  const receivedAt = useUpdatedAt(propReceivedAt);

  return (
    <TooltipProvider>
      <TableRow data-testid="event-row" onClick={() => onClick(event)}>
        <TableCell headers="event-type">{event.type}</TableCell>
        <TableCell headers="event-id">
          <TooltipCopyBadge value={event.id} variant="outline">
            {event.id.slice(0, 8)}
          </TooltipCopyBadge>
        </TableCell>
        <TableCell headers="event-source">{event.source}</TableCell>
        <TableCell headers="event-received-at">
          <Tooltip>
            <TooltipTrigger data-testid="receivedAt-tooltip-trigger">
              {t("pages.events.history.tableRow.realtiveTime", {
                relativeTime: receivedAt,
              })}
            </TooltipTrigger>
            <TooltipContent data-testid="receivedAt-tooltip-content">
              {receivedAt}
            </TooltipContent>
          </Tooltip>
        </TableCell>
      </TableRow>
    </TooltipProvider>
  );
};

export default Row;

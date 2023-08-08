import { TableCell, TableRow } from "~/design/Table";

import { EventSchemaType } from "~/api/events/schema";
import TooltipCopyBadge from "../../../../design/TooltipCopyBadge";
import { TooltipProvider } from "~/design/Tooltip";

const Row = ({
  event,
  namespace,
}: {
  event: EventSchemaType;
  namespace: string;
}) => (
  <TooltipProvider>
    <TableRow data-testid={`event-row-${event.id}`}>
      <TableCell>{event.type}</TableCell>
      <TableCell>
        <TooltipCopyBadge value={event.id} variant="outline">
          {event.id.slice(0, 8)}
        </TooltipCopyBadge>
      </TableCell>
      <TableCell>{event.source}</TableCell>
      <TableCell>{event.receivedAt}</TableCell>
    </TableRow>
  </TooltipProvider>
);

export default Row;

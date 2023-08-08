import { TableCell, TableRow } from "~/design/Table";

import { EventSchemaType } from "~/api/events/schema";
import { TooltipProvider } from "~/design/Tooltip";

const Row = ({
  event,
  namespace,
}: {
  event: EventSchemaType;
  namespace: string;
}) => (
  <TooltipProvider>
    <TableRow>
      <TableCell>{event.type}</TableCell>
      <TableCell>{event.id}</TableCell>
      <TableCell>{event.source}</TableCell>
      <TableCell>{event.receivedAt}</TableCell>
    </TableRow>
  </TooltipProvider>
);

export default Row;

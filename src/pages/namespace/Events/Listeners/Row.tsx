import { TableCell, TableRow } from "~/design/Table";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";

import { EventListenerSchemaType } from "~/api/eventListeners/schema";
import { useTranslation } from "react-i18next";
import useUpdatedAt from "~/hooksNext/useUpdatedAt";

const Row = ({
  listener,
}: {
  listener: EventListenerSchemaType;
  namespace: string;
}) => {
  const { t } = useTranslation();
  const createdAt = useUpdatedAt(listener.createdAt);

  const eventTypes = listener.events.map((event) => event.type).join(", ");

  return (
    <TooltipProvider>
      <TableRow>
        <TableCell>{listener.workflow}</TableCell>
        <TableCell>TODO: Where does this come from?</TableCell>
        <TableCell>{listener.mode}</TableCell>
        <TableCell>
          <Tooltip>
            <TooltipTrigger data-testid="receivedAt-tooltip-trigger">
              {t("pages.events.listeners.tableRow.realtiveTime", {
                relativeTime: createdAt,
              })}
            </TooltipTrigger>
            <TooltipContent data-testid="receivedAt-tooltip-content">
              {listener.createdAt}
            </TooltipContent>
          </Tooltip>
        </TableCell>
        <TableCell>{eventTypes}</TableCell>
      </TableRow>
    </TooltipProvider>
  );
};

export default Row;

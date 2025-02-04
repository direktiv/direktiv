import { TableCell, TableRow } from "~/design/Table";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";

import ContextFilters from "./ContextFilters";
import CopyButton from "~/design/CopyButton";
import { EventListenerSchemaType } from "~/api/eventListeners/schema";
import { Link } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";
import useUpdatedAt from "~/hooks/useUpdatedAt";

const Row = ({
  listener,
  namespace,
}: {
  listener: EventListenerSchemaType;
  namespace: string;
}) => {
  const { t } = useTranslation();
  const createdAt = useUpdatedAt(listener.createdAt);

  const { triggerWorkflow: workflow, triggerInstance: instance } = listener;

  const listenerType = instance ? "instance" : "workflow";
  const target = workflow || instance;
  const contextFilters = listener.eventContextFilters.filter(
    (item) => !!Object.keys(item.context).length
  );

  const eventTypes = listener.listeningForEventTypes
    .map((eventType) => eventType)
    .join(", ");

  return (
    <TooltipProvider>
      <TableRow>
        <TableCell>
          {t(`pages.events.listeners.tableRow.type.${listenerType}`)}
        </TableCell>
        <TableCell>
          {workflow ? (
            <Link
              to="/n/$namespace/explorer/workflow/overview/$"
              params={{ namespace, _splat: workflow }}
            >
              {target}
            </Link>
          ) : (
            instance && (
              <Link
                to="/n/$namespace/explorer/workflow/overview/$"
                params={{ namespace, _splat: instance }}
              >
                {target}
              </Link>
            )
          )}
        </TableCell>
        <TableCell>{listener.triggerType}</TableCell>
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
        <TableCell>
          <Tooltip>
            <TooltipTrigger>
              <div className="w-40 truncate text-left">{eventTypes}</div>
            </TooltipTrigger>
            <TooltipContent>
              {eventTypes}
              <CopyButton
                value={eventTypes}
                buttonProps={{
                  size: "sm",
                  onClick: (e) => {
                    e.stopPropagation();
                  },
                }}
              />
            </TooltipContent>
          </Tooltip>
        </TableCell>
        <TableCell>
          <ContextFilters filters={contextFilters} />
        </TableCell>
      </TableRow>
    </TooltipProvider>
  );
};

export default Row;

import { TableCell, TableRow } from "~/design/Table";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";

import CopyButton from "~/design/CopyButton";
import { EventListenerSchemaType } from "~/api/eventListenersv2/schema";
import { Link } from "react-router-dom";
import { pages } from "~/util/router/pages";
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

  // TODO: rename triggerWorkflow, triggerInstance in API?
  const { triggerWorkflow: workflow, triggerInstance: instance } = listener;
  const listenerType = instance ? "instance" : "workflow";
  const target = workflow || instance;

  let linkTarget;

  if (workflow) {
    linkTarget = pages.explorer.createHref({
      namespace,
      path: workflow,
      subpage: "workflow",
    });
  }

  if (instance) {
    linkTarget = pages.instances.createHref({
      namespace,
      instance,
    });
  }

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
          {linkTarget ? <Link to={linkTarget}>{target}</Link> : <>{target}</>}
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
      </TableRow>
    </TooltipProvider>
  );
};

export default Row;

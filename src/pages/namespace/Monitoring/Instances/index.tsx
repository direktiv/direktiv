import { Boxes, CheckCircle2, XCircle } from "lucide-react";
import { NoResult, Table, TableBody } from "~/design/Table";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";

import { InstanceCard } from "./InstanceCard";
import { InstanceRow } from "./Row";
import RefreshButton from "~/design/RefreshButton";
import { ScrollArea } from "~/design/ScrollArea";
import { useInstances } from "~/api/instances/query/get";
import { useTranslation } from "react-i18next";

export const Instances = () => {
  const {
    data: SuccessfulInstances,
    isFetched: isFetchedSuccessfulInstances,
    isFetching: isFetchingSuccessfulInstances,
    refetch: refetchSuccessfulInstances,
  } = useInstances({
    limit: 10,
    offset: 0,
    filters: {
      STATUS: {
        type: "MATCH",
        value: "complete",
      },
    },
  });

  const {
    data: failedInstances,
    isFetched: isFetchedFailedInstances,
    isFetching: isFetchingFailedInstances,
    refetch: refetchFailedInstances,
  } = useInstances({
    limit: 10,
    offset: 0,
    filters: {
      STATUS: {
        type: "MATCH",
        value: "failed",
      },
    },
  });

  const { t } = useTranslation();

  const refetchButton = (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <RefreshButton
            icon
            size="sm"
            variant="ghost"
            disabled={isFetchingSuccessfulInstances}
            onClick={() => {
              refetchSuccessfulInstances();
            }}
          />
        </TooltipTrigger>
        <TooltipContent>
          {t(`pages.monitoring.instances.updateTooltip`)}
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );

  if (!isFetchedSuccessfulInstances || !isFetchedFailedInstances) return null;

  return (
    <>
      <InstanceCard
        headline={t("pages.monitoring.instances.successfulExecutions.title")}
        icon={CheckCircle2}
        refetchButton={refetchButton}
      >
        {SuccessfulInstances?.instances?.results.length === 0 ? (
          <NoResult icon={Boxes}>
            {t("pages.monitoring.instances.successfulExecutions.empty")}
          </NoResult>
        ) : (
          <ScrollArea className="h-full">
            <Table>
              <TableBody>
                {SuccessfulInstances?.instances?.results.map((instance) => (
                  <InstanceRow key={instance.id} instance={instance} />
                ))}
              </TableBody>
            </Table>
          </ScrollArea>
        )}
      </InstanceCard>
      <InstanceCard
        headline={t("pages.monitoring.instances.failedExecutions.title")}
        icon={XCircle}
        refetchButton={
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <RefreshButton
                  icon
                  size="sm"
                  variant="ghost"
                  disabled={isFetchingFailedInstances}
                  onClick={() => {
                    refetchFailedInstances();
                  }}
                />
              </TooltipTrigger>
              <TooltipContent>
                {t(`pages.monitoring.instances.updateTooltip`)}
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        }
      >
        {failedInstances?.instances?.results.length === 0 ? (
          <NoResult icon={Boxes}>
            {t("pages.monitoring.instances.failedExecutions.empty")}
          </NoResult>
        ) : (
          <ScrollArea className="h-full">
            <Table>
              <TableBody>
                {failedInstances?.instances?.results.map((instance) => (
                  <InstanceRow key={instance.id} instance={instance} />
                ))}
              </TableBody>
            </Table>
          </ScrollArea>
        )}
      </InstanceCard>
    </>
  );
};

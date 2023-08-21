import { Boxes, CheckCircle2, XCircle } from "lucide-react";
import { NoResult, Table, TableBody } from "~/design/Table";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";

import { InstanceCard } from "./instanceCard";
import { InstanceRow } from "./Row";
import RefreshButton from "~/design/RefreshButton";
import { ScrollArea } from "~/design/ScrollArea";
import { useInstances } from "~/api/instances/query/get";
import { useTranslation } from "react-i18next";

export const Instances = () => {
  const {
    data: sucessfullInstances,
    isFetched: isFetchedSucessfullInstances,
    isFetching: isFetchingSucessfullinstances,
    refetch: refetchSucessfullInstances,
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
            disabled={isFetchingSucessfullinstances}
            onClick={() => {
              refetchSucessfullInstances();
            }}
          />
        </TooltipTrigger>
        <TooltipContent>
          {t(`pages.monitoring.instances.updateTooltip`)}
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );

  if (!isFetchedSucessfullInstances || !isFetchedFailedInstances) return null;

  return (
    <>
      <InstanceCard
        headline={t("pages.monitoring.instances.successfullExecutions.title")}
        icon={CheckCircle2}
        refetchButton={refetchButton}
      >
        {sucessfullInstances?.instances?.results.length === 0 ? (
          <NoResult icon={Boxes}>
            {t("pages.monitoring.instances.successfullExecutions.empty")}
          </NoResult>
        ) : (
          <ScrollArea className="h-full">
            <Table>
              <TableBody>
                {sucessfullInstances?.instances?.results.map((instance) => (
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

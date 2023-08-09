import { CheckCircle2, RefreshCcw, XCircle } from "lucide-react";
import { Table, TableBody } from "~/design/Table";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";

import Button from "~/design/Button";
import { InstanceCard } from "./instanceCard";
import { InstanceRow } from "./Row";
import NoResult from "./NoResult";
import { useInstances } from "~/api/instances/query/get";
import { useTranslation } from "react-i18next";

export const Instances = () => {
  const {
    data: completedInstances,
    isFetched: isFetchedCompleted,
    isFetching: isFetchingCompleted,
    refetch: refetchCompletedInstances,
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
    isFetched: isFetchedFailed,
    isFetching: isFetchingFailed,
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
  if (!isFetchedCompleted || !isFetchedFailed) return null;
  return (
    <>
      <InstanceCard
        headline={t("pages.monitoring.instances.successfullExecutions.title")}
        icon={CheckCircle2}
        refetchButton={
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  icon
                  size="sm"
                  variant="ghost"
                  disabled={isFetchingCompleted}
                  onClick={() => {
                    refetchCompletedInstances();
                  }}
                >
                  <RefreshCcw />
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                {t(`pages.monitoring.instances.updateTooltip`)}
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        }
      >
        {completedInstances?.instances?.results.length === 0 ? (
          <NoResult
            message={t(
              "pages.monitoring.instances.successfullExecutions.empty"
            )}
          />
        ) : (
          <Table>
            <TableBody>
              {completedInstances?.instances?.results.map((instance) => (
                <InstanceRow key={instance.id} instance={instance} />
              ))}
            </TableBody>
          </Table>
        )}
      </InstanceCard>
      <InstanceCard
        headline={t("pages.monitoring.instances.failedExecutions.title")}
        icon={XCircle}
        refetchButton={
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  icon
                  size="sm"
                  variant="ghost"
                  disabled={isFetchingFailed}
                  onClick={() => {
                    refetchFailedInstances();
                  }}
                >
                  <RefreshCcw />
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                {t(`pages.monitoring.instances.updateTooltip`)}
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        }
      >
        {failedInstances?.instances?.results.length === 0 ? (
          <NoResult
            message={t("pages.monitoring.instances.failedExecutions.empty")}
          />
        ) : (
          <Table>
            <TableBody>
              {failedInstances?.instances?.results.map((instance) => (
                <InstanceRow key={instance.id} instance={instance} />
              ))}
            </TableBody>
          </Table>
        )}
      </InstanceCard>
    </>
  );
};

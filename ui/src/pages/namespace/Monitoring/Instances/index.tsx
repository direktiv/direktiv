import { Boxes, CheckCircle2, XCircle } from "lucide-react";
import { NoPermissions, NoResult, Table, TableBody } from "~/design/Table";

import { InstanceCard } from "./InstanceCard";
import { InstanceRow } from "./Row";
import RefreshButton from "~/design/RefreshButton";
import { ScrollArea } from "~/design/ScrollArea";
import { useInstanceList } from "~/api/instances/query/get";
import { useTranslation } from "react-i18next";

export const Instances = () => {
  const {
    data: sucessfullInstances,
    isFetched: isFetchedSucessfullInstances,
    isFetching: isFetchingSucessfullinstances,
    refetch: refetchSucessfullInstances,
    isAllowed: isAllowedSucessfullInstances,
    noPermissionMessage: noPermissionMessageSucessfullInstances,
  } = useInstanceList({
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
    isAllowed: isAllowedFailedInstances,
    noPermissionMessage: noPermissionMessageFailedInstances,
  } = useInstanceList({
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

  if (!isFetchedSucessfullInstances || !isFetchedFailedInstances) return null;

  return (
    <>
      <InstanceCard
        headline={t("pages.monitoring.instances.successfulExecutions.title")}
        icon={CheckCircle2}
        refetchButton={
          <RefreshButton
            icon
            size="sm"
            variant="ghost"
            disabled={isFetchingSucessfullinstances}
            onClick={() => {
              refetchSucessfullInstances();
            }}
          />
        }
      >
        {isAllowedSucessfullInstances ? (
          <>
            {sucessfullInstances?.length === 0 ? (
              <NoResult icon={Boxes}>
                {t("pages.monitoring.instances.successfulExecutions.empty")}
              </NoResult>
            ) : (
              <ScrollArea className="h-full">
                <Table>
                  <TableBody>
                    {sucessfullInstances?.map((instance) => (
                      <InstanceRow key={instance.id} instance={instance} />
                    ))}
                  </TableBody>
                </Table>
              </ScrollArea>
            )}
          </>
        ) : (
          <NoPermissions>
            {noPermissionMessageSucessfullInstances}
          </NoPermissions>
        )}
      </InstanceCard>
      <InstanceCard
        headline={t("pages.monitoring.instances.failedExecutions.title")}
        icon={XCircle}
        refetchButton={
          <RefreshButton
            icon
            size="sm"
            variant="ghost"
            disabled={isFetchingFailedInstances}
            onClick={() => {
              refetchFailedInstances();
            }}
          />
        }
      >
        {isAllowedFailedInstances ? (
          <>
            {failedInstances?.length === 0 ? (
              <NoResult icon={Boxes}>
                {t("pages.monitoring.instances.failedExecutions.empty")}
              </NoResult>
            ) : (
              <ScrollArea className="h-full">
                <Table>
                  <TableBody>
                    {failedInstances?.map((instance) => (
                      <InstanceRow key={instance.id} instance={instance} />
                    ))}
                  </TableBody>
                </Table>
              </ScrollArea>
            )}
          </>
        ) : (
          <NoPermissions>{noPermissionMessageFailedInstances}</NoPermissions>
        )}
      </InstanceCard>
    </>
  );
};

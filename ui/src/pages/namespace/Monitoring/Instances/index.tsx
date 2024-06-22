import { Boxes, CheckCircle2, XCircle } from "lucide-react";
import { NoPermissions, NoResult, Table, TableBody } from "~/design/Table";

import { InstanceCard } from "./InstanceCard";
import { InstanceRow } from "./Row";
import { ScrollArea } from "~/design/ScrollArea";
import { useInstances } from "~/api/instances/query/get";
import { useTranslation } from "react-i18next";

export const Instances = () => {
  const {
    data: dataSuccessfulInstances,
    isFetched: isFetchedSuccessfulInstances,
    isAllowed: isAllowedSuccessfulInstances,
    noPermissionMessage: noPermissionMessageSuccessfulInstances,
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
    data: dataFailedInstances,
    isFetched: isFetchedFailedInstances,
    isAllowed: isAllowedFailedInstances,
    noPermissionMessage: noPermissionMessageFailedInstances,
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

  const successfulInstances = dataSuccessfulInstances?.data ?? [];

  const failedInstances = dataFailedInstances?.data ?? [];

  if (!isFetchedSuccessfulInstances || !isFetchedFailedInstances) return null;

  return (
    <>
      <InstanceCard
        headline={t("pages.monitoring.instances.successfulExecutions.title")}
        icon={CheckCircle2}
      >
        {isAllowedSuccessfulInstances ? (
          <>
            {successfulInstances.length === 0 ? (
              <NoResult icon={Boxes}>
                {t("pages.monitoring.instances.successfulExecutions.empty")}
              </NoResult>
            ) : (
              <ScrollArea className="h-full">
                <Table>
                  <TableBody>
                    {successfulInstances.map((instance) => (
                      <InstanceRow key={instance.id} instance={instance} />
                    ))}
                  </TableBody>
                </Table>
              </ScrollArea>
            )}
          </>
        ) : (
          <NoPermissions>
            {noPermissionMessageSuccessfulInstances}
          </NoPermissions>
        )}
      </InstanceCard>
      <InstanceCard
        headline={t("pages.monitoring.instances.failedExecutions.title")}
        icon={XCircle}
      >
        {isAllowedFailedInstances ? (
          <>
            {failedInstances.length === 0 ? (
              <NoResult icon={Boxes}>
                {t("pages.monitoring.instances.failedExecutions.empty")}
              </NoResult>
            ) : (
              <ScrollArea className="h-full">
                <Table>
                  <TableBody>
                    {failedInstances.map((instance) => (
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

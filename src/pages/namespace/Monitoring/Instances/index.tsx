import { CheckCircle2, XCircle } from "lucide-react";
import { Table, TableBody } from "~/design/Table";

import { InstanceCard } from "./instanceCard";
import { InstanceRow } from "./Row";
import NoResult from "../../Instances/List/NoResult";
import { useInstances } from "~/api/instances/query/get";
import { useTranslation } from "react-i18next";

const useInstancesBatch = () => {
  const { data: failedInstances, isFetched: isFetchedFailed } = useInstances({
    limit: 10,
    offset: 0,
    filters: {
      STATUS: {
        type: "MATCH",
        value: "failed",
      },
    },
  });

  const { data: completedInstances, isFetched: isFetchedCompleted } =
    useInstances({
      limit: 10,
      offset: 0,
      filters: {
        STATUS: {
          type: "MATCH",
          value: "complete",
        },
      },
    });

  return {
    isFetched: isFetchedFailed && isFetchedCompleted,
    failedInstances,
    completedInstances,
  };
};

export const Instances = () => {
  const { isFetched, completedInstances, failedInstances } =
    useInstancesBatch();

  const { t } = useTranslation();
  if (!isFetched) return null;
  return (
    <>
      <InstanceCard
        headline={t("pages.monitoring.instances.successfullExecutions.title")}
        icon={CheckCircle2}
      >
        {completedInstances?.instances?.results.length === 0 ? (
          <div className="flex grow justify-center">
            <NoResult
              message={t(
                "pages.monitoring.instances.successfullExecutions.empty"
              )}
            />
          </div>
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
      >
        {failedInstances?.instances?.results.length === 0 ? (
          <div className="flex grow justify-center">
            <NoResult
              message={t("pages.monitoring.instances.failedExecutions.empty")}
            />
          </div>
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

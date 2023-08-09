import { Card } from "~/design/Card";
import { InstanceRow } from "./Row";
import { useInstances } from "~/api/instances/query/get";

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

  if (!isFetched) return null;

  return (
    <>
      <Card className="p-5">
        {completedInstances?.instances?.results.map((instance) => (
          <InstanceRow key={instance.id} instance={instance} />
        ))}
      </Card>
      <Card className="p-5">
        {failedInstances?.instances?.results.map((instance) => (
          <InstanceRow key={instance.id} instance={instance} />
        ))}
      </Card>
    </>
  );
};

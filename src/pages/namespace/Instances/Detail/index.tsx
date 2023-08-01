import { FC, useState } from "react";
import { FiltersObj, useLogsStream } from "~/api/logs/query/get";

import { InstanceStateProvider } from "./state/instanceContext";
import InstancesDetail from "./InstanceDetail";
import { pages } from "~/util/router/pages";
import { useInstanceDetailsStream } from "~/api/instances/query/details";

const Instance: FC<{ instance: string }> = ({ instance }) => {
  const [query, setQuery] = useState<FiltersObj>({});

  useInstanceDetailsStream(
    { instanceId: instance ?? "" },
    { enabled: !!instance }
  );
  useLogsStream(
    {
      instanceId: instance ?? "",
      filters: query,
    },
    { enabled: !!instance }
  );

  if (!instance) return null;

  // Details page is moved into a separate component to give us a state
  // where the id alwawys defined. This is required for the data fetching
  // hook that require the id (and hooks can not be conditionally called)
  return <InstancesDetail query={query} setQuery={setQuery} />;
};

const InstanceWithContextProvider = () => {
  const { instance } = pages.instances.useParams();
  if (!instance) return null;

  return (
    <InstanceStateProvider instance={instance}>
      <Instance instance={instance} />
    </InstanceStateProvider>
  );
};

export default InstanceWithContextProvider;

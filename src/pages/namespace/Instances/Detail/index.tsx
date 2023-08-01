import {
  InstanceStateProvider,
  useFilters,
  useInstanceId,
} from "./state/instanceContext";

import InstancesDetail from "./InstanceDetail";
import { pages } from "~/util/router/pages";
import { useInstanceDetailsStream } from "~/api/instances/query/details";
import { useLogsStream } from "~/api/logs/query/get";

const Instance = () => {
  const instanceId = useInstanceId();
  const filters = useFilters();

  useInstanceDetailsStream({ instanceId });
  useLogsStream({ instanceId, filters });

  // Details page is moved into a separate component to give us a state
  // where the id alwawys defined. This is required for the data fetching
  // hook that require the id (and hooks can not be conditionally called)
  return <InstancesDetail />;
};

const InstanceWithContextProvider = () => {
  const { instance: instanceId } = pages.instances.useParams();
  if (!instanceId) return null;

  return (
    <InstanceStateProvider instanceId={instanceId}>
      <Instance />
    </InstanceStateProvider>
  );
};

export default InstanceWithContextProvider;

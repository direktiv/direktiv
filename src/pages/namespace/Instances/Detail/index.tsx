import {
  InstanceStateProvider,
  useFilters,
  useInstanceId,
} from "./store/instanceContext";

import { InstanceStreamingProvider } from "~/api/instances/query/details";
import InstancesDetail from "./InstanceDetail";
import { LogStreamingProvider } from "~/api/logs/query/get";
import { pages } from "~/util/router/pages";

const InstanceStreaming = () => {
  const instanceId = useInstanceId();
  const filters = useFilters();
  return (
    <>
      <InstanceStreamingProvider instanceId={instanceId} />
      <LogStreamingProvider filters={filters} instanceId={instanceId} />
      <InstancesDetail />
    </>
  );
};

const InstanceWithContextProvider = () => {
  const { instance: instanceId } = pages.instances.useParams();
  if (!instanceId) return null;

  return (
    <InstanceStateProvider instanceId={instanceId}>
      <InstanceStreaming />
    </InstanceStateProvider>
  );
};

export default InstanceWithContextProvider;

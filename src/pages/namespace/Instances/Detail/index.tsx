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

  /**
   * the streaming hooks will update the react query cache
   * when it received new data. This will trigger a rerender
   * of all components that consume this data via useQuery.
   *
   * This is why it's important to place this hook in a separate
   * parent component on top of the consuming components. This
   * will ensure that the hook will not rerun itself (canceling
   * the stream and restarting a new one when it updates the cache
   */
  useInstanceDetailsStream({ instanceId });
  useLogsStream({ instanceId, filters });

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

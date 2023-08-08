import {
  InstanceStateProvider,
  useFilters,
  useInstanceId,
} from "./store/instanceContext";

import InstancesDetail from "./InstanceDetail";
import { pages } from "~/util/router/pages";
import { useInstanceDetailsStream } from "~/api/instances/query/details";
import { useLogsStream } from "~/api/logs/query/get";

const InstanceStreaming = () => {
  const instanceId = useInstanceId();
  const filters = useFilters();

  /**
   * each of these hooks will subscribe to a SSE stream. They will
   * setup a connection on mount and cancel the connection on unmount.
   * To avoid unnecessary reconnects, make sure to place this hooks
   * in a isolated a parent component that will not rerender very often.
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
      <InstanceStreaming />
    </InstanceStateProvider>
  );
};

export default InstanceWithContextProvider;

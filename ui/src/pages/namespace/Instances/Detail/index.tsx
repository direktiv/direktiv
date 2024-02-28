import {
  InstanceStateProvider,
  useFilters,
  useInstanceId,
} from "./store/instanceContext";

import { InstanceStreamingSubscriber } from "~/api/instances/query/details";
import InstancesDetail from "./InstanceDetail";
import { LogStreamingSubscriber } from "~/api/logs/query/logs";
import { LogStreamingSubscriber as LogStreamingSubscriberDeprecated } from "~/api/logs_DEPRECATED/query/get";
import { pages } from "~/util/router/pages";

const InstanceStreaming = () => {
  const instanceId = useInstanceId();
  const filters = useFilters();
  return (
    <>
      <InstanceStreamingSubscriber instanceId={instanceId} />
      <LogStreamingSubscriberDeprecated
        filters={filters}
        instanceId={instanceId}
      />
      <LogStreamingSubscriber instance={instanceId} />
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

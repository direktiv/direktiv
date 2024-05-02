import { InstanceStateProvider, useInstanceId } from "./store/instanceContext";

import { InstanceStreamingSubscriber } from "~/api/instances/query/details/streaming";
import InstancesDetail from "./InstanceDetail";
import { LogStreamingSubscriber } from "~/api/logs/query/LogStreamingSubscriber";
import { pages } from "~/util/router/pages";
import { useInstanceDetails } from "~/api/instances/query/details";

const InstanceStreaming = () => {
  const instanceId = useInstanceId();
  const { data, isSuccess } = useInstanceDetails({ instanceId });
  const stopStreaming = isSuccess && data?.status !== "pending";

  return (
    <>
      <InstanceStreamingSubscriber
        instanceId={instanceId}
        enabled={!stopStreaming}
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

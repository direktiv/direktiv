import { InstanceStateProvider, useInstanceId } from "./store/instanceContext";

import { InstanceStreamingSubscriber } from "~/api/instances_obsolete/query/details";
import InstancesDetail from "./InstanceDetail";
import { LogStreamingSubscriber } from "~/api/logs/query/LogStreamingSubscriber";
import { pages } from "~/util/router/pages";

const InstanceStreaming = () => {
  const instanceId = useInstanceId();
  return (
    <>
      <InstanceStreamingSubscriber instanceId={instanceId} />
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

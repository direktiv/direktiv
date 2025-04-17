import { InstanceStateProvider, useInstanceId } from "./store/instanceContext";

import { InstanceStreamingSubscriber } from "~/api/instances/query/details/streaming";
import InstancesDetail from "./InstanceDetail";
import { LogStreamingSubscriber } from "~/api/logs/query/LogStreamingSubscriber";
import { useInstanceDetails } from "~/api/instances/query/details";
import { useParams } from "@tanstack/react-router";

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
  const { id } = useParams({ strict: false });
  if (!id) return null;

  return (
    <InstanceStateProvider instanceId={id}>
      <InstanceStreaming />
    </InstanceStateProvider>
  );
};

export default InstanceWithContextProvider;

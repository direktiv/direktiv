import InstancesDetail from "./InstanceDetail";
import { pages } from "~/util/router/pages";

const Instance = () => {
  const { instance } = pages.instances.useParams();
  if (!instance) return null;

  // Details page is moved into a separate component to give us a state
  // where the id alwawys defined. This is required for the data fetching
  // hook that require the id (and hooks can not be conditionally called)
  return <InstancesDetail instanceId={instance} />;
};

export default Instance;

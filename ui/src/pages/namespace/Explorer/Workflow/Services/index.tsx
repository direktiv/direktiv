import { FC } from "react";
import ServiceDetails from "./ServiceDetails";
import ServicesList from "./ServicesList";
import { usePages } from "~/util/router/pages";

const WorkflowServicesPage: FC = () => {
  const pages = usePages();
  const { path, serviceId } = pages.explorer.useParams();

  if (!path) return null;

  if (serviceId) {
    return <ServiceDetails serviceId={serviceId} />;
  }

  return <ServicesList workflow={path} />;
};

export default WorkflowServicesPage;

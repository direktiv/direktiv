import { FC } from "react";
import ServiceDetails from "./ServiceDetails";
import ServicesList from "./ServicesList";
import { pages } from "~/util/router/pages";

const WorkflowServicesPage: FC = () => {
  const { path, serviceId } = pages.explorer.useParams();

  if (!path) return null;

  if (serviceId) {
    return <ServiceDetails serviceId={serviceId} />;
  }

  return <ServicesList workflow={path} />;
};

export default WorkflowServicesPage;

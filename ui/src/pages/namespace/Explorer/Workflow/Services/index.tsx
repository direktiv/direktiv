import { FC } from "react";
import ServiceDetails from "./ServiceDetails";
import ServicesList from "./ServicesList";
import { useParams } from "@tanstack/react-router";

const WorkflowServicesPage: FC = () => {
  const { _splat: path, id: serviceId } = useParams({ strict: false });

  if (!path) return null;

  if (serviceId) {
    return <ServiceDetails serviceId={serviceId} />;
  }

  return <ServicesList workflow={path} />;
};

export default WorkflowServicesPage;

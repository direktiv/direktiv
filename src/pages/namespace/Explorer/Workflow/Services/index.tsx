import { FC } from "react";
import RevisionDetails from "./RevisionDetails";
import ServiceDetails from "./ServiceDetails";
import ServicesList from "./ServicesList";
import { pages } from "~/util/router/pages";
import { useSearchParams } from "react-router-dom";

const WorkflowServicesPage: FC = () => {
  const { path } = pages.explorer.useParams();
  const [searchParams] = useSearchParams();

  const name = searchParams.get("name");
  const version = searchParams.get("version");
  const revision = searchParams.get("revision");

  if (!path) return null;

  if (name && version && revision) {
    return <RevisionDetails />;
  }

  if (name && version) {
    return <ServiceDetails workflow={path} service={name} version={version} />;
  }

  return <ServicesList workflow={path} />;
};

export default WorkflowServicesPage;

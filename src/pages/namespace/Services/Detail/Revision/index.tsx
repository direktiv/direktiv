import { pages } from "~/util/router/pages";

const ServiceRevisionPage = () => {
  const { service, revision } = pages.services.useParams();
  if (!service) return null;

  return (
    <h1>
      revisions page: service: {service} - revision: {revision}
    </h1>
  );
};

export default ServiceRevisionPage;

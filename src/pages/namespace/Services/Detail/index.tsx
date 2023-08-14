import { pages } from "~/util/router/pages";

const ServiceDetailPage = () => {
  const { service } = pages.services.useParams();
  if (!service) return null;

  return <h1>service detail page: {service}</h1>;
};

export default ServiceDetailPage;

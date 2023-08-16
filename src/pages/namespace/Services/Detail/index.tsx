import { pages } from "~/util/router/pages";
import { useServiceDetails } from "~/api/services/query/details";

const ServiceDetailPage = () => {
  const { service } = pages.services.useParams();

  const { data } = useServiceDetails({
    service,
  });
  if (!service) return null;

  return <h1>service detail page: {service}</h1>;
};

export default ServiceDetailPage;

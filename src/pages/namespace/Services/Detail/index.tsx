import { pages } from "~/util/router/pages";
import { useServiceDetails } from "~/api/services/query/details";

const ServiceDetailPage = () => {
  const { service } = pages.services.useParams();

  const { data } = useServiceDetails({
    service: service ?? "",
  });
  if (!service) return null;

  return (
    <div>
      <h1>service detail</h1>
      {data?.revisions?.map((revision) => (
        <div key={revision.name}>{revision.name}</div>
      ))}
    </div>
  );
};

export default ServiceDetailPage;

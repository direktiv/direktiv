import Header from "./Header";
import { Pods } from "./Pods";
import { ServiceRevisionStreamingSubscriber } from "~/api/services/query/revision/getAll";
import { pages } from "~/util/router/pages";

const ServiceRevisionPage = () => {
  const { service, revision } = pages.services.useParams();

  if (!service) return null;
  if (!revision) return null;

  return (
    <div className="grid grow grid-rows-[auto_1fr]">
      <ServiceRevisionStreamingSubscriber
        revision={revision}
        service={service}
      />
      <Header service={service} revision={revision} />
      <div className="p-5">
        <Pods revision={revision} service={service} />
      </div>
    </div>
  );
};

export default ServiceRevisionPage;

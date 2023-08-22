import Header from "./Header";
import { Pods } from "./Pods";
import { ServiceRevisionStreamingSubscriber } from "~/api/services/query/revision/getAll";
import { pages } from "~/util/router/pages";

const ServiceRevisionPage = () => {
  const { service, revision } = pages.services.useParams();

  if (!service) return null;
  if (!revision) return null;

  return (
    <div className="flex grow flex-col border">
      <ServiceRevisionStreamingSubscriber
        revision={revision}
        service={service}
      />
      <div className="flex-none">
        <Header service={service} revision={revision} />
      </div>
      <div className="grow border">
        <Pods revision={revision} service={service} />
      </div>
    </div>
  );
};

export default ServiceRevisionPage;

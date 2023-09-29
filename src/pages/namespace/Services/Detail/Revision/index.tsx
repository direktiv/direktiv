import { Card } from "~/design/Card";
import Header from "./Header";
import { NoPermissions } from "~/design/Table";
import { Pods } from "./Pods";
import { ServiceRevisionStreamingSubscriber } from "~/api/services/query/revision/getAll";
import { pages } from "~/util/router/pages";
import { usePods } from "~/api/services/query/revision/pods/getAll";

const ServiceRevisionPage = () => {
  const { service, revision } = pages.services.useParams();
  const { isFetched, isAllowed, noPermissionMessage } = usePods({
    service: service ?? "",
    revision: revision ?? "",
  });

  if (!service) return null;
  if (!revision) return null;
  if (!isFetched) return null;
  if (!isAllowed)
    return (
      <Card className="m-5 flex grow flex-col p-4">
        <NoPermissions>{noPermissionMessage}</NoPermissions>
      </Card>
    );

  return (
    <div className="flex grow flex-col">
      <ServiceRevisionStreamingSubscriber
        revision={revision}
        service={service}
      />
      <div className="flex-none">
        <Header service={service} revision={revision} />
      </div>
      <Pods revision={revision} service={service} />
    </div>
  );
};

export default ServiceRevisionPage;

import { Card } from "~/design/Card";
import Header from "./Header";
import { NoPermissions } from "~/design/Table";
import { Pods } from "./Pods";
import { useParams } from "@tanstack/react-router";
import { usePods } from "~/api/services/query/pods";

const ServiceDetailPage = () => {
  const { service: serviceId } = useParams({ strict: false });
  const { isFetched, isAllowed, noPermissionMessage } = usePods(
    serviceId ?? ""
  );

  if (!serviceId) return null;
  if (!isFetched) return null;
  if (!isAllowed)
    return (
      <Card className="m-5 flex grow flex-col p-4">
        <NoPermissions>{noPermissionMessage}</NoPermissions>
      </Card>
    );

  return (
    <div className="flex grow flex-col">
      <div className="flex-none">
        <Header serviceId={serviceId} />
      </div>
      <Pods serviceId={serviceId} />
    </div>
  );
};

export default ServiceDetailPage;

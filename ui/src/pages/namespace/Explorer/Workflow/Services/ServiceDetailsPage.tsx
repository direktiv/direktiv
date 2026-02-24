import Header from "~/pages/namespace/Services/Detail/Header";
import { Pods } from "~/pages/namespace/Services/Detail/Pods";
import { useParams } from "@tanstack/react-router";

const ServiceDetailsPage = () => {
  const { service } = useParams({
    from: "/n/$namespace/explorer/workflow/services/$service/$",
  });

  return (
    <div className="flex grow flex-col">
      <div className="flex-none">
        <Header serviceId={service} />
      </div>
      <Pods
        serviceId={service}
        className="md:h-[calc(100vh-38rem)] lg:h-[calc(100vh-33rem)]"
      />
    </div>
  );
};

export default ServiceDetailsPage;

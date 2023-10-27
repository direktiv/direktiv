import Header from "~/pages/namespace/Services/Detail/Header";
import { Pods } from "~/pages/namespace/Services/Detail/Pods";

const ServiceDetails = ({ serviceId }: { serviceId: string }) => (
  <div className="flex grow flex-col">
    <div className="flex-none">
      <Header serviceId={serviceId} />
    </div>
    <Pods
      serviceId={serviceId}
      className="md:h-[calc(100vh-38rem)] lg:h-[calc(100vh-33rem)]"
    />
  </div>
);

export default ServiceDetails;

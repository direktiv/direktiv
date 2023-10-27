import { Pods } from "~/pages/namespace/Services/Detail/Pods";

const ServiceDetails = ({ serviceId }: { serviceId: string }) => (
  <div className="flex grow flex-col">
    <Pods serviceId={serviceId} />
  </div>
);

export default ServiceDetails;

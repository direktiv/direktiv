import GatewayInfo from "./GatewayInfo";
import OpenAPISpec from "./OpenAPISpec";
import ResizeablePanel from "~/util/resizablePanel";

const InfoPage = () => (
  <div className="flex grow flex-col gap-y-4 p-5 w-full">
    <ResizeablePanel leftPanel={<GatewayInfo />} rightPanel={<OpenAPISpec />} />
  </div>
);

export default InfoPage;

import GatewayInfo from "./GatewayInfo";
import OpenAPISpec from "./OpenAPISpec";
import ResizablePanel from "./ResizablePanel";

const InfoPage = () => (
  <div className="flex grow flex-col gap-y-4 p-5 w-full">
    <ResizablePanel leftPanel={<GatewayInfo />} rightPanel={<OpenAPISpec />} />
  </div>
);

export default InfoPage;

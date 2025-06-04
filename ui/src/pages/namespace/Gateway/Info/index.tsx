import GatewayInfo from "./GatewayInfo";
import OpenAPISpec from "./OpenAPISpec";
import ResizablePanel from "./ResizablePanel";

const InfoPage = () => (
  <div className="flex w-full grow flex-col gap-y-4 p-5">
    <ResizablePanel leftPanel={<GatewayInfo />} rightPanel={<OpenAPISpec />} />
  </div>
);

export default InfoPage;

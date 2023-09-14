import Header from "~/pages/namespace/Services/Detail/Revision/Header";
import { Pods } from "~/pages/namespace/Services/Detail/Revision/Pods";
import { ServiceRevisionStreamingSubscriber } from "~/api/services/query/revision/getAll";
import { pages } from "~/util/router/pages";
import { useSearchParams } from "react-router-dom";

const RevisionDetails = () => {
  const [searchParams] = useSearchParams();
  const { path: workflow } = pages.explorer.useParams();

  const service = searchParams.get("name");
  const serviceVersion = searchParams.get("version");
  const serviceRevision = searchParams.get("revision");

  if (!workflow || !service || !serviceVersion || !serviceRevision) return null;

  return (
    <div className="flex grow flex-col">
      <ServiceRevisionStreamingSubscriber
        workflow={workflow}
        service={service}
        revision={serviceRevision}
        version={serviceVersion}
      />
      <div className="flex-none">
        <Header service={service} revision={serviceRevision} />
      </div>
      <Pods
        workflow={workflow}
        service={service}
        revision={serviceRevision}
        version={serviceVersion}
        className="md:h-[calc(100vh-38rem)] lg:h-[calc(100vh-33rem)]"
      />
    </div>
  );
};

export default RevisionDetails;

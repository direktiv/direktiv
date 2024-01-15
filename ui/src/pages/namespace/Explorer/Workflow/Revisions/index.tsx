import RevisionsDetailPage from "./Detail";
import RevisionsOverviewPage from "./Overview";
import { pages } from "~/util/router/pages";

const RevisionsPage = () => {
  const { revision: selectedRevision } = pages.explorer.useParams();
  return (
    <div className="flex grow flex-col gap-y-4 p-5">
      {selectedRevision ? <RevisionsDetailPage /> : <RevisionsOverviewPage />}
    </div>
  );
};

export default RevisionsPage;

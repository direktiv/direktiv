import { Outlet } from "react-router-dom";
import { UnsavedChangesStateProvider } from "./Workflow/store/unsavedChangesContext";
import { isApiErrorSchema } from "~/api/errorHandling";
import { useFile } from "~/api/files/query/file";
import { usePages } from "~/util/router/pages";

const ExplorerPage = () => {
  const pages = usePages();
  const { path } = pages.explorer.useParams();
  const { isError, error, isFetched } = useFile({ path });
  if (!isFetched) return null;

  // forward 404 errors to the routers error boundary
  if (isError && isApiErrorSchema(error) && error.status === 404) {
    throw error;
  }

  return (
    <UnsavedChangesStateProvider>
      <Outlet />
    </UnsavedChangesStateProvider>
  );
};

export default ExplorerPage;

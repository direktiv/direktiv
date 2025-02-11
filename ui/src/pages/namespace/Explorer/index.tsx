import { Outlet, useParams } from "@tanstack/react-router";

import { UnsavedChangesStateProvider } from "./Workflow/store/unsavedChangesContext";
import { isApiErrorSchema } from "~/api/errorHandling";
import { useFile } from "~/api/files/query/file";

const ExplorerWrapper = () => {
  const { _splat: path } = useParams({ strict: false });
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

export default ExplorerWrapper;

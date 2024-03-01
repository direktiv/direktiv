import { Outlet } from "react-router-dom";
import { isApiErrorSchema } from "~/api/errorHandling";
import { pages } from "~/util/router/pages";
import { useFile } from "~/api/files/query/file";

const ExplorerPage = () => {
  const { path } = pages.explorer.useParams();
  const { isError, error, isFetched } = useFile({ path });
  if (!isFetched) return null;

  if (isError && isApiErrorSchema(error) && error.response.status === 404) {
    throw new Error("this file does not exist");
  }

  return <Outlet />;
};

export default ExplorerPage;

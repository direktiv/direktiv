import NoMirror from "./NoMirror";
import { Outlet } from "react-router-dom";
import { useNamespaceDetail } from "~/api/namespaces/query/get";

const MirrorPage = () => {
  const { data, isSuccess } = useNamespaceDetail();

  if (!isSuccess) return null;

  const isMirror = data?.mirror;

  return (
    <div className="flex grow flex-col">
      {isMirror ? <Outlet /> : <NoMirror />}
    </div>
  );
};

export default MirrorPage;

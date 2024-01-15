import NoMirror from "./NoMirror";
import { Outlet } from "react-router-dom";
import { useNodeContent } from "~/api/tree/query/node";

const MirrorPage = () => {
  const { data, isSuccess } = useNodeContent({ path: "/" });

  if (!isSuccess) return null;

  const isMirror = data.node.expandedType === "git";

  return (
    <div className="flex grow flex-col">
      {isMirror ? <Outlet /> : <NoMirror />}
    </div>
  );
};

export default MirrorPage;

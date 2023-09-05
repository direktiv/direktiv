import NoMirror from "./NoMirror";
import { Outlet } from "react-router-dom";
import { useNodeContent } from "~/api/tree/query/node";

const MirrorPage = () => {
  const { data } = useNodeContent({ path: "/" });

  const isMirror = data?.node?.expandedType === "git";

  return <div>{isMirror ? <Outlet /> : <NoMirror />}</div>;
};

export default MirrorPage;

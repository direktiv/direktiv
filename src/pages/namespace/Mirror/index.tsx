import { Outlet } from "react-router-dom";

const MirrorPage = () => {
  const text = "Mirror page";
  return (
    <div>
      <h2>{text}</h2>
      <Outlet />
    </div>
  );
};

export default MirrorPage;

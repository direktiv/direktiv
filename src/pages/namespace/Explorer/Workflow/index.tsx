import { FC } from "react";
import Header from "./Header";
import { Outlet } from "react-router-dom";

const WorkflowPage: FC = () => (
  <>
    <Header />
    <Outlet />
  </>
);

export default WorkflowPage;

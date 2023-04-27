import { Breadcrumb, BreadcrumbRoot } from "./index";

import { Home } from "lucide-react";

export default {
  title: "Components/Breadcrumb",
};

export const Default = () => (
  <div>
    <BreadcrumbRoot>
      <Breadcrumb>Home</Breadcrumb>
      <Breadcrumb>Subfolder</Breadcrumb>
      <Breadcrumb>some-file.yml</Breadcrumb>
    </BreadcrumbRoot>
  </div>
);

export const DefaultWithIcons = () => (
  <div className="bg-white dark:bg-black">
    <BreadcrumbRoot>
      <Breadcrumb>
        <Home /> My-namespace
      </Breadcrumb>
      <Breadcrumb>
        <Home /> My-namespace
      </Breadcrumb>
      <Breadcrumb>
        <Home /> My-namespace
      </Breadcrumb>
    </BreadcrumbRoot>
  </div>
);

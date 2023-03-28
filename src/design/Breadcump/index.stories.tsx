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
  <div>
    <BreadcrumbRoot>
      <Breadcrumb>
        <a className="gap-2">
          <Home className="h-4 w-auto" /> My-namespace
        </a>
      </Breadcrumb>
      <Breadcrumb>
        <a className="gap-2">
          <Home /> My-namespace
        </a>
      </Breadcrumb>
      <Breadcrumb className="gap-2">
        <Home /> My-namespace
      </Breadcrumb>
    </BreadcrumbRoot>
  </div>
);

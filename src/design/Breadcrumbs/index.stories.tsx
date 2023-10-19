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

export const WithIcons = () => (
  <BreadcrumbRoot>
    <Breadcrumb noArrow>
      <Home /> My-namespace
    </Breadcrumb>
    <Breadcrumb>
      <Home /> My-namespace
    </Breadcrumb>
    <Breadcrumb>
      <Home /> My-namespace
    </Breadcrumb>
  </BreadcrumbRoot>
);
export const WithIconsAndLinks = () => (
  <BreadcrumbRoot>
    <Breadcrumb noArrow>
      <a href="#">
        <Home /> My-namespace
      </a>
    </Breadcrumb>
    <Breadcrumb>
      <a href="#">
        <Home /> My-namespace
      </a>
    </Breadcrumb>
    <Breadcrumb>
      <a href="#">
        <Home /> My-namespace
      </a>
    </Breadcrumb>
  </BreadcrumbRoot>
);

export const ResponsiveBreadcrumbs = () => (
  <BreadcrumbRoot>
    <Breadcrumb noArrow>
      <a href="#">
        <Home /> My-namespace
      </a>
    </Breadcrumb>
    <Breadcrumb>
      <a href="#">
        <Home /> My-namespace
      </a>
    </Breadcrumb>
    <Breadcrumb>
      <a href="#">
        <Home /> My-namespace
      </a>
    </Breadcrumb>
    <Breadcrumb>
      <a href="#">
        <Home /> My-namespace
      </a>
    </Breadcrumb>
    <Breadcrumb>
      <a href="#">
        <Home /> My-namespace
      </a>
    </Breadcrumb>
    <Breadcrumb>
      <a href="#">
        <Home /> My-namespace
      </a>
    </Breadcrumb>
    <Breadcrumb>
      <a href="#">
        <Home /> My-namespace
      </a>
    </Breadcrumb>
    <Breadcrumb>
      <a href="#">
        <Home /> My-namespace
      </a>
    </Breadcrumb>
    <Breadcrumb>
      <a href="#">
        <Home /> My-namespace
      </a>
    </Breadcrumb>
    <Breadcrumb>
      <a href="#">
        <Home /> My-namespace
      </a>
    </Breadcrumb>
  </BreadcrumbRoot>
);

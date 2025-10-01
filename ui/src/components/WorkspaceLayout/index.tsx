import { FC } from "react";
import { LayoutsType } from "~/util/store/editor";

type WorkspaceLayoutProps = {
  layout: LayoutsType;
  editorComponent: JSX.Element;
};

export const WorkspaceLayout: FC<WorkspaceLayoutProps> = ({
  layout,
  editorComponent,
}) => {
  switch (layout) {
    case "code":
      return editorComponent;
  }
};

import { FC } from "react";
import { LayoutsType } from "~/util/store/editor";

type WorkspaceLayoutProps = {
  layout: LayoutsType;
  editorComponent: JSX.Element;
  diagramComponent: JSX.Element;
};

export const WorkspaceLayout: FC<WorkspaceLayoutProps> = ({
  layout,
  diagramComponent,
  editorComponent,
}) => {
  switch (layout) {
    case "code":
      return editorComponent;
    case "diagram":
      return diagramComponent;
    case "splitVertically":
      return (
        <div className="grid grow grid-cols-2 gap-x-5">
          {editorComponent}
          {diagramComponent}
        </div>
      );
    case "splitHorizontally":
      return (
        <div className="grid grow grid-rows-2 gap-y-5">
          {editorComponent}
          {diagramComponent}
        </div>
      );
  }
};

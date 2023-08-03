import { Card } from "~/design/Card";
import { FC } from "react";
import { InstanceLayout } from "~/util/store/logs";
import { twMergeClsx } from "~/util/helpers";

type WorkspaceLayoutProps = {
  layout: InstanceLayout;
  logComponent: JSX.Element;
  diagramComponent: JSX.Element;
  inputOutputComponent: JSX.Element;
};

const WorkspaceLayout: FC<WorkspaceLayoutProps> = ({
  layout,
  logComponent,
  diagramComponent,
  inputOutputComponent,
}) => {
  switch (layout) {
    case "logs":
      return (
        <div
          className={twMergeClsx(
            "grid grow gap-5 p-5",
            "grid-rows-[calc(100vh-20rem)]",
            "sm:grid-rows-[calc(100vh-17rem)]",
            "lg:grid-rows-[calc(100vh-12rem)]"
          )}
        >
          <Card className="relative grid p-5">{logComponent}</Card>
        </div>
      );
    case "diagram":
      return (
        <div className="grid p-5">
          <Card className="flex">{diagramComponent}</Card>
        </div>
      );
    case "input-output":
      return (
        <div className="grid p-5">
          <Card className="flex p-5">{inputOutputComponent}</Card>
        </div>
      );
    case "none":
      return (
        <div
          className={twMergeClsx(
            "grid grow gap-5 p-5",
            "grid-rows-[minmax(300px,45vh)_1fr]",
            "grid-cols-[1fr_500px]"
          )}
        >
          <Card className="relative col-span-2 grid p-5">{logComponent}</Card>
          <Card className="flex">{diagramComponent}</Card>
          <Card className="flex p-5">{inputOutputComponent}</Card>
        </div>
      );
  }
};

export default WorkspaceLayout;

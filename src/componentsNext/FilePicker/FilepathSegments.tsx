import { Breadcrumb, BreadcrumbRoot } from "~/design/Breadcrumbs";

import { FC } from "react";
import { Home } from "lucide-react";

export type FilePathSegmentsType = {
  segments: {
    relative: string;
    absolute: string;
  }[];
  setPath: (path: string) => void;
};

export const FilePathSegments: FC<FilePathSegmentsType> = ({
  segments,
  setPath,
}) => (
  <BreadcrumbRoot className="py-3">
    <Breadcrumb
      noArrow
      onClick={() => {
        setPath("/");
      }}
      className="h-5 hover:underline"
    >
      <Home />
    </Breadcrumb>
    {segments.map((file) => {
      const isEmpty = file.absolute === "";

      if (isEmpty) return null;
      return (
        <Breadcrumb
          key={file.relative}
          onClick={() => {
            setPath(file.absolute);
          }}
          className="h-5 hover:underline"
        >
          {file.relative}
        </Breadcrumb>
      );
    })}
  </BreadcrumbRoot>
);

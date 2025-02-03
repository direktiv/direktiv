import { FC, PropsWithChildren } from "react";
import { getFilenameFromPath, isPreviewable } from "~/api/files/utils";

import { BaseFileSchemaType } from "~/api/files/schema";
import { DialogTrigger } from "~/design/Dialog";
import { Link } from "@tanstack/react-router";

type ConditionalLinkProps = PropsWithChildren & {
  file: BaseFileSchemaType;
  namespace: string;
  onPreviewClicked: (file: BaseFileSchemaType) => void;
};

export const ConditionalLink: FC<ConditionalLinkProps> = ({
  file,
  namespace,
  onPreviewClicked,
  children,
}) => {
  const linkToPreview = isPreviewable(file.type);
  if (linkToPreview)
    return (
      <DialogTrigger
        className="flex-1 hover:underline"
        role="button"
        onClick={() => {
          onPreviewClicked(file);
        }}
        asChild
      >
        <a>{children}</a>
      </DialogTrigger>
    );

  return (
    <Link
      data-testid={`explorer-item-link-${getFilenameFromPath(file.path)}`}
      to="/n/$namespace/explorer/tree/$"
      params={{ namespace, _splat: file.path }}
      className="flex-1 hover:underline"
    >
      {children}
    </Link>
  );
};

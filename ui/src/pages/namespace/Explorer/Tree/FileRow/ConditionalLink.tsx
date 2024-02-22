import { FC, PropsWithChildren } from "react";
import {
  fileTypeToExplorerSubpage,
  getFilenameFromPath,
  isPreviewable,
} from "~/api/files/utils";

import { BaseFileSchemaType } from "~/api/files/schema";
import { DialogTrigger } from "~/design/Dialog";
import { Link } from "react-router-dom";
import { pages } from "~/util/router/pages";

type ConditionalLinkProps = PropsWithChildren & {
  node: BaseFileSchemaType;
  namespace: string;
  onPreviewClicked: (file: BaseFileSchemaType) => void;
};

export const ConditionalLink: FC<ConditionalLinkProps> = ({
  node,
  namespace,
  onPreviewClicked,
  children,
}) => {
  const linkToPreview = isPreviewable(node.type);
  if (linkToPreview)
    return (
      <DialogTrigger
        className="flex-1 hover:underline"
        role="button"
        onClick={() => {
          onPreviewClicked(node);
        }}
        asChild
      >
        <a>{children}</a>
      </DialogTrigger>
    );

  const linkTarget = pages.explorer.createHref({
    namespace,
    path: node.path,
    subpage: fileTypeToExplorerSubpage(node.type),
  });

  return (
    <Link
      data-testid={`explorer-item-link-${getFilenameFromPath(node.path)}`}
      to={linkTarget}
      className="flex-1 hover:underline"
    >
      {children}
    </Link>
  );
};

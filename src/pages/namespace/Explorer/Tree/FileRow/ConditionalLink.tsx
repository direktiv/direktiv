import { FC, PropsWithChildren } from "react";

import { Link } from "react-router-dom";
import { NodeSchemaType } from "~/api/tree/schema";
import { pages } from "~/util/router/pages";

export const ConditionalLink: FC<
  PropsWithChildren & { file: NodeSchemaType; namespace: string }
> = ({ file, namespace, children }) => {
  const isFile = file.expandedType === "file";
  if (isFile) return <a className="flex-1 hover:underline">{children}</a>;
  const linkTarget = pages.explorer.createHref({
    namespace,
    path: file.path,
    subpage: file.expandedType === "workflow" ? "workflow" : undefined,
  });

  return (
    <Link
      data-testid={`explorer-item-link-${file.name}`}
      to={linkTarget}
      className="flex-1 hover:underline"
    >
      {children}
    </Link>
  );
};

import {
  LinkComponentProps,
  Link as TanStackLink,
} from "@tanstack/react-router";
import {
  activeClassNames,
  baseClassNames,
  inactiveClassNames,
} from "~/design/NavigationLink";

import { twMergeClsx } from "~/util/helpers";

export const RouterNavigationLink = ({ ...props }: LinkComponentProps) => (
  <TanStackLink
    {...props}
    className={twMergeClsx(baseClassNames, props.className)}
    activeProps={{
      className: activeClassNames,
    }}
    inactiveProps={{
      className: inactiveClassNames,
    }}
  />
);

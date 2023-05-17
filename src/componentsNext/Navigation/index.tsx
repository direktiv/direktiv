import { FC } from "react";
import { NavLink } from "react-router-dom";
import { createClassNames } from "~/design/NavigationLink";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";

const Navigation: FC = () => {
  const namespace = useNamespace();
  if (!namespace) return null;
  return (
    <>
      {Object.entries(pages)
        .filter(([, item]) => !!item.icon || !!item.name)
        .map(([key, item]) => {
          if (!item.icon || !item.name) return null;
          return (
            <NavLink
              key={key}
              to={item.createHref({ namespace })}
              className={({ isActive }) => createClassNames(isActive)}
              end={false}
            >
              <item.icon aria-hidden="true" /> {item.name}
            </NavLink>
          );
        })}
    </>
  );
};

export default Navigation;
